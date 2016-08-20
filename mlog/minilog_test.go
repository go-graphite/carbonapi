package mlog

import (
	"bytes"
	"log"
	"testing"
)

func TestGetOutput(t *testing.T) {
	actual := GetOutput()
	if actual != nil {
		t.Errorf("actual output %v is different from the expected (nil)", actual)
	}
}

func TestSetRawStream(t *testing.T) {
	log.SetFlags(0)
	expectedOutput := &bytes.Buffer{}
	SetRawStream(expectedOutput)
	actualOutput := GetOutput()
	if actualOutput == nil {
		t.Errorf("actual output %v is different from the expected (not nil)", actualOutput)
	}

	logMessage := "sample message"
	expectedLog := logMessage + "\n"
	log.Print(logMessage)
	actualLog := expectedOutput.String()
	if actualLog != expectedLog {
		t.Errorf("actual log '%s' is different from the expected '%s'", actualLog, expectedLog)
	}
}
