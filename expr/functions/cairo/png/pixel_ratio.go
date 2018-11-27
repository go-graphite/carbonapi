// +build cairo

package png

import "github.com/evmar/gocairo/cairo"

// interface with all used cairo.Context methods
type cairoContext interface {
	Rectangle(x, y, width, height float64) // pixel ratio required
	GetLineWidth() float64                 // pixel ratio required
	LineTo(x, y float64)                   // pixel ratio required
	MoveTo(x, y float64)                   // pixel ratio required
	SetLineWidth(width float64)            // pixel ratio required
	SetFontSize(size float64)              // pixel ratio required
	SetFontOptions(options *cairo.FontOptions)
	Stroke()
	SetDash(dashes []float64, offset float64)            // pixel ratio required
	TextExtents(utf8 string, extents *cairo.TextExtents) // pixel ratio required
	FontExtents(extents *cairo.FontExtents)              // pixel ratio required
	Rotate(angle float64)
	SetLineCap(lineCap cairo.LineCap)
	SetLineJoin(lineJoin cairo.LineJoin)
	RelMoveTo(dx, dy float64) // pixel ratio required
	SetSourceRGBA(red, green, blue, alpha float64)
	SetMatrix(matrix *cairo.Matrix) // pixel ratio required
	GetMatrix(matrix *cairo.Matrix) // pixel ratio required
	Clip()
	Fill()
	ClosePath()
	SelectFontFace(family string, slant cairo.FontSlant, weight cairo.FontWeight) // pixel ratio required
	TextPath(utf8 string)
	Save()
	Restore()
	FillPreserve()
	AppendPath(path *cairo.Path)
	CopyPath() *cairo.Path
}

type pixelRatioContext struct {
	*cairo.Context
	pr float64 // pixel ratio
}

type cairoSurfaceContext struct {
	context cairoContext
}

func isDefaultRatio(pixelRatio float64) bool {
	if pixelRatio > 0.9999 && pixelRatio < 1.0001 {
		return true
	}
	return false
}

func svgSurfaceCreate(filename string, widthInPoints, heightInPoints float64, pixelRatio float64) *cairo.SVGSurface {
	if isDefaultRatio(pixelRatio) {
		return cairo.SVGSurfaceCreate(filename, widthInPoints, heightInPoints)
	}
	return cairo.SVGSurfaceCreate(filename, pixelRatio*widthInPoints, pixelRatio*heightInPoints)
}

func imageSurfaceCreate(format cairo.Format, width, height float64, pixelRatio float64) *cairo.ImageSurface {
	if isDefaultRatio(pixelRatio) {
		return cairo.ImageSurfaceCreate(format, int(width), int(height))
	}
	return cairo.ImageSurfaceCreate(format, int(pixelRatio*float64(width)), int(pixelRatio*float64(height)))
}

func createContext(surface *cairo.Surface, pixelRatio float64) *cairoSurfaceContext {
	if isDefaultRatio(pixelRatio) {
		return &cairoSurfaceContext{context: cairo.Create(surface)}
	}

	return &cairoSurfaceContext{
		context: &pixelRatioContext{
			Context: cairo.Create(surface),
			pr:      pixelRatio,
		},
	}
}

func (c *pixelRatioContext) Rectangle(x, y, width, height float64) {
	c.Context.Rectangle(c.pr*x, c.pr*y, c.pr*width, c.pr*height)
}

func (c *pixelRatioContext) GetLineWidth() float64 {
	return c.Context.GetLineWidth() / c.pr
}

func (c *pixelRatioContext) LineTo(x, y float64) {
	c.Context.LineTo(c.pr*x, c.pr*y)
}

func (c *pixelRatioContext) MoveTo(x, y float64) {
	c.Context.MoveTo(c.pr*x, c.pr*y)
}

func (c *pixelRatioContext) SetLineWidth(width float64) {
	c.Context.SetLineWidth(c.pr * width)
}

func (c *pixelRatioContext) SetFontSize(size float64) {
	c.Context.SetFontSize(c.pr * size)
}

func (c *pixelRatioContext) SetDash(dashes []float64, offset float64) {
	dr := make([]float64, len(dashes))
	for i := 0; i < len(dashes); i++ {
		dr[i] = dashes[i] * c.pr
	}
	c.Context.SetDash(dr, offset*c.pr)
}

func (c *pixelRatioContext) TextExtents(utf8 string, extents *cairo.TextExtents) {
	var e cairo.TextExtents
	c.Context.TextExtents(utf8, &e)
	extents.XBearing = e.XBearing / c.pr
	extents.YBearing = e.YBearing / c.pr
	extents.Width = e.Width / c.pr
	extents.Height = e.Height / c.pr
	extents.XAdvance = e.XAdvance / c.pr
	extents.YAdvance = e.YAdvance / c.pr
}

func (c *pixelRatioContext) FontExtents(extents *cairo.FontExtents) {
	var e cairo.FontExtents
	c.Context.FontExtents(&e)
	extents.Ascent = e.Ascent / c.pr
	extents.Descent = e.Descent / c.pr
	extents.Height = e.Height / c.pr
	extents.MaxXAdvance = e.MaxXAdvance / c.pr
	extents.MaxYAdvance = e.MaxYAdvance / c.pr
}

func (c *pixelRatioContext) RelMoveTo(dx, dy float64) {
	c.Context.RelMoveTo(c.pr*dx, c.pr*dy)
}

func (c *pixelRatioContext) SetMatrix(matrix *cairo.Matrix) {
	var m cairo.Matrix
	m.Xx = matrix.Xx * c.pr
	m.Yx = matrix.Yx * c.pr
	m.Xy = matrix.Xy * c.pr
	m.Yy = matrix.Yy * c.pr
	m.X0 = matrix.X0 * c.pr
	m.Y0 = matrix.Y0 * c.pr
	c.Context.SetMatrix(&m)
}

func (c *pixelRatioContext) GetMatrix(matrix *cairo.Matrix) {
	var m cairo.Matrix
	c.Context.GetMatrix(&m)
	matrix.Xx = m.Xx / c.pr
	matrix.Yx = m.Yx / c.pr
	matrix.Xy = m.Xy / c.pr
	matrix.Yy = m.Yy / c.pr
	matrix.X0 = m.X0 / c.pr
	matrix.Y0 = m.Y0 / c.pr
}

func (c *pixelRatioContext) SelectFontFace(family string, slant cairo.FontSlant, weight cairo.FontWeight) {
	c.Context.SelectFontFace(family, slant, cairo.FontWeight(c.pr*float64(weight)))
}
