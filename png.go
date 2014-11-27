package main

import (
	"bytes"
	"image"
	"image/png"
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
	p.Title.Text = results[0].GetName()
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

	height := getInt(r.FormValue("height"), 300)
	width := getInt(r.FormValue("width"), 400)

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

func resultXYs(r *pb.FetchResponse) plotter.XYs {
	pts := make(plotter.XYs, len(r.GetValues()))
	start := float64(r.GetStartTime())
	step := float64(r.GetStepTime())
	for i, v := range r.GetValues() {
		pts[i].X = start + float64(i)*step
		pts[i].Y = v
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
