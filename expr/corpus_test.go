package expr

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestCorpus(t *testing.T) {

	corpusFiles, err := filepath.Glob("workdir/crashers/*")
	if err != nil {
		t.Errorf("unable to glob: %v", err)
	}

	for _, corpusFile := range corpusFiles {

		if strings.HasSuffix(corpusFile, ".quoted") || strings.HasSuffix(corpusFile, ".output") {
			continue
		}

		t.Log(corpusFile)

		contents, err := ioutil.ReadFile(corpusFile)

		if err != nil {
			t.Errorf("error opening workdir/crashers/%s: %v", corpusFile, err)
			return
		}

		_, rem, err := ParseExpr(strings.TrimSpace(string(contents)))
		if rem != "" || err != nil {
			t.Errorf("error parsing: %s: %q: %v, rem=%q", corpusFile, contents, err, rem)
		}
	}

}
