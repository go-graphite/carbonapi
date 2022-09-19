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

	fmt.Fprintln(writer, `package functions

import (
	"sort"
	"strings"
`)
	for _, m := range funcs {
		if m == "config" {
			continue
		}
		fmt.Fprintf(writer, "	\"github.com/go-graphite/carbonapi/expr/functions/%s\"\n", m)
	}
	fmt.Fprintf(writer, `	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
)

type initFunc struct {
	name     string
	filename string
	order    interfaces.Order
	f        func(configFile string) []interfaces.FunctionMetadata
}

func New(configs map[string]string) {
	funcs := []initFunc{`)
	for _, m := range funcs {
		if m == "config" {
			continue
		}
		fmt.Fprintf(writer, `
		{name: "%s", filename: "%s", order: %s.GetOrder(), f: %s.New},`, m, m, m, m)

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
			metadata.RegisterFunctionWithFilename(m.Name, f.filename, m.F)
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
