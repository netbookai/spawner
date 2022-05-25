package common

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func Get100thOfCentsInIntegerForDollar(cost float64) int64 {
	// rounding dollar cost to 4 decimals to convert it further into 1/100 of cents
	cost = RoundTo(cost, 4)

	// returning cost in 1/100 of cents
	return int64(cost * 10000)
}

func ConverDecimalCostMapToIntCostMap(costMap map[string]decimal.Decimal) (map[string]int64, error) {
	costMapInt := make(map[string]int64)

	for key, value := range costMap {

		costInFloat, ok := value.Float64()
		if !ok {
			return nil, fmt.Errorf("failed to convert cost to float64, costInDecimal: %v", value)
		}

		costIn100thCents := Get100thOfCentsInIntegerForDollar(costInFloat)

		costMapInt[key] = costIn100thCents

	}

	return costMapInt, nil
}

func ConverDecimalCostMapOfMapToIntCostMapOfMap(costMap map[string]map[string]decimal.Decimal) (map[string]map[string]int64, error) {
	costMapInt := make(map[string]map[string]int64)

	for key, valueMap := range costMap {

		costMapInt[key] = make(map[string]int64)

		for k, v := range valueMap {
			costInFloat, ok := v.Float64()
			if !ok {
				return nil, fmt.Errorf("failed to convert cost to float64, costInDecimal: %v", v)
			}

			costIn100thCents := Get100thOfCentsInIntegerForDollar(costInFloat)

			costMapInt[key][k] = costIn100thCents

		}

	}

	return costMapInt, nil
}
