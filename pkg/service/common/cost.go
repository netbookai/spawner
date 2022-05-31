package common

import (
	"github.com/shopspring/decimal"
)

func Get100thOfCentsInIntegerForDollar(cost decimal.Decimal) int64 {

	// returning cost in 1/100 of cents
	return cost.Mul(decimal.NewFromInt32(10000)).IntPart()
}

func ConverDecimalCostMapToIntCostMap(costMap map[string]decimal.Decimal) (map[string]int64, error) {
	costMapInt := make(map[string]int64)

	for key, value := range costMap {

		costIn100thCents := Get100thOfCentsInIntegerForDollar(value)

		costMapInt[key] = costIn100thCents

	}

	return costMapInt, nil
}

func ConverDecimalCostMapOfMapToIntCostMapOfMap(costMap map[string]map[string]decimal.Decimal) (map[string]map[string]int64, error) {
	costMapInt := make(map[string]map[string]int64)

	for key, valueMap := range costMap {

		costMapInt[key] = make(map[string]int64)

		for k, v := range valueMap {

			costIn100thCents := Get100thOfCentsInIntegerForDollar(v)

			costMapInt[key][k] = costIn100thCents

		}

	}

	return costMapInt, nil
}
