package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func asDecimal(s string) *decimal.Decimal {
	dec, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return &dec
}

func TestNumber(t *testing.T) {
	tests := []struct {
		val *decimal.Decimal
	}{
		{asDecimal("0.0001")},
		{asDecimal("99999999999.99999")},
		{asDecimal("-1")},
		{nil},
	}

	for _, test := range tests {
		// Round 1: Native -> JSON
		num := Number{test.val}
		data, err := num.MarshalJSON()
		assert.Nil(t, err)

		// Round 2: JSON -> Native
		num2 := Number{}
		err = num2.UnmarshalJSON(data)
		assert.Nil(t, err)

		// Round 3: Native -> JSON
		data2, err := num2.MarshalJSON()
		assert.Nil(t, err)

		// Ensure equality.
		assert.Equal(t, num, num2)
		assert.Equal(t, data, data2)
		assert.Equal(t, num.String(), num2.String())
	}

	num := Number{}
	err := num.UnmarshalJSON([]byte(`"0.1.2.3"`))
	require.NotNil(t, err)
	assert.Equal(t, "can't convert 0.1.2.3 to decimal: too many .s", err.Error())
}
