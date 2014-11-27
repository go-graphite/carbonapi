package main

import (
	"bytes"
	"image"
	"image/png"
	"math"
	"net/http"
	"strconv"
	"time"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"code.google.com/p/plotinum/vg/vgimg"
)

var linesColors = `blue,green,red,purple,brown,yellow,aqua,grey,magenta,pink,gold,rose`

func marshalPNG(r *http.Request, results []*pb.FetchResponse) []byte {

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	// need different timeMarker's based on step size
	p.Title.Text = r.FormValue("title")
	p.X.Tick.Marker = timeMarker

	p.Add(plotter.NewGrid())

	var lines []plot.Plotter
	for i, r := range results {

		l := NewResponsePlotter(r)
		l.Color = plotutil.Color(i)

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

func timeMarker(min, max float64) []plot.Tick {
	ticks := plot.DefaultTicks(min, max)

	for i, t := range ticks {
		if !t.IsMinor() {
			t0 := time.Unix(int64(t.Value), 0)
			ticks[i].Label = t0.Format("15:04:05")
		}
	}

	return ticks
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

type ResponsePlotter struct {
	Response *pb.FetchResponse
	plot.LineStyle
}

func NewResponsePlotter(r *pb.FetchResponse) *ResponsePlotter {
	return &ResponsePlotter{
		Response:  r,
		LineStyle: plotter.DefaultLineStyle,
	}
}

// Plot draws the Line, implementing the plot.Plotter
// interface.
func (rp *ResponsePlotter) Plot(da plot.DrawArea, plt *plot.Plot) {
	trX, trY := plt.Transforms(&da)

	start := float64(rp.Response.GetStartTime())
	step := float64(rp.Response.GetStepTime())
	absent := rp.Response.GetIsAbsent()

	lines := make([][]plot.Point, 1, 1)

	lines[0] = make([]plot.Point, 0, len(rp.Response.GetValues()))
	currentLine := 0

	lastAbsent := false
	for i, v := range rp.Response.GetValues() {
		if absent[i] {
			lastAbsent = true
		} else if lastAbsent {
			currentLine++
			lines = append(lines, make([]plot.Point, 1, len(rp.Response.GetValues())))
			lines[currentLine][0] = plot.Point{trX(start + float64(i)*step), trY(v)}
			lastAbsent = false
		} else {
			lines[currentLine] = append(lines[currentLine], plot.Point{trX(start + float64(i)*step), trY(v)})
		}
	}

	da.StrokeLines(rp.LineStyle, lines...)
}

func (rp *ResponsePlotter) DataRange() (xmin, xmax, ymin, ymax float64) {
	ymin = math.Inf(1)
	ymax = math.Inf(-1)
	absent := rp.Response.GetIsAbsent()

	for i, v := range rp.Response.GetValues() {
		if absent[i] {
			continue
		}
		ymin = math.Min(ymin, v)
		ymax = math.Max(ymax, v)
	}
	xmin = float64(rp.Response.GetStartTime())
	xmax = float64(rp.Response.GetStopTime())
	return
}
