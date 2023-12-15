package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/node/app"
	nodetypes "github.com/bnb-chain/node/common/types"

	mt "github.com/txaty/go-merkletree"

	"github.com/bnb-chain/node-dump/types"
)

const (
	flagTraceStore = "trace-store"
)

func NewHashFunc(data []byte) ([]byte, error) {
	return crypto.Keccak256(data), nil
}

type leafNode struct {
	Address sdk.AccAddress `json:"address"`
	Coin    sdk.Coin       `json:"coin"`
}

// Serialize implements merkle tree data Serialize method.
func (node *leafNode) Serialize() ([]byte, error) {
	var symbol [32]byte
	copy(symbol[:], node.Coin.Denom)
	return crypto.Keccak256(
		node.Address.Bytes(),
		symbol[:],
		big.NewInt(node.Coin.Amount).FillBytes(make([]byte, 32)),
	), nil
}

func (node *leafNode) Print() string {
	buf, _ := node.Serialize()

	return "0x" + common.Bytes2Hex(crypto.Keccak256(buf))
}

// ExportAccountsBalanceWithProof exports blockchain world state to json.
func ExportAccountsBalanceWithProof(app *app.BNBBeaconChain, outputPath string) (err error) {
	ctx := app.NewContext(sdk.RunTxModeCheck, abci.Header{})

	allowedTokens := make(map[string]struct{})
	tokens := app.TokenMapper.GetTokenList(ctx, false, true)
	for _, token := range tokens {
		if token.GetContractAddress() != "" {
			allowedTokens[token.GetSymbol()] = struct{}{}
		}
	}
	allowedTokens["BNB"] = struct{}{}

	// iterate to get the accounts
	accounts := []*types.ExportedAccount{}
	mtData := []mt.DataBlock{}

	appendAccount := func(acc sdk.Account) (stop bool) {
		namedAcc := acc.(nodetypes.NamedAccount)
		addr := namedAcc.GetAddress()
		coins := namedAcc.GetCoins()
		frozenCoins := namedAcc.GetFrozenCoins()
		lockedCoins := namedAcc.GetLockedCoins()

		allCoins := coins.Plus(frozenCoins)
		allCoins = allCoins.Plus(lockedCoins)

		account := types.ExportedAccount{
			Address:       addr,
			AccountNumber: namedAcc.GetAccountNumber(),
			Coins:         allCoins.Sort(),
		}
		accounts = append(accounts, &account)

		for index := range allCoins {
			if _, allowed := allowedTokens[allCoins[index].Denom]; !allowed {
				// skip if not cross-chain token
				continue
			}

			if allCoins[index].Amount > 0 {
				mtData = append(mtData, &leafNode{
					Address: addr,
					Coin:    allCoins[index],
				})
			}
		}

		trace("address", acc.GetAddress(), "account:", account)

		return false
	}

	trace("iterate accounts...")
	app.AccountKeeper.IterateAccounts(ctx, appendAccount)

	trace("make merkle tree...")
	// create a Merkle Tree config and set parallel run parameters
	config := &mt.Config{
		HashFunc:           NewHashFunc,
		RunInParallel:      true,
		SortSiblingPairs:   true,
		DisableLeafHashing: true,
	}

	tree, err := mt.New(config, mtData)
	if err != nil {
		return err
	}

	trace("make proofs...")
	proofs := tree.Proofs
	maxProofLength := 0
	exportedProof := make([]*types.ExportedProof, 0, len(proofs))
	trace("proofs length", len(proofs))
	for i := 0; i < len(mtData); i++ {
		proof := proofs[i]
		nProof := make([]string, 0, len(proof.Siblings))
		for j := 0; j < len(proof.Siblings); j++ {
			nProof = append(nProof, "0x"+common.Bytes2Hex(proof.Siblings[j]))
		}

		leaf := mtData[i].(*leafNode)
		exportedProof = append(exportedProof, &types.ExportedProof{
			Address: leaf.Address,
			Coin:    leaf.Coin,
			Proof:   nProof,
		})
		if proofLength := len(proof.Siblings); proofLength > maxProofLength {
			maxProofLength = proofLength
		}
		trace("address:", accounts[i].Address.String(), "proof:", nProof, "leaf:", leaf.Print())
	}
	trace("max proof length:", maxProofLength)

	genState := types.ExportedAccountState{
		ChainID:     app.CheckState.Ctx.ChainID(),
		BlockHeight: app.LastBlockHeight(),
		CommitID:    app.LastCommitID(),
		Accounts:    accounts,
		StateRoot:   "0x" + common.Bytes2Hex(tree.Root),
		Proofs:      exportedProof,
	}

	trace("write to file...")

	// write the state to the file
	baseFile, err := os.OpenFile(path.Join(outputPath, "base.json"), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer baseFile.Close()
	err = writeJSONFile(baseFile, genState)
	if err != nil {
		return err
	}

	// write the accounts to the file
	accountFile, err := os.OpenFile(path.Join(outputPath, "accounts.json"), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer accountFile.Close()
	err = writeJSONFileInStream(accountFile, func(encoder *json.Encoder) error {
		for i, account := range genState.Accounts {
			err = encoder.Encode(account)
			if err != nil {
				return err
			}
			if i < len(accounts)-1 {
				_, err = accountFile.WriteString(`,`)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// write the proofs to the file
	proofFile, err := os.OpenFile(path.Join(outputPath, "proofs.json"), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer proofFile.Close()
	err = writeJSONFileInStream(proofFile, func(encoder *json.Encoder) error {
		for i, proof := range genState.Proofs {
			err = encoder.Encode(proof)
			if err != nil {
				return err
			}
			if i < len(proofs)-1 {
				_, err = proofFile.WriteString(`,`)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func writeJSONFile(file *os.File, data interface{}) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	return encoder.Encode(data)
}
func writeJSONFileInStream(file *os.File, marshal func(*json.Encoder) error) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	if _, err := file.WriteString(`[`); err != nil {
		return err
	}
	if err := marshal(encoder); err != nil {
		return err
	}
	if _, err := file.WriteString(`]`); err != nil {
		return err
	}
	return nil
}

// ExportCmd dumps app state to JSON.
func ExportCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "export <path>",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("<path/state.json> should be set")
			}
			if args[0] == "" {
				return fmt.Errorf("<path/state.json> should be set")
			}
			home := viper.GetString("home")
			traceWriterFile := viper.GetString(flagTraceStore)
			emptyState, err := isEmptyState(home)
			if err != nil {
				return err
			}

			if emptyState {
				fmt.Println("WARNING: State is not initialized. Returning genesis file.")
				genesisFile := path.Join(home, "config", "genesis.json")
				genesis, err := os.ReadFile(genesisFile)
				if err != nil {
					return err
				}
				fmt.Println(string(genesis))
				return nil
			}

			db, err := openDB(home)
			if err != nil {
				return err
			}
			traceWriter, err := openTraceWriter(traceWriterFile)
			if err != nil {
				return err
			}

			dapp := app.NewBNBBeaconChain(ctx.Logger, db, traceWriter)
			err = ExportAccountsBalanceWithProof(dapp, args[0])
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func isEmptyState(home string) (bool, error) {
	files, err := os.ReadDir(path.Join(home, "data"))
	if err != nil {
		return false, err
	}

	// only priv_validator_state.json is created
	return len(files) == 1 && files[0].Name() == "priv_validator_state.json", nil
}

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	db, err := dbm.NewGoLevelDB("application", dataDir)
	return db, err
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile != "" {
		w, err = os.OpenFile(
			traceWriterFile,
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0600,
		)
		return
	}
	return
}

func trace(a ...any) {
	if traceLog {
		a = append([]any{"time:", time.Now()}, a...)
		fmt.Println(a...)
	}
}

var (
	// TraceLog is a flag to print out full stack trace on errors
	traceLog = false
)

func main() {
	cdc := app.Codec
	ctx := app.ServerContext

	rootCmd := &cobra.Command{
		Use:               "dump",
		Short:             "BNBChain dump tool",
		PersistentPreRunE: app.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(ExportCmd(ctx.ToCosmosServerCtx(), cdc))
	rootCmd.PersistentFlags().BoolVar(&traceLog, "tracelog", false, "print out full stack trace on errors")
	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "BC", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		fmt.Println(err)
		return
	}
}
