package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExportedAccount is an exported account.
type ExportedAccount struct {
	Address       sdk.AccAddress `json:"address"`
	AccountNumber int64          `json:"account_number"`
	Coins         sdk.Coins      `json:"coins,omitempty"`
	FreeCoins     sdk.Coins      `json:"free_coins,omitempty"`
	FrozenCoins   sdk.Coins      `json:"frozen_coins,omitempty"`
	LockedCoins   sdk.Coins      `json:"locked_coins,omitempty"`
}

// ExportedAssets is a map of asset name to amount
type ExportedAssets map[string]*ExportedAsset

type ExportedAsset struct {
	Owner  sdk.AccAddress `json:"owner,omitempty"`
	Amount int64          `json:"amount"`
}

// ExportedProof is an exported proof.
type ExportedProof struct {
	Address sdk.AccAddress `json:"address"`
	Index   int64          `json:"index"`
	Coin    sdk.Coin       `json:"coin"`
	Proof   []string       `json:"proof"`
}

// ExportedAccountState is an exported account state.
type ExportedAccountState struct {
	ChainID     string             `json:"chain_id"`
	BlockHeight int64              `json:"block_height"`
	CommitID    sdk.CommitID       `json:"commit_id"`
	Accounts    []*ExportedAccount `json:"-"`
	Assets      ExportedAssets     `json:"assets"`
	StateRoot   string             `json:"state_root"`
	Proofs      []*ExportedProof   `json:"-"`
}
