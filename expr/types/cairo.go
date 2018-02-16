// +build cairo

package types

const DefaultStackName = "__DEFAULT__"

type GraphOptions struct {
	// extra options
	XStep     float64
	Color     string
	Alpha     float64
	LineWidth float64
	Invisible bool

	DrawAsInfinite bool
	SecondYAxis    bool
	Dashed         float64
	HasAlpha       bool
	HasLineWidth   bool
	Stacked        bool
	StackName      string
}
