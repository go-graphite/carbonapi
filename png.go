package main

import (
	"bytes"
	"image"
	"image/png"
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

	// need different timeMarker's based on step size
	p.Title.Text = r.FormValue("title")
	p.X.Tick.Marker = timeMarker

	p.Add(plotter.NewGrid())

	var lines []plot.Plotter
	for i, r := range results {

		t := resultXYs(r)

		l, _ := plotter.NewLine(t)
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

type xy struct {
	X, Y float64
}

func resultXYs(r *metricData) plotter.XYs {
	pts := make(plotter.XYs, 0, len(r.Values))
	start := float64(*r.StartTime)
	step := float64(*r.StepTime)
	absent := r.IsAbsent
	for i, v := range r.Values {
		if absent[i] {
			continue
		}
		pts = append(pts, xy{start + float64(i)*step, v})
	}
	return pts
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
