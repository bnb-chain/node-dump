package util

import (
	"bytes"

	"github.com/ethereum/go-ethereum/crypto"
)

func VerifyMerkleProof(rootHash []byte, proof [][]byte, leaf []byte) bool {
	hash := leaf
	for _, proofElement := range proof {
		hash = hashPair(hash, proofElement)
	}
	return bytes.Equal(hash, rootHash)

}

func hashPair(left, right []byte) []byte {
	if bytes.Compare(left, right) < 0 {
		return crypto.Keccak256(left, right)
	}
	return crypto.Keccak256(right, left)
}
