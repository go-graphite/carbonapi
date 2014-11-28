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

	// set grid
	grid := plotter.NewGrid()
	grid.Vertical.Color = fgcolor
	grid.Horizontal.Color = fgcolor
	p.Add(grid)

	// need different timeMarker's based on step size
	p.Title.Text = r.FormValue("title")
	p.X.Tick.Marker = makeTimeMarker(*results[0].StepTime)


	var lines []plot.Plotter
	for i, r := range results {

		l := NewResponsePlotter(r)
		l.Color = plotutil.Color(i)

		if r.color != "" {
			l.Color = string2Color(r.color)
		} else {
			l.Color = plotutil.Color(i)
		}

		lines = append(lines, l)
	}
	p.Add(lines...)

	height := getInt(r.FormValue("height"), 250)
	width := getInt(r.FormValue("width"), 330)

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

func getString(s string, def string) string{
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

func string2Color(clr string) color.Color {
	switch clr {
	case "black":
		return hexToColor("#000000")
	case "navy":
		return hexToColor("#000080")
	case "darkblue":
		return hexToColor("#00008b")
	case "mediumblue":
		return hexToColor("#0000cd")
	case "blue":
		return hexToColor("#0000ff")
	case "darkgreen":
		return hexToColor("#006400")
	case "green":
		return hexToColor("#008000")
	case "teal":
		return hexToColor("#008080")
	case "darkcyan":
		return hexToColor("#008b8b")
	case "deepskyblue":
		return hexToColor("#00bfff")
	case "darkturquoise":
		return hexToColor("#00ced1")
	case "mediumspringgreen":
		return hexToColor("#00fa9a")
	case "lime":
		return hexToColor("#00ff00")
	case "springgreen":
		return hexToColor("#00ff7f")
	case "aqua":
		return hexToColor("#00ffff")
	case "cyan":
		return hexToColor("#00ffff")
	case "midnightblue":
		return hexToColor("#191970")
	case "dodgerblue":
		return hexToColor("#1e90ff")
	case "lightseagreen":
		return hexToColor("#20b2aa")
	case "forestgreen":
		return hexToColor("#228b22")
	case "seagreen":
		return hexToColor("#2e8b57")
	case "darkslategray":
		return hexToColor("#2f4f4f")
	case "limegreen":
		return hexToColor("#32cd32")
	case "mediumseagreen":
		return hexToColor("#3cb371")
	case "turquoise":
		return hexToColor("#40e0d0")
	case "royalblue":
		return hexToColor("#4169e1")
	case "steelblue":
		return hexToColor("#4682b4")
	case "darkslateblue":
		return hexToColor("#483d8b")
	case "mediumturquoise":
		return hexToColor("#48d1cc")
	case "indigo":
		return hexToColor("#4b0082")
	case "darkolivegreen":
		return hexToColor("#556b2f")
	case "cadetblue":
		return hexToColor("#5f9ea0")
	case "cornflowerblue":
		return hexToColor("#6495ed")
	case "mediumaquamarine":
		return hexToColor("#66cdaa")
	case "dimgray":
		return hexToColor("#696969")
	case "slateblue":
		return hexToColor("#6a5acd")
	case "olivedrab":
		return hexToColor("#6b8e23")
	case "slategray":
		return hexToColor("#708090")
	case "lightslategray":
		return hexToColor("#778899")
	case "mediumslateblue":
		return hexToColor("#7b68ee")
	case "lawngreen":
		return hexToColor("#7cfc00")
	case "chartreuse":
		return hexToColor("#7fff00")
	case "aquamarine":
		return hexToColor("#7fffd4")
	case "lavender":
		return hexToColor("#e6e6fa")
	case "darksalmon":
		return hexToColor("#e9967a")
	case "violet":
		return hexToColor("#ee82ee")
	case "palegoldenrod":
		return hexToColor("#eee8aa")
	case "lightcoral":
		return hexToColor("#f08080")
	case "khaki":
		return hexToColor("#f0e68c")
	case "aliceblue":
		return hexToColor("#f0f8ff")
	case "honeydew":
		return hexToColor("#f0fff0")
	case "azure":
		return hexToColor("#f0ffff")
	case "sandybrown":
		return hexToColor("#f4a460")
	case "wheat":
		return hexToColor("#f5deb3")
	case "beige":
		return hexToColor("#f5f5dc")
	case "whitesmoke":
		return hexToColor("#f5f5f5")
	case "mintcream":
		return hexToColor("#f5fffa")
	case "ghostwhite":
		return hexToColor("#f8f8ff")
	case "salmon":
		return hexToColor("#fa8072")
	case "antiquewhite":
		return hexToColor("#faebd7")
	case "linen":
		return hexToColor("#faf0e6")
	case "lightgoldenrodyellow":
		return hexToColor("#fafad2")
	case "oldlace":
		return hexToColor("#fdf5e6")
	case "red":
		return hexToColor("#ff0000")
	case "fuchsia":
		return hexToColor("#ff00ff")
	case "magenta":
		return hexToColor("#ff00ff")
	case "deeppink":
		return hexToColor("#ff1493")
	case "orangered":
		return hexToColor("#ff4500")
	case "tomato":
		return hexToColor("#ff6347")
	case "hotpink":
		return hexToColor("#ff69b4")
	case "coral":
		return hexToColor("#ff7f50")
	case "darkorange":
		return hexToColor("#ff8c00")
	case "lightsalmon":
		return hexToColor("#ffa07a")
	case "orange":
		return hexToColor("#ffa500")
	case "lightpink":
		return hexToColor("#ffb6c1")
	case "pink":
		return hexToColor("#ffc0cb")
	case "gold":
		return hexToColor("#ffd700")
	case "peachpuff":
		return hexToColor("#ffdab9")
	case "navajowhite":
		return hexToColor("#ffdead")
	case "moccasin":
		return hexToColor("#ffe4b5")
	case "bisque":
		return hexToColor("#ffe4c4")
	case "mistyrose":
		return hexToColor("#ffe4e1")
	case "blanchedalmond":
		return hexToColor("#ffebcd")
	case "papayawhip":
		return hexToColor("#ffefd5")
	case "lavenderblush":
		return hexToColor("#fff0f5")
	case "seashell":
		return hexToColor("#fff5ee")
	case "cornsilk":
		return hexToColor("#fff8dc")
	case "lemonchiffon":
		return hexToColor("#fffacd")
	case "floralwhite":
		return hexToColor("#fffaf0")
	case "snow":
		return hexToColor("#fffafa")
	case "yellow":
		return hexToColor("#ffff00")
	case "lightyellow":
		return hexToColor("#ffffe0")
	case "ivory":
		return hexToColor("#fffff0")
	case "white":
		return hexToColor("#ffffff")
	default:
		return hexToColor(clr)
	}
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
}

func NewResponsePlotter(r *metricData) *ResponsePlotter {
	return &ResponsePlotter{
		Response:  r,
		LineStyle: plotter.DefaultLineStyle,
	}
}

// Plot draws the Line, implementing the plot.Plotter
// interface.
func (rp *ResponsePlotter) Plot(da plot.DrawArea, plt *plot.Plot) {
	trX, trY := plt.Transforms(&da)

	start := float64(*rp.Response.StartTime)
	step := float64(*rp.Response.StepTime)
	absent := rp.Response.IsAbsent

	lines := make([][]plot.Point, 1, 1)

	lines[0] = make([]plot.Point, 0, len(rp.Response.Values))
	currentLine := 0

	lastAbsent := false
	for i, v := range rp.Response.Values {
		if absent[i] {
			lastAbsent = true
		} else if lastAbsent {
			currentLine++
			lines = append(lines, make([]plot.Point, 1, len(rp.Response.Values)))
			lines[currentLine][0] = plot.Point{X: trX(start + float64(i)*step), Y: trY(v)}
			lastAbsent = false
		} else {
			lines[currentLine] = append(lines[currentLine], plot.Point{X: trX(start + float64(i)*step), Y: trY(v)})
		}
	}

	da.StrokeLines(rp.LineStyle, lines...)
}

func (rp *ResponsePlotter) DataRange() (xmin, xmax, ymin, ymax float64) {
	ymin = math.Inf(1)
	ymax = math.Inf(-1)
	absent := rp.Response.IsAbsent

	for i, v := range rp.Response.Values {
		if absent[i] {
			continue
		}
		ymin = math.Min(ymin, v)
		ymax = math.Max(ymax, v)
	}
	xmin = float64(*rp.Response.StartTime)
	xmax = float64(*rp.Response.StopTime)
	return
}
