package merry

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"

	"runtime"
)

func init() {
	RegisterDetail("User Message", userMessage)
	RegisterDetail("HTTP Code", httpCode)
}

var detailsLock sync.Mutex
var detailFields map[string]interface{}

// RegisterDetail registers an error property key in a global registry, with a label.
// The registry is used by the Details() function.  Registered error properties will
// be included in Details() output, if the value of that error property is not nil.
// For example:
//
//     err := New("boom")
//     err = err.WithValue(colorKey, "red")
//     fmt.Println(Details(err))
//
//     // Output:
//     // boom
//     //
//     // <stacktrace>
//
//     RegisterDetail("Color", colorKey)
//     fmt.Println(Details(err))
//
//     // Output:
//     // boom
//     // Color: red
//     //
//     // <stacktrace>
//
// Error property keys are typically not exported by the packages which define them.
// Packages instead export functions which let callers access that property.
// It's therefore up to the package
// to register those properties which would make sense to include in the Details() output.
// In other words, it's up to the author of the package which generates the errors
// to publish printable error details, not the callers of the package.
func RegisterDetail(label string, key interface{}) {
	detailsLock.Lock()
	defer detailsLock.Unlock()

	if detailFields == nil {
		detailFields = map[string]interface{}{}
	}
	detailFields[label] = key
}

// Location returns zero values if e has no stacktrace
func Location(e error) (file string, line int) {
	s := Stack(e)
	if len(s) > 0 {
		fnc, _ := runtime.CallersFrames(s[:1]).Next()
		return fnc.File, fnc.Line
	}
	return "", 0
}

// SourceLine returns the string representation of
// Location's result or an empty string if there's
// no stracktrace.
func SourceLine(e error) string {
	file, line := Location(e)
	if line != 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}

// Stacktrace returns the error's stacktrace as a string formatted
// the same way as golangs runtime package.
// If e has no stacktrace, returns an empty string.
func Stacktrace(e error) string {
	s := Stack(e)
	if len(s) > 0 {
		buf := bytes.Buffer{}
		frames := runtime.CallersFrames(s)
		for {
			frame, more := frames.Next()
			buf.WriteString(frame.Function)
			buf.WriteString(fmt.Sprintf("\n\t%s:%d\n", frame.File, frame.Line))
			if !more {
				break
			}

		}
		return buf.String()
	}
	return ""
}

// Details returns e.Error(), e's stacktrace, and any additional details which have
// be registered with RegisterDetail.  User message and HTTP code are already registered.
//
// The details of each error in e's cause chain will also be printed.
func Details(e error) string {
	if e == nil {
		return ""
	}
	msg := Message(e)

	detailsLock.Lock()

	var dets []string
	for label, key := range detailFields {
		v := Value(e, key)
		if v != nil {
			dets = append(dets, fmt.Sprintf("%s: %v", label, v))
		}
	}

	detailsLock.Unlock()
	if len(dets) > 0 {
		// sort so output is predictable
		sort.Strings(dets)
		msg += "\n" + strings.Join(dets, "\n")
	}

	//userMsg := UserMessage(e)
	//if userMsg != "" {
	//	msg = fmt.Sprintf("%s\n\nUser Message: %s", msg, userMsg)
	//}
	s := Stacktrace(e)
	if s != "" {
		msg += "\n\n" + s
	}

	if c := Cause(e); c != nil {
		msg += "\n\nCaused By: " + Details(c)
	}
	return msg
}
