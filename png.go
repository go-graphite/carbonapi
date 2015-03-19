package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"strconv"
	"time"

	"code.google.com/p/gogoprotobuf/proto"
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"code.google.com/p/plotinum/vg/vgimg"
)

var linesColors = `blue,green,red,purple,brown,yellow,aqua,grey,magenta,pink,gold,rose`

func marshalPNG(r *http.Request, results []*metricData) []byte {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	// set bg/fg colors
	bgcolor := string2Color(getString(r.FormValue("bgcolor"), "black"))
	p.BackgroundColor = bgcolor

	fgcolor := string2Color(getString(r.FormValue("fgcolor"), "white"))
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
	width := getInt(r.FormValue("width"), 330)
	height := getInt(r.FormValue("height"), 250)

	// need different timeMarker's based on step size
	p.Title.Text = r.FormValue("title")
	p.X.Tick.Marker = makeTimeMarker(results[0].GetStepTime())

	hideLegend := getBool(r.FormValue("hideLegend"), false)

	graphOnly := getBool(r.FormValue("graphOnly"), false)
	if graphOnly {
		p.HideAxes()
	}

	var lines []plot.Plotter
	for i, r := range results {
		l := NewResponsePlotter(r)

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

	// Draw the plot to an in-memory image.
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	da := plot.MakeDrawArea(vgimg.NewImage(img))
	p.Draw(da)

	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		panic(err)
	}

	return b.Bytes()
}

func makeTimeMarker(step int32) func(min, max float64) []plot.Tick {

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

	return func(min, max float64) []plot.Tick {
		ticks := plot.DefaultTicks(min, max)

		for i, t := range ticks {
			if !t.IsMinor() {
				t0 := time.Unix(int64(t.Value), 0)
				ticks[i].Label = t0.Format(format)
			}
		}

		return ticks

	}
}

func getBool(s string, def bool) bool {
	if s == "" {
		return def
	}

	switch s {
	case "True", "true", "1":
		return true
	case "False", "false", "0":
		return false
	}

	return def
}

func getString(s string, def string) string {
	if s == "" {
		return def
	}

	return s
}

func getInt(s string, def int) int {
	if s == "" {
		return def
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}

	return n
}

var colors = map[string]color.RGBA{
	"black":                color.RGBA{0x00, 0x00, 0x00, 0xff},
	"navy":                 color.RGBA{0x00, 0x00, 0x80, 0xff},
	"darkblue":             color.RGBA{0x00, 0x00, 0x8b, 0xff},
	"mediumblue":           color.RGBA{0x00, 0x00, 0xcd, 0xff},
	"blue":                 color.RGBA{0x00, 0x00, 0xff, 0xff},
	"darkgreen":            color.RGBA{0x00, 0x64, 0x00, 0xff},
	"green":                color.RGBA{0x00, 0x80, 0x00, 0xff},
	"teal":                 color.RGBA{0x00, 0x80, 0x80, 0xff},
	"darkcyan":             color.RGBA{0x00, 0x8b, 0x8b, 0xff},
	"deepskyblue":          color.RGBA{0x00, 0xbf, 0xff, 0xff},
	"darkturquoise":        color.RGBA{0x00, 0xce, 0xd1, 0xff},
	"mediumspringgreen":    color.RGBA{0x00, 0xfa, 0x9a, 0xff},
	"lime":                 color.RGBA{0x00, 0xff, 0x00, 0xff},
	"springgreen":          color.RGBA{0x00, 0xff, 0x7f, 0xff},
	"aqua":                 color.RGBA{0x00, 0xff, 0xff, 0xff},
	"cyan":                 color.RGBA{0x00, 0xff, 0xff, 0xff},
	"midnightblue":         color.RGBA{0x19, 0x19, 0x70, 0xff},
	"dodgerblue":           color.RGBA{0x1e, 0x90, 0xff, 0xff},
	"lightseagreen":        color.RGBA{0x20, 0xb2, 0xaa, 0xff},
	"forestgreen":          color.RGBA{0x22, 0x8b, 0x22, 0xff},
	"seagreen":             color.RGBA{0x2e, 0x8b, 0x57, 0xff},
	"darkslategray":        color.RGBA{0x2f, 0x4f, 0x4f, 0xff},
	"limegreen":            color.RGBA{0x32, 0xcd, 0x32, 0xff},
	"mediumseagreen":       color.RGBA{0x3c, 0xb3, 0x71, 0xff},
	"turquoise":            color.RGBA{0x40, 0xe0, 0xd0, 0xff},
	"royalblue":            color.RGBA{0x41, 0x69, 0xe1, 0xff},
	"steelblue":            color.RGBA{0x46, 0x82, 0xb4, 0xff},
	"darkslateblue":        color.RGBA{0x48, 0x3d, 0x8b, 0xff},
	"mediumturquoise":      color.RGBA{0x48, 0xd1, 0xcc, 0xff},
	"indigo":               color.RGBA{0x4b, 0x00, 0x82, 0xff},
	"darkolivegreen":       color.RGBA{0x55, 0x6b, 0x2f, 0xff},
	"cadetblue":            color.RGBA{0x5f, 0x9e, 0xa0, 0xff},
	"cornflowerblue":       color.RGBA{0x64, 0x95, 0xed, 0xff},
	"mediumaquamarine":     color.RGBA{0x66, 0xcd, 0xaa, 0xff},
	"dimgray":              color.RGBA{0x69, 0x69, 0x69, 0xff},
	"slateblue":            color.RGBA{0x6a, 0x5a, 0xcd, 0xff},
	"olivedrab":            color.RGBA{0x6b, 0x8e, 0x23, 0xff},
	"slategray":            color.RGBA{0x70, 0x80, 0x90, 0xff},
	"lightslategray":       color.RGBA{0x77, 0x88, 0x99, 0xff},
	"mediumslateblue":      color.RGBA{0x7b, 0x68, 0xee, 0xff},
	"lawngreen":            color.RGBA{0x7c, 0xfc, 0x00, 0xff},
	"chartreuse":           color.RGBA{0x7f, 0xff, 0x00, 0xff},
	"aquamarine":           color.RGBA{0x7f, 0xff, 0xd4, 0xff},
	"lavender":             color.RGBA{0xe6, 0xe6, 0xfa, 0xff},
	"darksalmon":           color.RGBA{0xe9, 0x96, 0x7a, 0xff},
	"violet":               color.RGBA{0xee, 0x82, 0xee, 0xff},
	"palegoldenrod":        color.RGBA{0xee, 0xe8, 0xaa, 0xff},
	"lightcoral":           color.RGBA{0xf0, 0x80, 0x80, 0xff},
	"khaki":                color.RGBA{0xf0, 0xe6, 0x8c, 0xff},
	"aliceblue":            color.RGBA{0xf0, 0xf8, 0xff, 0xff},
	"honeydew":             color.RGBA{0xf0, 0xff, 0xf0, 0xff},
	"azure":                color.RGBA{0xf0, 0xff, 0xff, 0xff},
	"sandybrown":           color.RGBA{0xf4, 0xa4, 0x60, 0xff},
	"wheat":                color.RGBA{0xf5, 0xde, 0xb3, 0xff},
	"beige":                color.RGBA{0xf5, 0xf5, 0xdc, 0xff},
	"whitesmoke":           color.RGBA{0xf5, 0xf5, 0xf5, 0xff},
	"mintcream":            color.RGBA{0xf5, 0xff, 0xfa, 0xff},
	"ghostwhite":           color.RGBA{0xf8, 0xf8, 0xff, 0xff},
	"salmon":               color.RGBA{0xfa, 0x80, 0x72, 0xff},
	"antiquewhite":         color.RGBA{0xfa, 0xeb, 0xd7, 0xff},
	"linen":                color.RGBA{0xfa, 0xf0, 0xe6, 0xff},
	"lightgoldenrodyellow": color.RGBA{0xfa, 0xfa, 0xd2, 0xff},
	"oldlace":              color.RGBA{0xfd, 0xf5, 0xe6, 0xff},
	"red":                  color.RGBA{0xff, 0x00, 0x00, 0xff},
	"fuchsia":              color.RGBA{0xff, 0x00, 0xff, 0xff},
	"magenta":              color.RGBA{0xff, 0x00, 0xff, 0xff},
	"deeppink":             color.RGBA{0xff, 0x14, 0x93, 0xff},
	"orangered":            color.RGBA{0xff, 0x45, 0x00, 0xff},
	"tomato":               color.RGBA{0xff, 0x63, 0x47, 0xff},
	"hotpink":              color.RGBA{0xff, 0x69, 0xb4, 0xff},
	"coral":                color.RGBA{0xff, 0x7f, 0x50, 0xff},
	"darkorange":           color.RGBA{0xff, 0x8c, 0x00, 0xff},
	"lightsalmon":          color.RGBA{0xff, 0xa0, 0x7a, 0xff},
	"orange":               color.RGBA{0xff, 0xa5, 0x00, 0xff},
	"lightpink":            color.RGBA{0xff, 0xb6, 0xc1, 0xff},
	"pink":                 color.RGBA{0xff, 0xc0, 0xcb, 0xff},
	"gold":                 color.RGBA{0xff, 0xd7, 0x00, 0xff},
	"peachpuff":            color.RGBA{0xff, 0xda, 0xb9, 0xff},
	"navajowhite":          color.RGBA{0xff, 0xde, 0xad, 0xff},
	"moccasin":             color.RGBA{0xff, 0xe4, 0xb5, 0xff},
	"bisque":               color.RGBA{0xff, 0xe4, 0xc4, 0xff},
	"mistyrose":            color.RGBA{0xff, 0xe4, 0xe1, 0xff},
	"blanchedalmond":       color.RGBA{0xff, 0xeb, 0xcd, 0xff},
	"papayawhip":           color.RGBA{0xff, 0xef, 0xd5, 0xff},
	"lavenderblush":        color.RGBA{0xff, 0xf0, 0xf5, 0xff},
	"seashell":             color.RGBA{0xff, 0xf5, 0xee, 0xff},
	"cornsilk":             color.RGBA{0xff, 0xf8, 0xdc, 0xff},
	"lemonchiffon":         color.RGBA{0xff, 0xfa, 0xcd, 0xff},
	"floralwhite":          color.RGBA{0xff, 0xfa, 0xf0, 0xff},
	"snow":                 color.RGBA{0xff, 0xfa, 0xfa, 0xff},
	"yellow":               color.RGBA{0xff, 0xff, 0x00, 0xff},
	"lightyellow":          color.RGBA{0xff, 0xff, 0xe0, 0xff},
	"ivory":                color.RGBA{0xff, 0xff, 0xf0, 0xff},
	"white":                color.RGBA{0xff, 0xff, 0xff, 0xff},
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
	plot.LineStyle
	lineMode string
}

func NewResponsePlotter(r *metricData) *ResponsePlotter {
	return &ResponsePlotter{
		Response:  r,
		LineStyle: plotter.DefaultLineStyle,
	}
}

// Plot draws the Line, implementing the plot.Plotter interface.
func (rp *ResponsePlotter) Plot(da plot.DrawArea, plt *plot.Plot) {
	trX, trY := plt.Transforms(&da)

	start := float64(rp.Response.GetStartTime())
	step := float64(rp.Response.GetStepTime())
	absent := rp.Response.IsAbsent

	lines := make([][]plot.Point, 1)
	lines[0] = make([]plot.Point, 0, len(rp.Response.Values))

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
				lines = append(lines, make([]plot.Point, 1))
				lines[currentLine][0] = plot.Point{X: trX(start + float64(i)*step), Y: trY(v)}
				lastAbsent = false
			} else {
				lines[currentLine] = append(lines[currentLine], plot.Point{X: trX(start + float64(i)*step), Y: trY(v)})
			}
		}

	case "connected":
		for i, v := range rp.Response.Values {
			if absent[i] {
				continue
			}

			lines[0] = append(lines[0], plot.Point{X: trX(start + float64(i)*step), Y: trY(v)})
		}

	case "drawAsInfinite":
		for i, v := range rp.Response.Values {
			if !absent[i] && v > 0 {
				infiniteLine := []plot.Point{
					plot.Point{X: trX(start + float64(i)*step), Y: da.Y(1)},
					plot.Point{X: trX(start + float64(i)*step), Y: da.Y(0)},
				}
				lines = append(lines, infiniteLine)
			}
		}

	//case "staircase": // TODO
	default:
		panic("Unimplemented " + rp.lineMode)
	}

	da.StrokeLines(rp.LineStyle, lines...)
}

func (rp *ResponsePlotter) Thumbnail(da *plot.DrawArea) {
	l := plotter.Line{LineStyle: rp.LineStyle}
	l.Thumbnail(da)
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

func (rp *ResponsePlotter) maybeConsolidateData(numberOfPixels int) {
	// idealy numberOfPixels should be size in pixels of char ares,
	// not char areay with Y axis and its label

	numberOfDataPoints := len(rp.Response.Values)
	pointsPerPixel := int(math.Ceil(float64(numberOfDataPoints) / float64(numberOfPixels)))

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
