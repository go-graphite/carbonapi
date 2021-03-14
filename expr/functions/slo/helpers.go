package slo

import (
	"fmt"

	"github.com/go-graphite/carbonapi/pkg/parser"
)

func (f *slo) buildDataPoint(bucketQtyMatched, bucketQtyNotNull int) (isAbsent bool, value float64) {
	if bucketQtyNotNull == 0 {
		isAbsent = true
		value = 0.0
	} else {
		isAbsent = false
		value = float64(bucketQtyMatched) / float64(bucketQtyNotNull)
	}
	return
}

func (f *slo) buildMethod(e parser.Expr, argNumber int, value float64) (func(float64) bool, string, error) {
	var methodFoo func(float64) bool = nil

	methodName, err := e.GetStringArg(argNumber)
	if err != nil {
		return nil, methodName, err
	}

	if methodName == "above" {
		methodFoo = func(testedValue float64) bool {
			return testedValue > value
		}
	}

	if methodName == "aboveOrEqual" {
		methodFoo = func(testedValue float64) bool {
			return testedValue >= value
		}
	}

	if methodName == "below" {
		methodFoo = func(testedValue float64) bool {
			return testedValue < value
		}
	}

	if methodName == "belowOrEqual" {
		methodFoo = func(testedValue float64) bool {
			return testedValue <= value
		}
	}

	if methodFoo == nil {
		return nil, methodName, fmt.Errorf("unknown method `%s`", methodName)
	}

	return methodFoo, methodName, nil
}
