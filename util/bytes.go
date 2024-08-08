package util

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func MustDecodeHexToBytes(hex string) []byte {
	data, _ := hexutil.Decode(hex)
	return data
}

func MustDecodeHexArrayToBytes(hexArray []string) [][]byte {
	data := make([][]byte, 0, len(hexArray))
	for _, v := range hexArray {
		data = append(data, MustDecodeHexToBytes(v))
	}
	return data
}
