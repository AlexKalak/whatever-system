package chaintools

import (
	"errors"
	"math/big"
	"strings"

	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
)

var (
	q96  = new(big.Int).Lsh(big.NewInt(1), 96)
	q192 = new(big.Int).Mul(q96, q96)
)

// CountTradePriceBySqrtPriceX96AndDecimals calculates the Uniswap V3 spot trade price
// from sqrtPriceX96 and token decimals.
//
// The returned price is token1 per token0 in human token units:
//
//	price = (sqrtPriceX96 / 2^96)^2 * 10^token0Decimals / 10^token1Decimals
//
// The result is returned as an exact rational number to avoid losing precision.
func CountTradePriceBySqrtPriceX96AndDecimals(sqrtPriceX96 *big.Int, token0Decimals, token1Decimals uint8) (*big.Rat, error) {
	if sqrtPriceX96 == nil {
		return nil, errors.New("sqrtPriceX96 is nil")
	}
	if sqrtPriceX96.Sign() < 0 {
		return nil, errors.New("sqrtPriceX96 cannot be negative")
	}

	numerator := new(big.Int).Mul(sqrtPriceX96, sqrtPriceX96)
	numerator.Mul(numerator, pow10(token0Decimals))

	denominator := new(big.Int).Mul(new(big.Int).Set(q192), pow10(token1Decimals))

	return new(big.Rat).SetFrac(numerator, denominator), nil
}

// CountReverseTradePriceBySqrtPriceX96AndDecimals calculates token0 per token1
// from sqrtPriceX96 and token decimals.
func CountReverseTradePriceBySqrtPriceX96AndDecimals(sqrtPriceX96 *big.Int, token0Decimals, token1Decimals uint8) (*big.Rat, error) {
	price, err := CountTradePriceBySqrtPriceX96AndDecimals(sqrtPriceX96, token0Decimals, token1Decimals)
	if err != nil {
		return nil, err
	}
	if price.Sign() == 0 {
		return nil, errors.New("cannot calculate reverse price when price is zero")
	}

	return new(big.Rat).Inv(price), nil
}

// CountTradePriceStringBySqrtPriceX96AndDecimals calculates token1 per token0
// and formats it with the requested number of decimal places.
func CountTradePriceStringBySqrtPriceX96AndDecimals(sqrtPriceX96 *big.Int, token0Decimals, token1Decimals uint8, precision int) (string, error) {
	price, err := CountTradePriceBySqrtPriceX96AndDecimals(sqrtPriceX96, token0Decimals, token1Decimals)
	if err != nil {
		return "", err
	}
	if precision < 0 {
		precision = 0
	}

	return price.FloatString(precision), nil
}

// CountTradePriceBigIntBySqrtPriceX96AndDecimals calculates token1 per token0
// and returns it as the project's custom BigInt, scaled by outputDecimals.
//
// Example: if outputDecimals is 18, a returned value of 1500000000000000000
// represents 1.5.
func CountTradePriceBigIntBySqrtPriceX96AndDecimals(sqrtPriceX96 dbtypes.BigInt, token0Decimals, token1Decimals, outputDecimals uint8) (dbtypes.BigInt, error) {
	price, err := CountTradePriceBySqrtPriceX96AndDecimals(sqrtPriceX96.Int, token0Decimals, token1Decimals)
	if err != nil {
		return dbtypes.BigInt{}, err
	}

	scaled := new(big.Rat).Mul(price, new(big.Rat).SetInt(pow10(outputDecimals)))
	result := new(big.Int).Quo(scaled.Num(), scaled.Denom())

	return dbtypes.NewBigInt(result), nil
}

// CountTradePriceBigIntStringBySqrtPriceX96AndDecimals calculates token1 per token0
// as a scaled custom BigInt and formats it back to a human decimal string.
func CountTradePriceBigIntStringBySqrtPriceX96AndDecimals(sqrtPriceX96 dbtypes.BigInt, token0Decimals, token1Decimals, outputDecimals uint8) (string, error) {
	price, err := CountTradePriceBigIntBySqrtPriceX96AndDecimals(sqrtPriceX96, token0Decimals, token1Decimals, outputDecimals)
	if err != nil {
		return "", err
	}

	return FormatScaledBigInt(price, outputDecimals), nil
}

// FormatScaledBigInt formats a custom BigInt scaled by decimals as a decimal string.
func FormatScaledBigInt(value dbtypes.BigInt, decimals uint8) string {
	if value.Int == nil {
		value = dbtypes.NewBigInt(nil)
	}
	if decimals == 0 {
		return value.String()
	}

	sign := ""
	n := new(big.Int).Set(value.Int)
	if n.Sign() < 0 {
		sign = "-"
		n.Abs(n)
	}

	s := n.String()
	dec := int(decimals)
	if len(s) <= dec {
		s = strings.Repeat("0", dec-len(s)+1) + s
	}

	whole := s[:len(s)-dec]
	fraction := strings.TrimRight(s[len(s)-dec:], "0")
	if fraction == "" {
		return sign + whole
	}

	return sign + whole + "." + fraction
}

func pow10(decimals uint8) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
}
