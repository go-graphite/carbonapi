package expr

import (
	"image/color"
	"net/http"
	"time"

	"math"
	"strconv"
	"strings"
)

var defaultColorList = []string{"blue", "green", "red", "purple", "brown", "yellow", "aqua", "grey", "magenta", "pink", "gold", "rose"}

type YAxisSide int

const (
	YAxisSideRight YAxisSide = 1
	YAxisSideLeft            = 2
)

func getAxisSide(s string, def YAxisSide) YAxisSide {
	if s == "" {
		return def
	}
	if s == "right" {
		return YAxisSideRight
	}
	return YAxisSideLeft
}

type LineMode int

const (
	LineModeSlope     LineMode = 1
	LineModeStaircase          = 2
	LineModeConnected          = 4
)

type AreaMode int

const (
	AreaModeNone    AreaMode = 1
	AreaModeFirst            = 2
	AreaModeAll              = 4
	AreaModeStacked          = 8
)

func getAreaMode(s string, def AreaMode) AreaMode {
	if s == "" {
		return def
	}
	switch s {
	case "first":
		return AreaModeFirst
	case "all":
		return AreaModeAll
	case "stacked":
		return AreaModeStacked
	}
	return AreaModeNone
}

type PieMode int

const (
	PieModeMaximum PieMode = 1
	PieModeMinimum         = 2
	PieModeAverage         = 4
)

func getPieMode(s string, def PieMode) PieMode {
	if s == "" {
		return def
	}
	if s == "maximum" {
		return PieModeMaximum
	}
	if s == "minimum" {
		return PieModeMinimum
	}
	return PieModeAverage
}

func getLineMode(s string, def LineMode) LineMode {
	if s == "" {
		return def
	}
	if s == "slope" {
		return LineModeSlope
	}
	if s == "staircase" {
		return LineModeStaircase
	}
	return LineModeConnected
}

type FontWeight int

const (
	FontWeightNormal FontWeight = iota
	FontWeightBold
)

func getFontWeight(s string) FontWeight {
	if TruthyBool(s) {
		return FontWeightBold
	}
	return FontWeightNormal
}

type FontSlant int

const (
	FontSlantNormal FontSlant = iota
	FontSlantItalic
	FontSlantOblique
)

func getFontItalic(s string) FontSlant {
	if TruthyBool(s) {
		return FontSlantItalic
	}
	return FontSlantNormal
}

type PictureParams struct {
	Width      float64
	Height     float64
	Margin     int
	LogBase    float64
	FgColor    color.RGBA
	BgColor    color.RGBA
	MajorLine  color.RGBA
	MinorLine  color.RGBA
	FontName   string
	FontSize   float64
	FontBold   FontWeight
	FontItalic FontSlant

	GraphOnly  bool
	HideLegend bool
	HideGrid   bool
	HideAxes   bool
	HideYAxis  bool
	HideXAxis  bool
	YAxisSide  YAxisSide

	Title       string
	Vtitle      string
	VtitleRight string

	Tz *time.Location

	ConnectedLimit int
	LineMode       LineMode
	AreaMode       AreaMode
	AreaAlpha      float64
	PieMode        PieMode
	LineWidth      float64
	ColorList      []string

	YMin    float64
	YMax    float64
	XMin    float64
	XMax    float64
	YStep   float64
	XStep   float64
	MinorY  int
	XFormat string

	YMaxLeft    float64
	YLimitLeft  float64
	YMaxRight   float64
	YLimitRight float64
	YMinLeft    float64
	YMinRight   float64
	YStepL      float64
	YStepR      float64

	UniqueLegend   bool
	DrawNullAsZero bool
	DrawAsInfinite bool

	YUnitSystem string
	YDivisors   []float64

	RightWidth  float64
	RightDashed bool
	RightColor  string
	LeftWidth   float64
	LeftDashed  bool
	LeftColor   string

	MinorGridLineColor string
	MajorGridLineColor string
}

func getPictureParams(r *http.Request, metricData []*MetricData) PictureParams {
	return PictureParams{
		Width:      getFloat64(r.FormValue("width"), 330),
		Height:     getFloat64(r.FormValue("height"), 250),
		Margin:     getInt(r.FormValue("margin"), 10),
		LogBase:    getLogBase(r.FormValue("logBase")),
		FgColor:    string2RGBA(getString(r.FormValue("fgcolor"), "white")),
		BgColor:    string2RGBA(getString(r.FormValue("bgcolor"), "black")),
		MajorLine:  string2RGBA(getString(r.FormValue("majorLine"), "rose")),
		MinorLine:  string2RGBA(getString(r.FormValue("minorLine"), "grey")),
		FontName:   getString(r.FormValue("fontName"), "Sans"),
		FontSize:   getFloat64(r.FormValue("fontSize"), 10.0),
		FontBold:   getFontWeight(r.FormValue("fontBold")),
		FontItalic: getFontItalic(r.FormValue("fontItalic")),

		GraphOnly:  getBool(r.FormValue("graphOnly"), false),
		HideLegend: getBool(r.FormValue("hideLegend"), len(metricData) > 10),
		HideGrid:   getBool(r.FormValue("hideGrid"), false),
		HideAxes:   getBool(r.FormValue("hideAxes"), false),
		HideYAxis:  getBool(r.FormValue("hideYAxis"), false),
		HideXAxis:  getBool(r.FormValue("hideXAxis"), false),
		YAxisSide:  getAxisSide(r.FormValue("yAxisSide"), YAxisSideLeft),

		Title:       getString(r.FormValue("title"), ""),
		Vtitle:      getString(r.FormValue("vtitle"), ""),
		VtitleRight: getString(r.FormValue("vtitleRight"), ""),

		Tz: getTimeZone(r.FormValue("tz"), time.Local),

		ConnectedLimit: getInt(r.FormValue("connectedLimit"), math.MaxUint32),
		LineMode:       getLineMode(r.FormValue("lineMode"), LineModeSlope),
		AreaMode:       getAreaMode(r.FormValue("areaMode"), AreaModeNone),
		AreaAlpha:      getFloat64(r.FormValue("areaAlpha"), math.NaN()),
		PieMode:        getPieMode(r.FormValue("pieMode"), PieModeAverage),
		LineWidth:      getFloat64(r.FormValue("lineWidth"), 1.2),
		ColorList:      getStringArray(r.FormValue("colorList"), defaultColorList),

		YMin:    getFloat64(r.FormValue("yMin"), math.NaN()),
		YMax:    getFloat64(r.FormValue("yMax"), math.NaN()),
		YStep:   getFloat64(r.FormValue("yStep"), math.NaN()),
		XMin:    getFloat64(r.FormValue("xMin"), math.NaN()),
		XMax:    getFloat64(r.FormValue("xMax"), math.NaN()),
		XStep:   getFloat64(r.FormValue("xStep"), math.NaN()),
		XFormat: getString(r.FormValue("xFormat"), ""),
		MinorY:  getInt(r.FormValue("minorY"), 1),

		UniqueLegend:   getBool(r.FormValue("uniqueLegend"), false),
		DrawNullAsZero: getBool(r.FormValue("drawNullAsZero"), false),
		DrawAsInfinite: getBool(r.FormValue("drawAsInfinite"), false),

		YMinLeft:    getFloat64(r.FormValue("yMinLeft"), math.NaN()),
		YMinRight:   getFloat64(r.FormValue("yMinRight"), math.NaN()),
		YMaxLeft:    getFloat64(r.FormValue("yMaxLeft"), math.NaN()),
		YMaxRight:   getFloat64(r.FormValue("yMaxRight"), math.NaN()),
		YStepL:      getFloat64(r.FormValue("yStepLeft"), math.NaN()),
		YStepR:      getFloat64(r.FormValue("yStepRight"), math.NaN()),
		YLimitLeft:  getFloat64(r.FormValue("yLimitLeft"), math.NaN()),
		YLimitRight: getFloat64(r.FormValue("yLimitRight"), math.NaN()),

		YUnitSystem: getString(r.FormValue("yUnitSystem"), "si"),
		YDivisors:   getFloatArray(r.FormValue("yDivisors"), []float64{4, 5, 6}),

		RightWidth:  getFloat64(r.FormValue("rightWidth"), 1.2),
		RightDashed: getBool(r.FormValue("rightDashed"), false),
		RightColor:  getString(r.FormValue("rightColor"), ""),
		LeftWidth:   getFloat64(r.FormValue("leftWidth"), 1.2),
		LeftDashed:  getBool(r.FormValue("leftDashed"), false),
		LeftColor:   getString(r.FormValue("leftColor"), ""),

		MajorGridLineColor: getString(r.FormValue("majorGridLineColor"), "white"),
		MinorGridLineColor: getString(r.FormValue("minorGridLineColor"), "grey"),
	}
}

func getStringArray(s string, def []string) []string {
	if s == "" {
		return def
	}

	ss := strings.Split(s, ",")
	var strs []string
	for _, v := range ss {
		strs = append(strs, strings.TrimSpace(v))
	}

	return strs
}

func getFloatArray(s string, def []float64) []float64 {
	if s == "" {
		return def
	}
	ss := strings.Split(s, ",")
	var fs []float64
	for _, v := range ss {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return def
		}
		fs = append(fs, f)
	}
	return fs
}

func getLogBase(s string) float64 {
	if s == "e" {
		return math.E
	}
	b, err := strconv.ParseFloat(s, 64)
	if err != nil || b < 1 {
		return 0
	}
	return b
}

func getTimeZone(s string, def *time.Location) *time.Location {
	if s == "" {
		return def
	}
	tz, err := time.LoadLocation(s)
	if err != nil {
		return def
	}
	return tz
}

func string2RGBA(clr string) color.RGBA {
	if c, ok := colors[clr]; ok {
		return c
	}
	return hexToRGBA(clr)
}

// https://code.google.com/p/sadbox/source/browse/color/hex.go
// hexToColor converts an Hex string to a RGB triple.
func hexToRGBA(h string) color.RGBA {
	var r, g, b uint8
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}

	if len(h) == 3 {
		h = h[:1] + h[:1] + h[1:2] + h[1:2] + h[2:] + h[2:]
	}

	alpha := byte(255)

	if len(h) == 6 {
		if rgb, err := strconv.ParseUint(string(h), 16, 32); err == nil {
			r = uint8(rgb >> 16)
			g = uint8(rgb >> 8)
			b = uint8(rgb)
		}
	}

	if len(h) == 8 {
		if rgb, err := strconv.ParseUint(string(h), 16, 32); err == nil {
			r = uint8(rgb >> 24)
			g = uint8(rgb >> 16)
			b = uint8(rgb >> 8)
			alpha = uint8(rgb)
		}
	}

	return color.RGBA{r, g, b, alpha}
}
