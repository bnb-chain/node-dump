package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExportedAccount is an exported account.
type ExportedAccount struct {
	Address       sdk.AccAddress `json:"address"`
	AccountNumber int64          `json:"account_number"`
	Coins         sdk.Coins      `json:"coins,omitempty"`
}

// ExportedProof is an exported proof.
type ExportedProof struct {
	Address sdk.AccAddress `json:"address"`
	Coin    sdk.Coin       `json:"coin"`
	Proof   []string       `json:"proof"`
}

// ExportedAccountState is an exported account state.
type ExportedAccountState struct {
	ChainID     string             `json:"chain_id"`
	BlockHeight int64              `json:"block_height"`
	CommitID    sdk.CommitID       `json:"commit_id"`
	Accounts    []*ExportedAccount `json:"-"`
	StateRoot   string             `json:"state_root"`
	Proofs      []*ExportedProof   `json:"-"`
}
