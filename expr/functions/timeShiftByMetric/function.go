package timeShiftByMetric

import (
	"context"
	"math"
	"regexp"
	"strings"

	"github.com/ansel1/merry"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type offsetByVersion map[string]int64

type timeShiftByMetric struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{
		F:    &timeShiftByMetric{},
		Name: "timeShiftByMetric",
	}}
}

// timeShiftByMetric(seriesList, markSource, versionRankIndex)
func (f *timeShiftByMetric) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (resultData []*types.MetricData, resultError error) {
	params, err := f.extractCallParams(ctx, e, from, until, values)
	if err != nil {
		return nil, err
	}

	latestMarks, err := f.locateLatestMarks(params)
	if err != nil {
		return nil, err
	}

	offsets := f.calculateOffsets(params, latestMarks)

	result := f.applyShift(params, offsets)

	return result, nil
}

// applyShift shifts timeline of those metrics which major version is less than top major version
func (f *timeShiftByMetric) applyShift(params *callParams, offsets offsetByVersion) []*types.MetricData {
	result := make([]*types.MetricData, 0, len(params.metrics))
	for _, metric := range params.metrics {
		offsetIsSet := false
		offset := int64(0)
		var possibleVersion string

		name := metric.Tags["name"]
		nameSplit := strings.Split(name, ".")

		// make sure that there is desired rank at all
		if params.versionRank >= len(nameSplit) {
			continue
		}
		possibleVersion = nameSplit[params.versionRank]

		if possibleOffset, ok := offsets[possibleVersion]; !ok {
			for key, value := range offsets {
				if strings.HasPrefix(key, possibleVersion) {
					offset = value
					offsetIsSet = true
					offsets[possibleVersion] = value
				}
			}
		} else {
			offset = possibleOffset
			offsetIsSet = true
		}

		// checking if it is some version after all, otherwise this series will be omitted
		if offsetIsSet {
			r := metric.CopyLinkTags()
			r.Name = "timeShiftByMetric(" + r.Name + ")"
			r.StopTime += offset
			r.StartTime += offset

			result = append(result, r)
		}
	}

	return result
}

func (f *timeShiftByMetric) calculateOffsets(params *callParams, versions versionInfos) offsetByVersion {
	result := make(offsetByVersion)
	topPosition := versions[0].position

	for _, version := range versions {
		result[version.mark] = int64(topPosition-version.position) * params.stepTime
	}

	return result
}

// extractCallParams (preliminarily) validates and extracts parameters of timeShiftByMetric's call as structure
func (f *timeShiftByMetric) extractCallParams(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (*callParams, error) {
	metrics, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	marks, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
	if err != nil {
		return nil, err
	}

	versionRank, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}

	// validating data sets: both metrics and marks must have at least 2 series each
	// also, all IsAbsent and Values lengths must be equal to each other
	pointsQty := -1
	stepTime := int64(-1)
	var dataSets map[string][]*types.MetricData = map[string][]*types.MetricData{
		"marks":   marks,
		"metrics": metrics,
	}
	for name, dataSet := range dataSets {
		if len(dataSet) < 2 {
			return nil, merry.WithMessagef(errTooFewDatasets, "bad data: need at least 2 %s data sets to process, got %d", name, len(dataSet))
		}

		for _, series := range dataSet {
			if pointsQty == -1 {
				pointsQty = len(series.Values)
				if pointsQty == 0 {
					return nil, merry.WithMessagef(errEmptySeries, "bad data: empty series %s", series.Name)
				}
			} else if pointsQty != len(series.Values) {
				return nil, merry.WithMessagef(errSeriesLengthMismatch, "bad data: length of Values for series %s differs from others", series.Name)
			}

			if stepTime == -1 {
				stepTime = series.StepTime
			}
		}
	}

	result := &callParams{
		metrics:     metrics,
		marks:       marks,
		versionRank: versionRank,
		pointsQty:   pointsQty,
		stepTime:    stepTime,
	}
	return result, nil
}

var reLocateMark *regexp.Regexp = regexp.MustCompile(`(\d+)_(\d+)`)

// locateLatestMarks gets the series with marks those look like "65_4"
// and looks for the latest ones by _major_ versions
// e.g. among set [63_0, 64_0, 64_1, 64_2, 65_0, 65_1] it locates 63_0, 64_4 and 65_1
// returns located elements
func (f *timeShiftByMetric) locateLatestMarks(params *callParams) (versionInfos, error) {

	versions := make(versionInfos, 0, len(params.marks))

	// noinspection SpellCheckingInspection
	for _, mark := range params.marks {
		markSplit := strings.Split(mark.Tags["name"], ".")
		markVersion := markSplit[len(markSplit)-1]

		// for mark that matches pattern (\d+)_(\d+), this should return slice of 3 strings exactly
		submatch := reLocateMark.FindStringSubmatch(markVersion)
		if len(submatch) != 3 {
			continue
		}

		position := -1
		for i := params.pointsQty - 1; i >= 0; i-- {
			if !math.IsNaN(mark.Values[i]) {
				position = i
				break
			}
		}

		if position == -1 {
			// weird, but mark series has no data in it - skipping
			continue
		}
		// collecting all marks found
		versions = append(versions, versionInfo{
			mark:         markVersion,
			position:     position,
			versionMajor: mustAtoi(submatch[1]),
			versionMinor: mustAtoi(submatch[2]),
		})
	}

	// obtain top versions for each major version
	result := versions.HighestVersions()
	if len(result) < 2 {
		return nil, merry.WithMessagef(errLessThan2Marks, "bad data: could not find 2 marks, only %d found", len(result))
	} else {
		return result, nil
	}
}

func (f *timeShiftByMetric) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"timeShiftByMetric": types.FunctionDescription{
			Description: "Takes a seriesList with wildcard in versionRankIndex rank and applies shift to the closest version from markSource\n\n.. code-block:: none\n\n  &target=timeShiftByMetric(carbon.agents.graphite.creates)",
			Function:    "timeShiftByMetric(seriesList, markSource, versionRankIndex)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "timeShiftByMetric",
			Params: []types.FunctionParam{
				types.FunctionParam{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				types.FunctionParam{
					Name:     "markSource",
					Required: true,
					Type:     types.SeriesList,
				},
				types.FunctionParam{
					Name:     "versionRankIndex",
					Required: true,
					Type:     types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
