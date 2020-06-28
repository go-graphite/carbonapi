// +build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	funcs := make([]string, 0)
	for _, f := range files {
		name := f.Name()
		if strings.Contains(name, ".go") {
			continue
		}
		if name == "example" {
			continue
		}

		funcs = append(funcs, name)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	fmt.Fprintln(writer, `package rewrite

import (
	"sort"
	"strings"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"`)
	for _, m := range funcs {
		fmt.Fprintf(writer, "	\"github.com/go-graphite/carbonapi/expr/rewrite/%s\"\n", m)
	}
	fmt.Fprintf(writer, `)

type initFunc struct {
	name  string
	order interfaces.Order
	f     func(configFile string) []interfaces.RewriteFunctionMetadata
}

func New(configs map[string]string) {
	funcs := []initFunc{`)
	for _, m := range funcs {
		fmt.Fprintf(writer, `
		{name: "%s", order: %s.GetOrder(), f: %s.New},`, m, m, m)

	}

	fmt.Fprintln(writer, `
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
}`)

	err = writer.Flush()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("glue.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	f.Write(b.Bytes())

	f.Close()
}
