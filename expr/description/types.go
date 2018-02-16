package description

// FunctionType specifies list of available types
type FunctionType int

const (
	AggFunc FunctionType = iota
	Boolean
	Date
	Float
	IntOrInterval
	Integer
	Interval
	Node
	NodeOrTag
	SeriesList
	SeriesLists
	String
	Tag
)

// FunctionParam contains list of all available parameters of function
type FunctionParam struct {
	Name string `json:"name"`
	Multiple bool `json:"multiple,omitempty"`
	Required bool `json:"required,omitempty"`
	Type FunctionType `json:"type,omitempty"`
	Options []string `json:"options,omitempty"`
}

// FunctionDescription contains full function description.
type FunctionDescription struct {
	Description string `json:"description"`
	Function string `json:"function"`
	Group string `json:"group"`
	Module string `json:"module"`
	Name string `json:"name"`
	Params []FunctionParam `json:"params,omitempty"`
}