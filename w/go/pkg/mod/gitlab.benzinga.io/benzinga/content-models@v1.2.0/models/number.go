package models

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// Number represents a numeric type represented as a decimal number.
type Number struct {
	Decimal *decimal.Decimal
}

func (n *Number) String() string {
	if n.Decimal == nil {
		return ""
	}

	return n.Decimal.String()
}

// MarshalJSON makes Number a json.Marshaler
func (n Number) MarshalJSON() ([]byte, error) {
	if n.Decimal == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + n.String() + `"`), nil
}

// UnmarshalJSON makes Number a json.Unmarshaler
func (n *Number) UnmarshalJSON(data []byte) error {
	var str string

	json.Unmarshal(data, &str)

	if str == "" {
		n.Decimal = nil
		return nil
	}

	dec, err := decimal.NewFromString(str)
	if err != nil {
		return err
	}

	n.Decimal = &dec
	return nil
}
