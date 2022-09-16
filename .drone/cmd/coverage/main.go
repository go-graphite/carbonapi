package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type aggregation struct {
	NumStmts        int
	NumCoveredStmts int
}

func (a *aggregation) CoveragePct() string {
	pct := (float64(a.NumCoveredStmts) / float64(a.NumStmts)) * 100
	return fmt.Sprintf("%.1f%%", pct)
}

func main() {
	covFilename := flag.String("f", "coverage.out", "Output of `go test -coverprofile=coverage.out ./...`")
	flag.Parse()

	f, err := os.Open(*covFilename)
	if err != nil {
		panic(err)
	}

	agg := make(map[string]*aggregation)

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	s.Scan() // First line specifies the mode; it doesn't affect what we do so we can just skip it.

	for s.Scan() {
		line := s.Text()

		cols := strings.Fields(line)
		key := strings.Split(cols[0], ":")[0]

		numStmts, err := strconv.Atoi(cols[1])
		if err != nil {
			panic(err)
		}
		numTimesCovered, err := strconv.Atoi(cols[2])
		if err != nil {
			panic(err)
		}

		if val, ok := agg[key]; ok {
			val.NumStmts += numStmts
			if numTimesCovered > 0 {
				val.NumCoveredStmts += numStmts
			}
		} else {
			numCoveredStmts := 0
			if numTimesCovered > 0 {
				numCoveredStmts = numStmts
			}

			agg[key] = &aggregation{
				NumStmts:        numStmts,
				NumCoveredStmts: numCoveredStmts,
			}
		}
	}

	keys := make([]string, 0, len(agg))
	for k := range agg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("Go coverage report:")
	fmt.Println("<details>")
	fmt.Println("<summary>Click to expand.</summary>")
	fmt.Println("")
	fmt.Println("| File | % |")
	fmt.Println("| ---- | - |")

	totalStmts := 0
	totalCoveredStmts := 0

	for _, k := range keys {
		a := agg[k]
		fmt.Printf("| %s | %s |\n", k, a.CoveragePct())

		totalStmts += a.NumStmts
		totalCoveredStmts += a.NumCoveredStmts
	}

	fmt.Printf("| total | %.1f%% |\n", (float64(totalCoveredStmts)/float64(totalStmts))*100)
	fmt.Println("</details>")
}
