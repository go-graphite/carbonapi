package consolidateBy

import (
	"testing"
	"time"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	th "github.com/go-graphite/carbonapi/tests"
)

var (
	md []interfaces.FunctionMetadata = New("")
)

func init() {
	for _, m := range md {
		metadata.RegisterFunction(m.Name, m.F)
	}
}

func TestConsolidateBy(t *testing.T) {
	now32 := time.Now().Unix()

	tests := []th.EvalTestItem{
		{
			"consolidateBy(metric1,\"sum\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"sum\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "sum").SetConsolidationFunc("sum")},
		},
		{
			"consolidateBy(metric1,\"avg\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"avg\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "avg").SetConsolidationFunc("avg")},
		},
		{
			"consolidateBy(metric1,\"min\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"min\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "min").SetConsolidationFunc("min")},
		},
		{
			"consolidateBy(metric1,\"max\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"max\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "max").SetConsolidationFunc("max")},
		},
		{
			"consolidateBy(metric1,\"first\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"first\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "first").SetConsolidationFunc("first")},
		},
		{
			"consolidateBy(metric1,\"last\")",
			map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {types.MakeMetricData("metric1", []float64{1, 2, 3, 4, 5}, 1, now32)},
			},
			[]*types.MetricData{types.MakeMetricData("consolidateBy(metric1,\"last\")",
				[]float64{1, 2, 3, 4, 5}, 1, now32).SetTag("consolidateBy", "last").SetConsolidationFunc("last")},
		},
	}

	for _, tt := range tests {
		testName := tt.Target
		t.Run(testName, func(t *testing.T) {
			eval := th.EvaluatorFromFunc(md[0].F)
			th.TestEvalExpr(t, eval, &tt)
		})
	}
}

func TestConsolidateByAggregation(t *testing.T) {
	now32 := time.Now().Unix()

	// Проверяем, что родительская функция AggregateFunction сброшена (закешированная функция) и пересчитывается по новому
	tests := []struct {
		name             string
		consolidateFunc  string
		values           []float64
		valuesPerPoint   int
		expectedAggValue float64
	}{
		{
			name:             "max consolidation",
			consolidateFunc:  "max",
			values:           []float64{1, 5, 3, 2, 10, 8},
			valuesPerPoint:   3, // читать как maxDataPoints = 2, значения консолидируются из 3 точек в одну, согласно функции consolidateFunc
			expectedAggValue: 5, // max(1, 5, 3) = 5
		},
		{
			name:             "min consolidation",
			consolidateFunc:  "min",
			values:           []float64{1, 5, 3, 2, 10, 8},
			valuesPerPoint:   3,
			expectedAggValue: 1, // min(1, 5, 3) = 1
		},
		{
			name:             "sum consolidation",
			consolidateFunc:  "sum",
			values:           []float64{1, 2, 3, 4, 5, 6},
			valuesPerPoint:   3,
			expectedAggValue: 6, // sum(1, 2, 3) = 6
		},
		{
			name:             "avg consolidation",
			consolidateFunc:  "avg",
			values:           []float64{3, 6, 8, 4, 5, 6},
			valuesPerPoint:   2,
			expectedAggValue: 4.5, // avg(3, 6) = (3+6)/2 = 9/2 = 4.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем метрику с начальными данными
			metric := types.MakeMetricData("metric1", tt.values, 1, now32)

			// Применяем функцию consolidateBy и maxDataPoints = 2, чтобы запустить консолидацию
			f := &consolidateBy{}
			target := "consolidateBy(metric1,\"" + tt.consolidateFunc + "\")"
			expr, _, err := parser.ParseExpr(target)
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}

			values := map[parser.MetricRequest][]*types.MetricData{
				{Metric: "metric1", From: 0, Until: 1}: {metric},
			}

			result, err := f.Do(nil, th.EvaluatorFromFunc(f), expr, 0, 1, values)
			if err != nil {
				t.Fatalf("consolidateBy failed: %v", err)
			}

			if len(result) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(result))
			}

			r := result[0]

			// Проверяем, что ConsolidationFunc установлен правильно
			if r.ConsolidationFunc != tt.consolidateFunc {
				t.Errorf("Expected ConsolidationFunc=%s, got %s", tt.consolidateFunc, r.ConsolidationFunc)
			}

			// Проверяем, что AggregateFunction nil (должна быть пересчитана по запросу)
			if r.AggregateFunction != nil {
				t.Errorf("Expected AggregateFunction to be nil after consolidateBy, but it was set")
			}

			// Устанавливаем ValuesPerPoint для запуска консолидации
			r.ValuesPerPoint = tt.valuesPerPoint

			// Запускаем агрегацию, вызывая GetAggregateFunction и AggregateValues
			aggFunc := r.GetAggregateFunction()
			if aggFunc == nil {
				t.Fatal("GetAggregateFunction returned nil")
			}

			r.AggregateValues()
			aggValues := r.AggregatedValues()

			if len(aggValues) == 0 {
				t.Fatal("Expected aggregated values, got empty slice")
			}

			// Проверяем, что первое агрегированное значение соответствует ожидаемому
			if aggValues[0] != tt.expectedAggValue {
				t.Errorf("Expected first aggregated value to be %v with %s consolidation, got %v",
					tt.expectedAggValue, tt.consolidateFunc, aggValues[0])
			}
		})
	}
}
