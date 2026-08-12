package cairo

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func (f *cairo) Description() map[string]types.FunctionDescription {
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

// TODO(civil): Split this into several separate functions.
func (f *cairo) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {

	switch e.Target() {

	case "color": // color(seriesList, theColor)
		arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
		if err != nil {
			return nil, err
		}

		color, err := e.GetStringArg(1) // get color
		if err != nil {
			return nil, err
		}

		results := make([]*types.MetricData, len(arg))

		for i, a := range arg {
			r := a.CopyLinkTags()
			r.Color = color
			results[i] = r
		}

		return results, nil

	case "stacked": // stacked(seriesList, stackname="__DEFAULT__")
		arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
		if err != nil {
			return nil, err
		}

		stackName, err := e.GetStringNamedOrPosArgDefault("stackname", 1, types.DefaultStackName)
		if err != nil {
			return nil, err
		}

		results := make([]*types.MetricData, len(arg))

		for i, a := range arg {
			r := a.CopyLinkTags()
			r.Stacked = true
			r.StackName = stackName
			r.Tags["stacked"] = stackName
			results[i] = r
		}

		return results, nil

	case "areaBetween":
		arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
		if err != nil {
			return nil, err
		}

		if len(arg) != 2 {
			return nil, fmt.Errorf("areaBetween needs exactly two arguments (%d given)", len(arg))
		}

		name := e.Target() + "(" + e.RawArgs() + ")"

		lower := arg[0].CopyTag(name, arg[0].Tags)
		lower.Stacked = true
		lower.StackName = types.DefaultStackName
		lower.Invisible = true

		upper := arg[1].CopyTag(name, arg[1].Tags)
		upper.Stacked = true
		upper.StackName = types.DefaultStackName

		vals := make([]float64, len(upper.Values))

		for i, v := range upper.Values {
			vals[i] = v - lower.Values[i]
		}

		upper.Values = vals

		return []*types.MetricData{lower, upper}, nil

	case "alpha": // alpha(seriesList, theAlpha)
		arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
		if err != nil {
			return nil, err
		}

		alpha, err := e.GetFloatArg(1)
		if err != nil {
			return nil, err
		}

		results := make([]*types.MetricData, len(arg))

		for i, a := range arg {
			r := a.CopyLinkTags()
			r.Alpha = alpha
			r.HasAlpha = true
			results[i] = r
		}

		return results, nil

	case "dashed", "drawAsInfinite", "secondYAxis":
		arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
		if err != nil {
			return nil, err
		}

		results := make([]*types.MetricData, len(arg))

		var dashLen float64
		var dashLenStr string
		if e.Target() == "dashed" {
			dashLen, err := e.GetFloatArgDefault(1, 2.5)
			if err != nil {
				return nil, err
			}
			dashLenStr = strconv.FormatFloat(dashLen, 'g', -1, 64)
		}

		for i, a := range arg {
			r := a.CopyLink()
			r.Name = e.Target() + "(" + a.Name + ")"

			switch e.Target() {
			case "dashed":
				r.Dashed = dashLen
				r.Tags["dashed"] = dashLenStr
			case "drawAsInfinite":
				r.DrawAsInfinite = true
				r.Tags["drawAsInfinite"] = "1"
			case "secondYAxis":
				r.SecondYAxis = true
				r.Tags["secondYAxis"] = "1"
			}

			results[i] = r
		}
		return results, nil

	case "lineWidth": // lineWidth(seriesList, width)
		arg, err := helper.GetSeriesArg(ctx, eval, e.Arg(0), from, until, values)
		if err != nil {
			return nil, err
		}

		width, err := e.GetFloatArg(1)
		if err != nil {
			return nil, err
		}

		results := make([]*types.MetricData, len(arg))

		for i, a := range arg {
			r := a.CopyLinkTags()
			r.LineWidth = width
			r.HasLineWidth = true
			results[i] = r
		}

		return results, nil

	case "threshold": // threshold(value, label=None, color=None)
		// XXX does not match graphite's signature
		// BUG(nnuss): the signature *does* match but there is an edge case because of named argument handling if you use it *just* wrong:
		//			   threshold(value, "gold", label="Aurum")
		//			   will result in:
		//			   value = value
		//			   label = "Aurum" (by named argument)
		//			   color = "" (by default as len(positionalArgs) == 2 and there is no named 'color' arg)

		value, err := e.GetFloatArg(0)
		if err != nil {
			return nil, err
		}

		defaultLabel := e.Arg(0).StringValue()

		name, err := e.GetStringNamedOrPosArgDefault("label", 1, defaultLabel)
		if err != nil {
			return nil, err
		}

		color, err := e.GetStringNamedOrPosArgDefault("color", 2, "")
		if err != nil {
			return nil, err
		}

		newValues := []float64{value, value}
		stepTime := until - from
		stopTime := from + stepTime*int64(len(newValues))
		p := &types.MetricData{
			FetchResponse: pb.FetchResponse{
				Name:              name,
				StartTime:         from,
				StopTime:          stopTime,
				StepTime:          stepTime,
				Values:            newValues,
				ConsolidationFunc: "average",
			},
			Tags:         map[string]string{"name": name},
			GraphOptions: types.GraphOptions{Color: color},
		}

		return []*types.MetricData{p}, nil

	}

	return nil, helper.ErrUnknownFunction(e.Target())
}
