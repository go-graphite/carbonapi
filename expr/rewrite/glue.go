package rewrite

import (
	"sort"
	"strings"

	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/metadata"
	"github.com/grafana/carbonapi/expr/rewrite/aboveSeries"
	"github.com/grafana/carbonapi/expr/rewrite/applyByNode"
)

type initFunc struct {
	name     string
	filename string
	order    interfaces.Order
	f        func(configFile string) []interfaces.RewriteFunctionMetadata
}

func New(configs map[string]string) {
	funcs := []initFunc{
		{name: "aboveSeries", filename: "aboveSeries", order: aboveSeries.GetOrder(), f: aboveSeries.New},
		{name: "applyByNode", filename: "applyByNode", order: applyByNode.GetOrder(), f: applyByNode.New},
	}

	sort.Slice(funcs, func(i, j int) bool {
		if funcs[i].order == interfaces.Any && funcs[j].order == interfaces.Last {
			return true
		}
		if funcs[i].order == interfaces.Last && funcs[j].order == interfaces.Any {
			return false
		}
		return funcs[i].name > funcs[j].name
	})

	for _, f := range funcs {
		md := f.f(configs[strings.ToLower(f.name)])
		for _, m := range md {
			metadata.RegisterRewriteFunctionWithFilename(m.Name, f.filename, m.F)
		}
	}
}
