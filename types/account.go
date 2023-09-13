package types

import (
	"bytes"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// ExportedAccount is an exported account.
type ExportedAccount struct {
	Address       sdk.AccAddress `json:"address"`
	AccountNumber int64          `json:"account_number"`
	SummaryCoins  sdk.Coins      `json:"summary_coins,omitempty"`
	Coins         sdk.Coins      `json:"coins,omitempty"`
	FrozenCoins   sdk.Coins      `json:"frozen_coins,omitempty"`
	LockedCoins   sdk.Coins      `json:"locked_coins,omitempty"`
}

// Serialize implements merkle tree data Serialize method.
func (acc *ExportedAccount) Serialize() ([]byte, error) {
	coinBytes := bytes.NewBuffer(nil)
	for _, coin := range acc.SummaryCoins {
		var b [32]byte
		copy(b[:], coin.Denom)
		coinBytes.Write(b[:])
		coinBytes.Write(big.NewInt(coin.Amount).Bytes())
	}
	return crypto.Keccak256Hash(
		acc.Address.Bytes(),
		big.NewInt(acc.AccountNumber).Bytes(),
		coinBytes.Bytes(),
	).Bytes(), nil
}

// ExportedAssets is a map of asset name to amount
type ExportedAssets map[string]*ExportedAsset

type ExportedAsset struct {
	Owner  sdk.AccAddress `json:"owner,omitempty"`
	Amount int64          `json:"amount"`
}

// ExportedProofs is a map of account address to merkle proof
type ExportedProofs map[string][]string

// ExportedAccountState is an exported account state.
type ExportedAccountState struct {
	ChainID     string             `json:"chain_id"`
	BlockHeight int64              `json:"block_height"`
	CommitID    sdk.CommitID       `json:"commit_id"`
	Accounts    []*ExportedAccount `json:"accounts"`
	Assets      ExportedAssets     `json:"assets"`
	StateRoot   string             `json:"state_root"`
	Proofs      ExportedProofs     `json:"proofs"`
}
