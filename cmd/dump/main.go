package main

import (
	"encoding/json"
	"fmt"
	"io"
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

// ExportAccounts exports blockchain world state to json.
func ExportAccounts(app *app.BNBBeaconChain, outputPath string) (err error) {
	ctx := app.NewContext(sdk.RunTxModeCheck, abci.Header{})

	// iterate to get the accounts
	accounts := []*types.ExportedAccount{}
	mtData := []mt.DataBlock{}
	assets := types.ExportedAssets{}

	appendAccount := func(acc sdk.Account) (stop bool) {
		namedAcc := acc.(nodetypes.NamedAccount)
		addr := namedAcc.GetAddress()
		coins := namedAcc.GetCoins()
		frozenCoins := namedAcc.GetFrozenCoins()
		lockedCoins := namedAcc.GetLockedCoins()

		summaryCoinsMap := map[string]sdk.Coin{}

		for _, coin := range coins {
			asset, exist := assets[coin.Denom]
			if exist {
				asset.Amount += coin.Amount
			} else {
				token, err := app.TokenMapper.GetToken(ctx, coin.Denom)
				if err != nil {
					trace(err)
					return true
				}
				assets[coin.Denom] = &types.ExportedAsset{
					Owner:  token.GetOwner(),
					Amount: coin.Amount,
				}
			}

			sCoin := summaryCoinsMap[coin.Denom]
			sCoin.Denom = coin.Denom
			sCoin.Amount += coin.Amount
			summaryCoinsMap[coin.Denom] = sCoin
		}
		for _, coin := range frozenCoins {
			asset, exist := assets[coin.Denom]
			if exist {
				asset.Amount += coin.Amount
			} else {
				token, err := app.TokenMapper.GetToken(ctx, coin.Denom)
				if err != nil {
					trace(err)
					return true
				}
				assets[coin.Denom] = &types.ExportedAsset{
					Owner:  token.GetOwner(),
					Amount: coin.Amount,
				}
			}
			sCoin := summaryCoinsMap[coin.Denom]
			sCoin.Denom = coin.Denom
			sCoin.Amount += coin.Amount
			summaryCoinsMap[coin.Denom] = sCoin
		}
		for _, coin := range lockedCoins {
			asset, exist := assets[coin.Denom]
			if exist {
				asset.Amount += coin.Amount
			} else {
				token, err := app.TokenMapper.GetToken(ctx, coin.Denom)
				if err != nil {
					trace(err)
					return true
				}
				assets[coin.Denom] = &types.ExportedAsset{
					Owner:  token.GetOwner(),
					Amount: coin.Amount,
				}
			}
			sCoin := summaryCoinsMap[coin.Denom]
			sCoin.Denom = coin.Denom
			sCoin.Amount += coin.Amount
			summaryCoinsMap[coin.Denom] = sCoin
		}

		summaryCoins := make(sdk.Coins, 0, len(summaryCoinsMap))
		for _, coin := range summaryCoinsMap {
			summaryCoins = append(summaryCoins, coin)
		}

		account := types.ExportedAccount{
			Address:       addr,
			AccountNumber: namedAcc.GetAccountNumber(),
			SummaryCoins:  summaryCoins.Sort(),
			Coins:         coins.Sort(),
			FrozenCoins:   frozenCoins.Sort(),
			LockedCoins:   lockedCoins.Sort(),
		}
		accounts = append(accounts, &account)
		mtData = append(mtData, &account)
		trace("address", acc.GetAddress(), "account:", account)
		if err != nil {
			fmt.Println(err)
			return true
		}

		return false
	}

	trace("iterate accounts...")
	app.AccountKeeper.IterateAccounts(ctx, appendAccount)

	trace("make merkle tree...")
	// create a Merkle Tree config and set parallel run parameters
	config := &mt.Config{
		HashFunc:         NewHashFunc,
		RunInParallel:    true,
		SortSiblingPairs: true,
	}

	tree, err := mt.New(config, mtData)
	if err != nil {
		return err
	}

	trace("make proofs...")
	proofs := tree.Proofs
	exportedProof := make(types.ExportedProofs, len(proofs))
	for i := 0; i < len(mtData); i++ {
		proof := proofs[i]
		nProof := make([]string, 0, len(proof.Siblings))
		for i := 0; i < len(proof.Siblings); i++ {
			nProof = append(nProof, "0x"+common.Bytes2Hex(proof.Siblings[i]))
		}
		exportedProof[accounts[i].Address.String()] = nProof
		trace("address:", accounts[i].Address.String(), "proof:", nProof)
	}

	genState := types.ExportedAccountState{
		ChainID:     app.CheckState.Ctx.ChainID(),
		BlockHeight: app.LastBlockHeight(),
		CommitID:    app.LastCommitID(),
		Accounts:    accounts,
		Assets:      assets,
		StateRoot:   "0x" + common.Bytes2Hex(tree.Root),
		Proofs:      exportedProof,
	}

	trace("write to file...")

	// write the accounts to the file
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	file.WriteString(`{
	"chain_id": "` + genState.ChainID + `",
	"block_height": ` + fmt.Sprint(genState.BlockHeight) + `,
	"commit_id": `)
	err = encoder.Encode(genState.CommitID)
	if err != nil {
		return err
	}
	file.WriteString(`,
	"state_root": "` + genState.StateRoot + `",
	"assets": `)
	err = encoder.Encode(genState.Assets)
	if err != nil {
		return err
	}
	file.WriteString(`,
	"accounts": [
	`)
	for i, account := range genState.Accounts {
		err = encoder.Encode(account)
		if err != nil {
			return err
		}
		if i < len(genState.Accounts)-1 {
			file.WriteString(`,
	`)
		}
	}

	maxProofs := len(genState.Proofs) - 1
	i := 0
	file.WriteString(`
	],
	"proofs": {`)
	for addr, proof := range genState.Proofs {
		file.WriteString(`
		"` + addr + `":`)
		err = encoder.Encode(proof)
		if err != nil {
			return err
		}
		if i < maxProofs {
			file.WriteString(`,`)
		}
		i++
	}
	file.WriteString(`}
}`)
	return nil
}

// ExportCmd dumps app state to JSON.
func ExportCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "export <path/state.json>",
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
			err = ExportAccounts(dapp, args[0])
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
