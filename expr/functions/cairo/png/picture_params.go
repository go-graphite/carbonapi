package png

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

var DefaultColorList = []string{"blue", "green", "red", "purple", "brown", "yellow", "aqua", "grey", "magenta", "pink", "gold", "rose"}

type YAxisSide int

const (
	YAxisSideRight YAxisSide = 1 << iota
	YAxisSideLeft
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
	LineModeSlope LineMode = 1 << iota
	LineModeStaircase
	LineModeConnected
)

type AreaMode int

const (
	AreaModeNone AreaMode = 1 << iota
	AreaModeFirst
	AreaModeAll
	AreaModeStacked
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
	PieModeMaximum PieMode = 1 << iota
	PieModeMinimum
	PieModeAverage
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

func getFontWeight(s string, def FontWeight) FontWeight {
	if s == "" {
		return def
	}
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

func getFontItalic(s string, def FontSlant) FontSlant {
	if s == "" {
		return def
	}
	if parser.TruthyBool(s) {
		return FontSlantItalic
	}
	return FontSlantNormal
}

type PictureParams struct {
	PixelRatio float64
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

// GetPictureParams returns PictureParams with default settings
func GetPictureParams(r *http.Request, metricData []*types.MetricData) PictureParams {
	return GetPictureParamsWithTemplate(r, "default", metricData)
}

// GetPictureParamsWithTemplate returns PictureParams with specified template
func GetPictureParamsWithTemplate(r *http.Request, template string, metricData []*types.MetricData) PictureParams {
	t, ok := templates[template]
	if !ok {
		t = templates["default"]
	}
	return PictureParams{
		PixelRatio: getFloat64(r.FormValue("pixelRatio"), 1.0),
		Width:      getFloat64(r.FormValue("width"), t.Width),
		Height:     getFloat64(r.FormValue("height"), t.Height),
		Margin:     getInt(r.FormValue("margin"), t.Margin),
		LogBase:    getLogBase(r.FormValue("logBase")),
		FgColor:    getString(r.FormValue("fgcolor"), t.FgColor),
		BgColor:    getString(r.FormValue("bgcolor"), t.BgColor),
		MajorLine:  getString(r.FormValue("majorLine"), t.MajorLine),
		MinorLine:  getString(r.FormValue("minorLine"), t.MinorLine),
		FontName:   getString(r.FormValue("fontName"), t.FontName),
		FontSize:   getFloat64(r.FormValue("fontSize"), t.FontSize),
		FontBold:   getFontWeight(r.FormValue("fontBold"), t.FontBold),
		FontItalic: getFontItalic(r.FormValue("fontItalic"), t.FontItalic),

		GraphOnly:  getBool(r.FormValue("graphOnly"), t.GraphOnly),
		HideLegend: getBool(r.FormValue("hideLegend"), len(metricData) > 10),
		HideGrid:   getBool(r.FormValue("hideGrid"), t.HideGrid),
		HideAxes:   getBool(r.FormValue("hideAxes"), t.HideAxes),
		HideYAxis:  getBool(r.FormValue("hideYAxis"), t.HideYAxis),
		HideXAxis:  getBool(r.FormValue("hideXAxis"), t.HideXAxis),
		YAxisSide:  getAxisSide(r.FormValue("yAxisSide"), t.YAxisSide),

		Title:       getString(r.FormValue("title"), t.Title),
		Vtitle:      getString(r.FormValue("vtitle"), t.Vtitle),
		VtitleRight: getString(r.FormValue("vtitleRight"), t.VtitleRight),

		Tz: getTimeZone(r.FormValue("tz"), t.Tz),

		ConnectedLimit: getInt(r.FormValue("connectedLimit"), t.ConnectedLimit),
		LineMode:       getLineMode(r.FormValue("lineMode"), t.LineMode),
		AreaMode:       getAreaMode(r.FormValue("areaMode"), t.AreaMode),
		AreaAlpha:      getFloat64(r.FormValue("areaAlpha"), t.AreaAlpha),
		PieMode:        getPieMode(r.FormValue("pieMode"), t.PieMode),
		LineWidth:      getFloat64(r.FormValue("lineWidth"), t.LineWidth),
		ColorList:      getStringArray(r.FormValue("colorList"), t.ColorList),

		YMin:    getFloat64(r.FormValue("yMin"), t.YMin),
		YMax:    getFloat64(r.FormValue("yMax"), t.YMax),
		YStep:   getFloat64(r.FormValue("yStep"), t.YStep),
		XMin:    getFloat64(r.FormValue("xMin"), t.XMin),
		XMax:    getFloat64(r.FormValue("xMax"), t.XMax),
		XStep:   getFloat64(r.FormValue("xStep"), t.XStep),
		XFormat: getString(r.FormValue("xFormat"), t.XFormat),
		MinorY:  getInt(r.FormValue("minorY"), t.MinorY),

		UniqueLegend:   getBool(r.FormValue("uniqueLegend"), t.UniqueLegend),
		DrawNullAsZero: getBool(r.FormValue("drawNullAsZero"), t.DrawNullAsZero),
		DrawAsInfinite: getBool(r.FormValue("drawAsInfinite"), t.DrawAsInfinite),

		YMinLeft:    getFloat64(r.FormValue("yMinLeft"), t.YMinLeft),
		YMinRight:   getFloat64(r.FormValue("yMinRight"), t.YMinRight),
		YMaxLeft:    getFloat64(r.FormValue("yMaxLeft"), t.YMaxLeft),
		YMaxRight:   getFloat64(r.FormValue("yMaxRight"), t.YMaxRight),
		YStepL:      getFloat64(r.FormValue("yStepLeft"), t.YStepL),
		YStepR:      getFloat64(r.FormValue("yStepRight"), t.YStepR),
		YLimitLeft:  getFloat64(r.FormValue("yLimitLeft"), t.YLimitLeft),
		YLimitRight: getFloat64(r.FormValue("yLimitRight"), t.YLimitRight),

		YUnitSystem: getString(r.FormValue("yUnitSystem"), t.YUnitSystem),
		YDivisors:   getFloatArray(r.FormValue("yDivisors"), t.YDivisors),

		RightWidth:  getFloat64(r.FormValue("rightWidth"), t.RightWidth),
		RightDashed: getBool(r.FormValue("rightDashed"), t.RightDashed),
		RightColor:  getString(r.FormValue("rightColor"), t.RightColor),
		LeftWidth:   getFloat64(r.FormValue("leftWidth"), t.LeftWidth),
		LeftDashed:  getBool(r.FormValue("leftDashed"), t.LeftDashed),
		LeftColor:   getString(r.FormValue("leftColor"), t.LeftColor),

		MajorGridLineColor: getString(r.FormValue("majorGridLineColor"), t.MajorGridLineColor),
		MinorGridLineColor: getString(r.FormValue("minorGridLineColor"), t.MinorGridLineColor),
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

// SetTemplate adds a picture param template with specified name and parameters
func SetTemplate(name string, params PictureParams) {
	templates[name] = params
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

var templates = map[string]PictureParams{
	"default": {
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
	},
}
