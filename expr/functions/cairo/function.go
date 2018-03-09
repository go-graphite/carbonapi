package cairo

// THIS PACKAGE SHOULD NOT BE IMPORTED
// USE IT AS AN EXAMPLE OF HOW TO WRITE NEW FUNCTION

import (
	"github.com/go-graphite/carbonapi/expr/functions/cairo/png"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type example struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &example{}
	functions := []string{"example", "examples"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *example) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return png.EvalExprGraph(e, from, until, values)
}

func (f *example) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"color": {
			Name: "color",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "theColor",
					Required: true,
					Type:     types.String,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Assigns the given color to the seriesList\n\nExample:\n\n.. code-block:: none\n\n  &target=color(collectd.hostname.cpu.0.user, 'green')\n  &target=color(collectd.hostname.cpu.0.system, 'ff0000')\n  &target=color(collectd.hostname.cpu.0.idle, 'gray')\n  &target=color(collectd.hostname.cpu.0.idle, '6464ffaa')",
			Function:    "color(seriesList, theColor)",
			Group:       "Graph",
		},
		"stacked": {
			Name: "stacked",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name: "stack",
					Type: types.String,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes one metric or a wildcard seriesList and change them so they are\nstacked. This is a way of stacking just a couple of metrics without having\nto use the stacked area mode (that stacks everything). By means of this a mixed\nstacked and non stacked graph can be made\n\nIt can also take an optional argument with a name of the stack, in case there is\nmore than one, e.g. for input and output metrics.\n\nExample:\n\n.. code-block:: none\n\n  &target=stacked(company.server.application01.ifconfig.TXPackets, 'tx')",
			Function:    "stacked(seriesLists, stackName='__DEFAULT__')",
			Group:       "Graph",
		},
		"areaBetween": {
			Name: "areaBetween",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Draws the vertical area in between the two series in seriesList. Useful for\nvisualizing a range such as the minimum and maximum latency for a service.\n\nareaBetween expects **exactly one argument** that results in exactly two series\n(see example below). The order of the lower and higher values series does not\nmatter. The visualization only works when used in conjunction with\n``areaMode=stacked``.\n\nMost likely use case is to provide a band within which another metric should\nmove. In such case applying an ``alpha()``, as in the second example, gives\nbest visual results.\n\nExample:\n\n.. code-block:: none\n\n  &target=areaBetween(service.latency.{min,max})&areaMode=stacked\n\n  &target=alpha(areaBetween(service.latency.{min,max}),0.3)&areaMode=stacked\n\nIf for instance, you need to build a seriesList, you should use the ``group``\nfunction, like so:\n\n.. code-block:: none\n\n  &target=areaBetween(group(minSeries(a.*.min),maxSeries(a.*.max)))",
			Function:    "areaBetween(seriesList)",
			Group:       "Graph",
		},
		"alpha": {
			Name: "alpha",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "alpha",
					Required: true,
					Type:     types.Float,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Assigns the given alpha transparency setting to the series. Takes a float value between 0 and 1.",
			Function:    "alpha(seriesList, alpha)",
			Group:       "Graph",
		},
		"dashed": {
			Name: "dashed",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(5),
					Name:    "dashLength",
					Type:    types.Integer,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes one metric or a wildcard seriesList, followed by a float F.\n\nDraw the selected metrics with a dotted line with segments of length F\nIf omitted, the default length of the segments is 5.0\n\nExample:\n\n.. code-block:: none\n\n  &target=dashed(server01.instance01.memory.free,2.5)",
			Function:    "dashed(seriesList, dashLength=5)",
			Group:       "Graph",
		},
		"drawAsInfinite": {
			Name: "drawAsInfinite",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes one metric or a wildcard seriesList.\nIf the value is zero, draw the line at 0.  If the value is above zero, draw\nthe line at infinity. If the value is null or less than zero, do not draw\nthe line.\n\nUseful for displaying on/off metrics, such as exit codes. (0 = success,\nanything else = failure.)\n\nExample:\n\n.. code-block:: none\n\n  drawAsInfinite(Testing.script.exitCode)",
			Function:    "drawAsInfinite(seriesList)",
			Group:       "Graph",
		},
		"secondYAxis": {
			Name: "secondYAxis",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Graph the series on the secondary Y axis.",
			Function:    "secondYAxis(seriesList)",
			Group:       "Graph",
		},
		"lineWidth": {
			Name: "lineWidth",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "width",
					Required: true,
					Type:     types.Float,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes one metric or a wildcard seriesList, followed by a float F.\n\nDraw the selected metrics with a line width of F, overriding the default\nvalue of 1, or the &lineWidth=X.X parameter.\n\nUseful for highlighting a single metric out of many, or having multiple\nline widths in one graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=lineWidth(server01.instance01.memory.free,5)",
			Function:    "lineWidth(seriesList, width)",
			Group:       "Graph",
		},
		"threshold": {
			Name: "threshold",
			Params: []types.FunctionParam{
				{
					Name:     "value",
					Required: true,
					Type:     types.Float,
				},
				{
					Name: "label",
					Type: types.String,
				},
				{
					Name: "color",
					Type: types.String,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes a float F, followed by a label (in double quotes) and a color.\n(See ``bgcolor`` in the render\\_api_ for valid color names & formats.)\n\nDraws a horizontal line at value F across the graph.\n\nExample:\n\n.. code-block:: none\n\n  &target=threshold(123.456, \"omgwtfbbq\", \"red\")",
			Function:    "threshold(value, label=None, color=None)",
			Group:       "Graph",
		},
	}
}
