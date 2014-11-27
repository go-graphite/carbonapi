package main

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"time"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"code.google.com/p/plotinum/vg/vgimg"
)

var linesColors = `blue,green,red,purple,brown,yellow,aqua,grey,magenta,pink,gold,rose`

func marshalPNG(results []*pb.FetchResponse) []byte {

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	// need different timeMarker's based on step size
	p.X.Tick.Marker = timeMarker

	var lines []interface{}
	for _, r := range results {
		log.Println(r.GetName(), r.GetValues())
		lines = append(lines, r.GetName(), resultXYs(r))
	}
	err = plotutil.AddLines(p, lines...)

	dpi := 100

	// Draw the plot to an in-memory image.
	img := image.NewRGBA(image.Rect(0, 0, 8*dpi, 6*dpi))
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
