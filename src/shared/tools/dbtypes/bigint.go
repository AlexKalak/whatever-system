package dbtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
)

type BigInt struct {
	*big.Int
}

func NewBigInt(v *big.Int) BigInt {
	if v == nil {
		return BigInt{Int: big.NewInt(0)}
	}
	return BigInt{Int: new(big.Int).Set(v)}
}

func NewBigIntFromString(v string) (BigInt, error) {
	if v == "" {
		return NewBigInt(nil), nil
	}

	n, ok := new(big.Int).SetString(v, 10)
	if !ok {
		return BigInt{}, fmt.Errorf("invalid big.Int value: %s", v)
	}
	return NewBigInt(n), nil
}

func (b BigInt) Value() (driver.Value, error) {
	if b.Int == nil {
		return "0", nil
	}
	return b.String(), nil
}

func (b *BigInt) Scan(value any) error {
	if value == nil {
		b.Int = big.NewInt(0)
		return nil
	}

	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	case int64:
		b.Int = big.NewInt(v)
		return nil
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type BigInt", value)
	}

	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return fmt.Errorf("invalid big.Int value: %s", s)
	}
	b.Int = n
	return nil
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	if b.Int == nil {
		return json.Marshal("0")
	}
	return json.Marshal(b.String())
}

func (b *BigInt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n, err := NewBigIntFromString(s)
	if err != nil {
		return err
	}
	*b = n
	return nil
}
