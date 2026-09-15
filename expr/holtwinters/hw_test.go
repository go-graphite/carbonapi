package holtwinters

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHoltWintersAnalysisNaN(t *testing.T) {
	predictions, deviations := HoltWintersAnalysis([]float64{math.NaN()}, 1, DefaultSeasonality)

	if assert.Len(t, predictions, 1) {
		assert.True(t, math.IsNaN(predictions[0]), "expected NaN prediction, got %v", predictions[0])
	}
	if assert.Len(t, deviations, 1) {
		assert.Zero(t, deviations[0])
	}
}
