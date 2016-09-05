package expr

// This holt-winters code copied from graphite's functions.py)
// It's "mostly" the same as a standard HW forecast

import (
	"math"
)

func holtWintersIntercept(alpha, actual, last_season, last_intercept, last_slope float64) float64 {
	return alpha*(actual-last_season) + (1-alpha)*(last_intercept+last_slope)
}

func holtWintersSlope(beta, intercept, last_intercept, last_slope float64) float64 {
	return beta*(intercept-last_intercept) + (1-beta)*last_slope
}

func holtWintersSeasonal(gamma, actual, intercept, last_season float64) float64 {
	return gamma*(actual-intercept) + (1-gamma)*last_season

}

func holtWintersAnalysis(series []float64, step int32) []float64 {
	const (
		alpha = 0.1
		beta  = 0.0035
		gamma = 0.1
	)

	// season is currently one day
	season_length := 24 * 60 * 60 / int(step)

	var (
		intercepts  []float64
		slopes      []float64
		seasonals   []float64
		predictions []float64
	)

	getLastSeasonal := func(i int) float64 {
		j := i - season_length
		if j >= 0 {
			return seasonals[j]
		}
		return 0
	}

	var next_pred float64 = math.NaN()

	for i, actual := range series {
		if math.IsNaN(actual) {
			// missing input values break all the math
			// do the best we can and move on
			intercepts = append(intercepts, math.NaN())
			slopes = append(slopes, 0)
			seasonals = append(seasonals, 0)
			predictions = append(predictions, next_pred)
			next_pred = math.NaN()
			continue
		}

		var (
			last_slope     float64
			last_intercept float64
			prediction     float64
		)
		if i == 0 {
			last_intercept = actual
			last_slope = 0
			// seed the first prediction as the first actual
			prediction = actual
		} else {
			last_intercept = intercepts[len(intercepts)-1]
			last_slope = slopes[len(slopes)-1]
			if math.IsNaN(last_intercept) {
				last_intercept = actual
			}
			prediction = next_pred
		}

		last_seasonal := getLastSeasonal(i)
		next_last_seasonal := getLastSeasonal(i + 1)

		intercept := holtWintersIntercept(alpha, actual, last_seasonal, last_intercept, last_slope)
		slope := holtWintersSlope(beta, intercept, last_intercept, last_slope)
		seasonal := holtWintersSeasonal(gamma, actual, intercept, last_seasonal)
		next_pred = intercept + slope + next_last_seasonal

		intercepts = append(intercepts, intercept)
		slopes = append(slopes, slope)
		seasonals = append(seasonals, seasonal)
		predictions = append(predictions, prediction)
	}

	return predictions
}
