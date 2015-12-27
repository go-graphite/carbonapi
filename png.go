// +build !cairo

package main

import (
	"bytes"
	"image/color"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	vgdraw "github.com/gonum/plot/vg/draw"
)

func marshalPNG(r *http.Request, results []*metricData) []byte {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	// set bg/fg colors
	bgcolor := string2Color(getString(r.FormValue("bgcolor"), "black"))
	p.BackgroundColor = bgcolor

	fgcolorstr := getString(r.FormValue("fgcolor"), "white")
	fgcolor := string2Color(fgcolorstr)
	p.Title.Color = fgcolor
	p.X.LineStyle.Color = fgcolor
	p.Y.LineStyle.Color = fgcolor
	p.X.Tick.LineStyle.Color = fgcolor
	p.Y.Tick.LineStyle.Color = fgcolor
	p.X.Tick.Label.Color = fgcolor
	p.Y.Tick.Label.Color = fgcolor
	p.X.Label.Color = fgcolor
	p.Y.Label.Color = fgcolor
	p.Legend.Color = fgcolor

	// set grid
	grid := plotter.NewGrid()
	grid.Vertical.Color = fgcolor
	grid.Horizontal.Color = fgcolor
	p.Add(grid)

	// line mode (ikruglow) TODO check values
	lineMode := getString(r.FormValue("lineMode"), "slope")

	// width and height
	width := getFloat64(r.FormValue("width"), 330)
	height := getFloat64(r.FormValue("height"), 250)

	// need different timeMarker's based on step size
	p.Title.Text = r.FormValue("title")
	if len(results) > 0 {
		p.X.Tick.Marker = NewTimeMarker(results[0].GetStepTime())
	}

	hideLegend := getBool(r.FormValue("hideLegend"), false)

	graphOnly := getBool(r.FormValue("graphOnly"), false)
	if graphOnly {
		p.HideAxes()
	}

	if len(results) == 1 && results[0].color == "" {
		results[0].color = fgcolorstr
	}

	var lines []plot.Plotter
	for i, r := range results {
		if r == nil {
			continue
		}
		l := NewResponsePlotter(r)

		l.LineStyle.Color = fgcolor

		// consolidate datapoints
		l.maybeConsolidateData(width)

		if r.drawAsInfinite {
			l.lineMode = "drawAsInfinite"
		} else {
			l.lineMode = lineMode
		}

		if r.color != "" {
			l.Color = string2Color(r.color)
		} else {
			l.Color = plotutil.Color(i)
		}

		lines = append(lines, l)

		if !graphOnly && !hideLegend {
			p.Legend.Add(r.GetName(), l)
		}
	}

	p.Add(lines...)

	p.Y.Max *= 1.05
	p.Y.Min *= 0.95

	writerTo, err := p.WriterTo(vg.Points(width), vg.Points(height), "png")
	var buffer bytes.Buffer
	if _, err := writerTo.WriteTo(&buffer); err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

type TimeMarker struct {
	format string
}

func NewTimeMarker(step int32) TimeMarker {
	var format string

	// heuristic yoinked from graphite, more or less
	switch {
	case step < 5:
		format = "15:04:05"
	case step < 60:
		format = "15:04"
	case step < 100:
		format = "Mon 3PM"
	case step < 255:
		format = "01/02 3PM"
	case step < 32000:
		format = "01/02"
	case step < 120000:
		format = "01/02 2006"
	}

	return TimeMarker{format}
}

func (tm TimeMarker) Ticks(min, max float64) []plot.Tick {
	ticks := []plot.Tick{
		plot.Tick{
			Value: min,
		},
		plot.Tick{
			Value: max,
		},
	}

	for i, t := range ticks {
		if !t.IsMinor() {
			t0 := time.Unix(int64(t.Value), 0)
			ticks[i].Label = t0.Format(tm.format)
		}
	}

	return ticks
}

func string2Color(clr string) color.Color {
	if c, ok := colors[clr]; ok {
		return c
	}
	return hexToColor(clr)
}

// https://code.google.com/p/sadbox/source/browse/color/hex.go
// hexToColor converts an Hex string to a RGB triple.
func hexToColor(h string) color.Color {
	var r, g, b uint8
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}

	if len(h) == 3 {
		h = h[:1] + h[:1] + h[1:2] + h[1:2] + h[2:] + h[2:]
	}

	if len(h) == 6 {
		if rgb, err := strconv.ParseUint(string(h), 16, 32); err == nil {
			r = uint8(rgb >> 16)
			g = uint8((rgb >> 8) & 0xFF)
			b = uint8(rgb & 0xFF)
		}
	}

	return color.RGBA{r, g, b, 255}
}

type ResponsePlotter struct {
	Response *metricData
	vgdraw.LineStyle
	lineMode string
}

func NewResponsePlotter(r *metricData) *ResponsePlotter {
	return &ResponsePlotter{
		Response:  r,
		LineStyle: plotter.DefaultLineStyle,
	}
}

// Plot draws the Line, implementing the plot.Plotter interface.
func (rp *ResponsePlotter) Plot(canvas vgdraw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&canvas)

	start := float64(rp.Response.GetStartTime())
	step := float64(rp.Response.GetStepTime())
	absent := rp.Response.IsAbsent

	lines := make([][]vgdraw.Point, 1)
	lines[0] = make([]vgdraw.Point, 0, len(rp.Response.Values))

	/* ikruglov
	 * swithing between lineMode and looping inside
	 * is more branch-prediction friendly i.e. potentially faster */
	switch rp.lineMode {
	case "slope":
		currentLine := 0
		lastAbsent := false
		for i, v := range rp.Response.Values {
			if absent[i] {
				lastAbsent = true
			} else if lastAbsent {
				currentLine++
				lines = append(lines, make([]vgdraw.Point, 1))
				lines[currentLine][0] = vgdraw.Point{X: trX(start + float64(i)*step), Y: trY(v)}
				lastAbsent = false
			} else {
				lines[currentLine] = append(lines[currentLine], vgdraw.Point{X: trX(start + float64(i)*step), Y: trY(v)})
			}
		}

	case "connected":
		for i, v := range rp.Response.Values {
			if absent[i] {
				continue
			}

			lines[0] = append(lines[0], vgdraw.Point{X: trX(start + float64(i)*step), Y: trY(v)})
		}

	case "drawAsInfinite":
		for i, v := range rp.Response.Values {
			if !absent[i] && v > 0 {
				infiniteLine := []vgdraw.Point{
					vgdraw.Point{X: trX(start + float64(i)*step), Y: canvas.Y(1)},
					vgdraw.Point{X: trX(start + float64(i)*step), Y: canvas.Y(0)},
				}
				lines = append(lines, infiniteLine)
			}
		}

	//case "staircase": // TODO
	default:
		panic("Unimplemented " + rp.lineMode)
	}

	canvas.StrokeLines(rp.LineStyle, lines...)
}

func (rp *ResponsePlotter) Thumbnail(canvas *vgdraw.Canvas) {
	l := plotter.Line{LineStyle: rp.LineStyle}
	l.Thumbnail(canvas)
}

func (rp *ResponsePlotter) DataRange() (xmin, xmax, ymin, ymax float64) {
	ymin = math.Inf(1)
	ymax = math.Inf(-1)
	absent := rp.Response.IsAbsent

	xmin = float64(rp.Response.GetStartTime())
	xmax = float64(rp.Response.GetStopTime())

	// same as rp.lineMode == "drawAsInfinite"
	if rp.Response.drawAsInfinite {
		ymin = 0
		ymax = 1
		return
	}

	for i, v := range rp.Response.Values {
		if absent[i] {
			continue
		}
		ymin = math.Min(ymin, v)
		ymax = math.Max(ymax, v)
	}
	return
}

func (rp *ResponsePlotter) maybeConsolidateData(numberOfPixels float64) {
	// idealy numberOfPixels should be size in pixels of char ares,
	// not char areay with Y axis and its label

	numberOfDataPoints := len(rp.Response.Values)
	pointsPerPixel := int(math.Ceil(float64(numberOfDataPoints) / numberOfPixels))

	if pointsPerPixel <= 1 {
		return
	}

	newNumberOfDataPoints := (numberOfDataPoints / pointsPerPixel) + 1
	values := make([]float64, newNumberOfDataPoints)
	absent := make([]bool, newNumberOfDataPoints)

	k := 0
	step := pointsPerPixel
	for i := 0; i < numberOfDataPoints; i += step {
		if i+step < numberOfDataPoints {
			values[k], absent[k] = consolidateAvg(rp.Response.Values[i:i+step], rp.Response.IsAbsent[i:i+step])
		} else {
			values[k], absent[k] = consolidateAvg(rp.Response.Values[i:], rp.Response.IsAbsent[i:])
		}

		k++
	}

	stepTime := rp.Response.GetStepTime()
	stepTime *= int32(pointsPerPixel)

	rp.Response.Values = values[:k]
	rp.Response.IsAbsent = absent[:k]
	rp.Response.StepTime = proto.Int32(stepTime)
}

func consolidateAvg(v []float64, a []bool) (float64, bool) {
	cnt := len(v)
	if cnt == 0 {
		return 0.0, true
	}

	abs := true
	var val float64
	var elts int

	for i := 0; i < cnt; i++ {
		if !a[i] {
			abs = false
			val += v[i]
			elts++
		}
	}

	// Don't need to check for divide-by-zero. We'll just return NaN and the point will be marked absent.
	return val / float64(elts), abs
}
