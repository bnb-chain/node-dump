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
	tmCrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/node/app"
	nodetypes "github.com/bnb-chain/node/common/types"

	mt "github.com/txaty/go-merkletree"

	"github.com/bnb-chain/node-dump/types"
	"github.com/bnb-chain/node-dump/util"
)

const (
	displayProcessInterval = time.Second
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

	// Escrow Accounts
	escrowAccs := make(map[string]struct{})
	// bnb prefix address: bnb1vu5max8wqn997ayhrrys0drpll2rlz4dh39s3h
	// tbnb prefix address: tbnb1vu5max8wqn997ayhrrys0drpll2rlz4deyv53x
	depositedCoinsAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainDepositedCoins")))
	// bnb prefix address: bnb1j725qk29cv4kwpers4addy9x93ukhw7czfkjaj
	// tbnb prefix address: tbnb1j725qk29cv4kwpers4addy9x93ukhw7cvulkar
	delegationAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainStakeDelegation")))
	// bnb prefix address: bnb1v8vkkymvhe2sf7gd2092ujc6hweta38xadu2pj
	// tbnb prefix address: tbnb1v8vkkymvhe2sf7gd2092ujc6hweta38xnc4wpr
	pegAccount := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainPegAccount")))
	// bnb prefix address: bnb1wxeplyw7x8aahy93w96yhwm7xcq3ke4f8ge93u
	// tbnb prefix address: tbnb1wxeplyw7x8aahy93w96yhwm7xcq3ke4ffasp3d
	atomicSwapCoinsAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainAtomicSwapCoins")))
	// bnb prefix address: bnb1hn8ym9xht925jkncjpf7lhjnax6z8nv24fv2yq
	// tbnb prefix address: tbnb1hn8ym9xht925jkncjpf7lhjnax6z8nv2mu9wy3
	timeLockCoinsAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainTimeLockCoins")))
	// nil address
	emptyAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte(nil)))
	// 0x0000... address
	zeroAccAddr, err := sdk.AccAddressFromHex("0000000000000000000000000000000000000000")
	if err != nil {
		return err
	}
	trace("escrow accounts",
		"depositedCoinsAccAddr:", depositedCoinsAccAddr.String(),
		"delegationAccAddr:", delegationAccAddr.String(),
		"pegAccount:", pegAccount.String(),
		"atomicSwapCoinsAccAddr:", atomicSwapCoinsAccAddr.String(),
		"timeLockCoinsAccAddr:", timeLockCoinsAccAddr.String(),
		"emptyAccAddr:", emptyAccAddr.String(),
		"zeroAccAddr:", zeroAccAddr.String(),
	)
	escrowAccs[depositedCoinsAccAddr.String()] = struct{}{}
	escrowAccs[delegationAccAddr.String()] = struct{}{}
	escrowAccs[pegAccount.String()] = struct{}{}
	escrowAccs[atomicSwapCoinsAccAddr.String()] = struct{}{}
	escrowAccs[timeLockCoinsAccAddr.String()] = struct{}{}
	escrowAccs[emptyAccAddr.String()] = struct{}{}
	escrowAccs[zeroAccAddr.String()] = struct{}{}

	// iterate to get the accounts
	accounts := []*types.ExportedAccount{}
	mtData := []mt.DataBlock{}

	appendAccount := func(acc sdk.Account) (stop bool) {
		namedAcc := acc.(nodetypes.NamedAccount)
		addr := namedAcc.GetAddress()
		if _, exist := escrowAccs[addr.String()]; exist {
			trace("skip escrow account:", addr.String())
			return false
		}

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
		trace("address:", leaf.Address.String(), "proof:", nProof, "leaf:", leaf.Print())
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

func VerifyProofsFromDatabase(app *app.BNBBeaconChain, proofPath string) (err error) {
	// load exported state
	stateFile, err := os.Open(path.Join(proofPath, "base.json"))
	if err != nil {
		return err
	}
	defer stateFile.Close()
	var state types.ExportedAccountState
	err = json.NewDecoder(stateFile).Decode(&state)
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	defer close(errChan)

	// load exported proofs
	stream := util.NewJSONStream(func() any {
		return &types.ExportedProof{}
	})

	proofs := make(map[string]*types.ExportedProof)
	go func() {
		for data := range stream.Watch() {
			if data.Error != nil {
				errChan <- data.Error
				return
			}
			proof := data.Data.(*types.ExportedProof)
			index := proof.Address.String() + ":" + proof.Coin.Denom
			proofs[index] = proof
		}
		errChan <- nil
	}()
	stream.Start(path.Join(proofPath, "proofs.json"))
	err = <-errChan
	if err != nil {
		return err
	}

	// prepare context
	ctx := app.NewContext(sdk.RunTxModeCheck, abci.Header{})
	// Escrow Accounts
	escrowAccs := make(map[string]struct{})
	// bnb prefix address: bnb1vu5max8wqn997ayhrrys0drpll2rlz4dh39s3h
	// tbnb prefix address: tbnb1vu5max8wqn997ayhrrys0drpll2rlz4deyv53x
	depositedCoinsAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainDepositedCoins")))
	// bnb prefix address: bnb1j725qk29cv4kwpers4addy9x93ukhw7czfkjaj
	// tbnb prefix address: tbnb1j725qk29cv4kwpers4addy9x93ukhw7cvulkar
	delegationAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainStakeDelegation")))
	// bnb prefix address: bnb1v8vkkymvhe2sf7gd2092ujc6hweta38xadu2pj
	// tbnb prefix address: tbnb1v8vkkymvhe2sf7gd2092ujc6hweta38xnc4wpr
	pegAccount := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainPegAccount")))
	// bnb prefix address: bnb1wxeplyw7x8aahy93w96yhwm7xcq3ke4f8ge93u
	// tbnb prefix address: tbnb1wxeplyw7x8aahy93w96yhwm7xcq3ke4ffasp3d
	atomicSwapCoinsAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainAtomicSwapCoins")))
	// bnb prefix address: bnb1hn8ym9xht925jkncjpf7lhjnax6z8nv24fv2yq
	// tbnb prefix address: tbnb1hn8ym9xht925jkncjpf7lhjnax6z8nv2mu9wy3
	timeLockCoinsAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte("BinanceChainTimeLockCoins")))
	// nil address
	emptyAccAddr := sdk.AccAddress(tmCrypto.AddressHash([]byte(nil)))
	// 0x0000... address
	zeroAccAddr, err := sdk.AccAddressFromHex("0000000000000000000000000000000000000000")
	if err != nil {
		return err
	}
	trace("escrow accounts",
		"depositedCoinsAccAddr:", depositedCoinsAccAddr.String(),
		"delegationAccAddr:", delegationAccAddr.String(),
		"pegAccount:", pegAccount.String(),
		"atomicSwapCoinsAccAddr:", atomicSwapCoinsAccAddr.String(),
		"timeLockCoinsAccAddr:", timeLockCoinsAccAddr.String(),
		"emptyAccAddr:", emptyAccAddr.String(),
		"zeroAccAddr:", zeroAccAddr.String(),
	)
	escrowAccs[depositedCoinsAccAddr.String()] = struct{}{}
	escrowAccs[delegationAccAddr.String()] = struct{}{}
	escrowAccs[pegAccount.String()] = struct{}{}
	escrowAccs[atomicSwapCoinsAccAddr.String()] = struct{}{}
	escrowAccs[timeLockCoinsAccAddr.String()] = struct{}{}
	escrowAccs[emptyAccAddr.String()] = struct{}{}
	escrowAccs[zeroAccAddr.String()] = struct{}{}

	// iterate to verify the accounts
	count := 0
	merkleRoot := util.MustDecodeHexToBytes(state.StateRoot)
	ticker := time.NewTicker(displayProcessInterval)
	defer ticker.Stop()
	app.AccountKeeper.IterateAccounts(ctx, func(acc sdk.Account) (stop bool) {
		select {
		case <-ticker.C:
			trace("process", fmt.Sprintf("%d", count*100/len(proofs))+"%",
				"total", len(proofs),
				"count", count)
		default:
		}

		namedAcc := acc.(nodetypes.NamedAccount)
		addr := namedAcc.GetAddress()
		if _, matched := escrowAccs[addr.String()]; matched {
			trace("skip escrow account:", addr.String())
			return false
		}

		coins := namedAcc.GetCoins()
		frozenCoins := namedAcc.GetFrozenCoins()
		lockedCoins := namedAcc.GetLockedCoins()

		allCoins := coins.Plus(frozenCoins)
		allCoins = allCoins.Plus(lockedCoins)

		for _, coin := range allCoins {
			if coin.Amount > 0 {
				proof, exist := proofs[addr.String()+":"+coin.Denom]
				if !exist {
					trace("proof not found", addr.String(), coin.Denom)
					return true
				}

				if coin.Amount != proof.Coin.Amount {
					trace("amount mismatch",
						"address", addr.String(),
						"symbol", coin.Denom,
						"expected", coin.Amount,
						"actual", proof.Coin.Amount)
					return true
				}

				// verify merkle proof
				leaf := &leafNode{
					Address: addr,
					Coin:    coin,
				}
				leafHash, err := leaf.Serialize()
				if err != nil {
					trace("merkle proof serialization failed",
						"address", addr.String(),
						"symbol", coin.Denom,
						"amount", coin.Amount)
					return true
				}

				if !util.VerifyMerkleProof(merkleRoot, util.MustDecodeHexArrayToBytes(proof.Proof), leafHash) {
					trace("merkle proof verification failed",
						"address", addr.String(),
						"symbol", coin.Denom,
						"amount", coin.Amount)
					return true
				}

				count++
			}
		}

		return false
	})

	if count != len(proofs) {
		return fmt.Errorf("account mismatch: %d != %d", count, len(proofs))
	}
	return nil
}

// VerificationCmd verify the proofs from database.
func VerificationCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "verify <path>",
		Short: "Verify the exported proofs from database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("<proof path> should be set")
			}
			if args[0] == "" {
				return fmt.Errorf("<proof path> should be set")
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
			err = VerifyProofsFromDatabase(dapp, args[0])
			if err != nil {
				return err
			}
			fmt.Println("Verification passed")

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
	rootCmd.AddCommand(VerificationCmd(ctx.ToCosmosServerCtx(), cdc))
	rootCmd.PersistentFlags().BoolVar(&traceLog, "tracelog", false, "print out full stack trace on errors")
	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "BC", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		fmt.Println(err)
		return
	}
}
