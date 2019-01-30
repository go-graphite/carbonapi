package msgpack

//go:generate msgp

type GraphiteFetchResponse struct {
	Start          uint32        `msg:"start"`
	End            uint32        `msg:"end"`
	Step           uint32        `msg:"step"`
	Name           string        `msg:"name"`
	PathExpression string        `msg:"pathExpression"`
	Values         []interface{} `msg:"values"`
}

type MultiGraphiteFetchResponse []GraphiteFetchResponse

type GraphiteGlobResponse struct {
	IsLeaf bool   `msg:"isLeaf"`
	Path   string `msg:"path"`
}

type MultiGraphiteGlobResponse []GraphiteGlobResponse
