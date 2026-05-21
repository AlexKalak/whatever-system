package chaintools

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func ParseUniswapV3TickFromTopic(topic common.Hash) int32 {
	// Convert bytes to big.Int
	b := new(big.Int).SetBytes(topic.Bytes())

	// Mask only the lowest 24 bits
	value := b.Int64() & 0xFFFFFF // 24 bits

	// If sign bit (bit 23) is set, convert to negative
	if value&0x800000 != 0 {
		value = value - 0x1000000
	}

	return int32(value)
}
