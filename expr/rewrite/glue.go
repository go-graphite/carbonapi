package rewrite

import (
	"sort"
	"strings"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/rewrite/aboveSeries"
	"github.com/go-graphite/carbonapi/expr/rewrite/applyByNode"
)

type initFunc struct {
	name  string
	order interfaces.Order
	f     func(configFile string) []interfaces.RewriteFunctionMetadata
}

func New(configs map[string]string) {
	funcs := []initFunc{
		{name: "aboveSeries", order: aboveSeries.GetOrder(), f: aboveSeries.New},
		{name: "applyByNode", order: applyByNode.GetOrder(), f: applyByNode.New},
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
			metadata.RegisterRewriteFunction(m.Name, m.F)
		}
	}
}
