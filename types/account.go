package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// ExportedAccount is an exported account.
type ExportedAccount struct {
	Address       sdk.AccAddress `json:"address"`
	AccountNumber int64          `json:"account_number"`
	Coins         sdk.Coins      `json:"coins,omitempty"`
	FrozenCoins   sdk.Coins      `json:"frozen_coins,omitempty"`
	LockedCoins   sdk.Coins      `json:"locked_coins,omitempty"`
}

// Serialize implements merkle tree data Serialize method.
func (acc *ExportedAccount) Serialize() ([]byte, error) {
	return crypto.Keccak256Hash(
		acc.Address.Bytes(),
		big.NewInt(acc.AccountNumber).Bytes(),
		[]byte(acc.Coins.String()),
		[]byte(acc.FrozenCoins.String()),
		[]byte(acc.LockedCoins.String()),
	).Bytes(), nil
}

// ExportedAssets is a map of asset name to amount
type ExportedAssets map[string]int64

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
