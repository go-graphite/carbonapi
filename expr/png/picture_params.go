package png

import (
	"image/color"
	"net/http"
	"time"

	"math"
	"strconv"
	"strings"

	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

var DefaultColorList = []string{"blue", "green", "red", "purple", "brown", "yellow", "aqua", "grey", "magenta", "pink", "gold", "rose"}

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
	if parser.TruthyBool(s) {
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
	if parser.TruthyBool(s) {
		return FontSlantItalic
	}
	return FontSlantNormal
}

type PictureParams struct {
	Width      float64
	Height     float64
	Margin     int
	LogBase    float64
	FgColor    string
	BgColor    string
	MajorLine  string
	MinorLine  string
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

func GetPictureParams(r *http.Request, metricData []*types.MetricData) PictureParams {
	return PictureParams{
		Width:      getFloat64(r.FormValue("width"), DefaultParams.Width),
		Height:     getFloat64(r.FormValue("height"), DefaultParams.Height),
		Margin:     getInt(r.FormValue("margin"), DefaultParams.Margin),
		LogBase:    getLogBase(r.FormValue("logBase")),
		FgColor:    getString(r.FormValue("fgcolor"), DefaultParams.FgColor),
		BgColor:    getString(r.FormValue("bgcolor"), DefaultParams.BgColor),
		MajorLine:  getString(r.FormValue("majorLine"), DefaultParams.MajorLine),
		MinorLine:  getString(r.FormValue("minorLine"), DefaultParams.MinorLine),
		FontName:   getString(r.FormValue("fontName"), DefaultParams.FontName),
		FontSize:   getFloat64(r.FormValue("fontSize"), DefaultParams.FontSize),
		FontBold:   getFontWeight(r.FormValue("fontBold")),
		FontItalic: getFontItalic(r.FormValue("fontItalic")),

		GraphOnly:  getBool(r.FormValue("graphOnly"), DefaultParams.GraphOnly),
		HideLegend: getBool(r.FormValue("hideLegend"), len(metricData) > 10),
		HideGrid:   getBool(r.FormValue("hideGrid"), DefaultParams.HideGrid),
		HideAxes:   getBool(r.FormValue("hideAxes"), DefaultParams.HideAxes),
		HideYAxis:  getBool(r.FormValue("hideYAxis"), DefaultParams.HideYAxis),
		HideXAxis:  getBool(r.FormValue("hideXAxis"), DefaultParams.HideXAxis),
		YAxisSide:  getAxisSide(r.FormValue("yAxisSide"), DefaultParams.YAxisSide),

		Title:       getString(r.FormValue("title"), DefaultParams.Title),
		Vtitle:      getString(r.FormValue("vtitle"), DefaultParams.Vtitle),
		VtitleRight: getString(r.FormValue("vtitleRight"), DefaultParams.VtitleRight),

		Tz: getTimeZone(r.FormValue("tz"), DefaultParams.Tz),

		ConnectedLimit: getInt(r.FormValue("connectedLimit"), DefaultParams.ConnectedLimit),
		LineMode:       getLineMode(r.FormValue("lineMode"), DefaultParams.LineMode),
		AreaMode:       getAreaMode(r.FormValue("areaMode"), DefaultParams.AreaMode),
		AreaAlpha:      getFloat64(r.FormValue("areaAlpha"), DefaultParams.AreaAlpha),
		PieMode:        getPieMode(r.FormValue("pieMode"), DefaultParams.PieMode),
		LineWidth:      getFloat64(r.FormValue("lineWidth"), DefaultParams.LineWidth),
		ColorList:      getStringArray(r.FormValue("colorList"), DefaultParams.ColorList),

		YMin:    getFloat64(r.FormValue("yMin"), DefaultParams.YMin),
		YMax:    getFloat64(r.FormValue("yMax"), DefaultParams.YMax),
		YStep:   getFloat64(r.FormValue("yStep"), DefaultParams.YStep),
		XMin:    getFloat64(r.FormValue("xMin"), DefaultParams.XMin),
		XMax:    getFloat64(r.FormValue("xMax"), DefaultParams.XMax),
		XStep:   getFloat64(r.FormValue("xStep"), DefaultParams.XStep),
		XFormat: getString(r.FormValue("xFormat"), DefaultParams.XFormat),
		MinorY:  getInt(r.FormValue("minorY"), DefaultParams.MinorY),

		UniqueLegend:   getBool(r.FormValue("uniqueLegend"), DefaultParams.UniqueLegend),
		DrawNullAsZero: getBool(r.FormValue("drawNullAsZero"), DefaultParams.DrawNullAsZero),
		DrawAsInfinite: getBool(r.FormValue("drawAsInfinite"), DefaultParams.DrawAsInfinite),

		YMinLeft:    getFloat64(r.FormValue("yMinLeft"), DefaultParams.YMinLeft),
		YMinRight:   getFloat64(r.FormValue("yMinRight"), DefaultParams.YMinRight),
		YMaxLeft:    getFloat64(r.FormValue("yMaxLeft"), DefaultParams.YMaxLeft),
		YMaxRight:   getFloat64(r.FormValue("yMaxRight"), DefaultParams.YMaxRight),
		YStepL:      getFloat64(r.FormValue("yStepLeft"), DefaultParams.YStepL),
		YStepR:      getFloat64(r.FormValue("yStepRight"), DefaultParams.YStepR),
		YLimitLeft:  getFloat64(r.FormValue("yLimitLeft"), DefaultParams.YLimitLeft),
		YLimitRight: getFloat64(r.FormValue("yLimitRight"), DefaultParams.YLimitRight),

		YUnitSystem: getString(r.FormValue("yUnitSystem"), DefaultParams.YUnitSystem),
		YDivisors:   getFloatArray(r.FormValue("yDivisors"), DefaultParams.YDivisors),

		RightWidth:  getFloat64(r.FormValue("rightWidth"), DefaultParams.RightWidth),
		RightDashed: getBool(r.FormValue("rightDashed"), DefaultParams.RightDashed),
		RightColor:  getString(r.FormValue("rightColor"), DefaultParams.RightColor),
		LeftWidth:   getFloat64(r.FormValue("leftWidth"), DefaultParams.LeftWidth),
		LeftDashed:  getBool(r.FormValue("leftDashed"), DefaultParams.LeftDashed),
		LeftColor:   getString(r.FormValue("leftColor"), DefaultParams.LeftColor),

		MajorGridLineColor: getString(r.FormValue("majorGridLineColor"), DefaultParams.MajorGridLineColor),
		MinorGridLineColor: getString(r.FormValue("minorGridLineColor"), DefaultParams.MinorGridLineColor),
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

var DefaultParams = PictureParams{
	Width:      330,
	Height:     250,
	Margin:     10,
	LogBase:    0,
	FgColor:    "white",
	BgColor:    "black",
	MajorLine:  "rose",
	MinorLine:  "grey",
	FontName:   "Sans",
	FontSize:   10,
	FontBold:   FontWeightNormal,
	FontItalic: FontSlantNormal,

	GraphOnly:  false,
	HideLegend: false,
	HideGrid:   false,
	HideAxes:   false,
	HideYAxis:  false,
	HideXAxis:  false,
	YAxisSide:  YAxisSideLeft,

	Title:       "",
	Vtitle:      "",
	VtitleRight: "",

	Tz: time.Local,

	ConnectedLimit: math.MaxInt32,
	LineMode:       LineModeSlope,
	AreaMode:       AreaModeNone,
	AreaAlpha:      math.NaN(),
	PieMode:        PieModeAverage,
	LineWidth:      1.2,
	ColorList:      DefaultColorList,

	YMin:    math.NaN(),
	YMax:    math.NaN(),
	YStep:   math.NaN(),
	XMin:    math.NaN(),
	XMax:    math.NaN(),
	XStep:   math.NaN(),
	XFormat: "",
	MinorY:  1,

	UniqueLegend:   false,
	DrawNullAsZero: false,
	DrawAsInfinite: false,

	YMinLeft:    math.NaN(),
	YMinRight:   math.NaN(),
	YMaxLeft:    math.NaN(),
	YMaxRight:   math.NaN(),
	YStepL:      math.NaN(),
	YStepR:      math.NaN(),
	YLimitLeft:  math.NaN(),
	YLimitRight: math.NaN(),

	YUnitSystem: "si",
	YDivisors:   []float64{4, 5, 6},

	RightWidth:  1.2,
	RightDashed: false,
	RightColor:  "",
	LeftWidth:   1.2,
	LeftDashed:  false,
	LeftColor:   "",

	MajorGridLineColor: "white",
	MinorGridLineColor: "grey",
}
