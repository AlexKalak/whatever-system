package chaintools

import (
	"math/big"
	"testing"

	"github.com/alexkalak/whatever-system/src/shared/tools/dbtypes"
)

func TestCountTradePriceBySqrtPriceX96AndDecimals(t *testing.T) {
	tests := []struct {
		name           string
		sqrtPriceX96   *big.Int
		token0Decimals uint8
		token1Decimals uint8
		want           string
	}{
		{
			name:           "same decimals at 1 to 1 raw price",
			sqrtPriceX96:   new(big.Int).Lsh(big.NewInt(1), 96),
			token0Decimals: 18,
			token1Decimals: 18,
			want:           "1",
		},
		{
			name:           "token0 has fewer decimals",
			sqrtPriceX96:   new(big.Int).Lsh(big.NewInt(1), 96),
			token0Decimals: 6,
			token1Decimals: 18,
			want:           "1/1000000000000",
		},
		{
			name:           "token0 has more decimals",
			sqrtPriceX96:   new(big.Int).Lsh(big.NewInt(1), 96),
			token0Decimals: 18,
			token1Decimals: 6,
			want:           "1000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountTradePriceBySqrtPriceX96AndDecimals(tt.sqrtPriceX96, tt.token0Decimals, tt.token1Decimals)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.RatString() != tt.want {
				t.Fatalf("got %s, want %s", got.RatString(), tt.want)
			}
		})
	}
}

func TestCountTradePriceStringBySqrtPriceX96AndDecimals(t *testing.T) {
	got, err := CountTradePriceStringBySqrtPriceX96AndDecimals(new(big.Int).Lsh(big.NewInt(1), 96), 6, 18, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "0.000000000001" {
		t.Fatalf("got %s, want 0.000000000001", got)
	}
}

func TestCountTradePriceBigIntBySqrtPriceX96AndDecimals(t *testing.T) {
	sqrtPriceX96 := dbtypes.NewBigInt(new(big.Int).Lsh(big.NewInt(1), 96))

	got, err := CountTradePriceBigIntBySqrtPriceX96AndDecimals(sqrtPriceX96, 6, 18, 18)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "1000000" {
		t.Fatalf("got %s, want 1000000", got.String())
	}
}

func TestCountTradePriceBigIntStringBySqrtPriceX96AndDecimals(t *testing.T) {
	sqrtPriceX96 := dbtypes.NewBigInt(new(big.Int).Lsh(big.NewInt(1), 96))

	got, err := CountTradePriceBigIntStringBySqrtPriceX96AndDecimals(sqrtPriceX96, 6, 18, 18)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "0.000000000001" {
		t.Fatalf("got %s, want 0.000000000001", got)
	}
}

func TestFormatScaledBigInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		decimals uint8
		want     string
	}{
		{name: "whole", value: "1000000000000000000", decimals: 18, want: "1"},
		{name: "fraction", value: "1500000000000000000", decimals: 18, want: "1.5"},
		{name: "small fraction", value: "1000000", decimals: 18, want: "0.000000000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := dbtypes.NewBigIntFromString(tt.value)
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			got := FormatScaledBigInt(value, tt.decimals)
			if got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}
