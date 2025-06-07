package utils

import "github.com/shopspring/decimal"

func DecimalToInt(dec decimal.Decimal) int {
	return int(dec.Mul(decimal.NewFromInt(100)).IntPart())
}
