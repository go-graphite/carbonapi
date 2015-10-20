// mini logging package
package mlog

import (
	"io"
	"log"
	"os"

	"github.com/lestrrat/go-file-rotatelogs"
)

type Level int

const (
	Normal Level = iota
	Debug
	Trace
)

func SetOutput(logdir, service string, stdout bool) {
	rl := rotatelogs.NewRotateLogs(
		logdir + "/" + service + ".%Y%m%d%H%M.log",
	)

	// Optional fields must be set afterwards
	rl.LinkName = logdir + "/" + service + ".log"

	if stdout {
		log.SetOutput(io.MultiWriter(os.Stdout, rl))
	} else {
		log.SetOutput(rl)
	}
}

func (ll Level) Debugf(format string, a ...interface{}) {
	if ll >= Debug {
		log.Printf(format, a...)
	}
}

func (ll Level) Debugln(a ...interface{}) {
	if ll >= Debug {
		log.Println(a...)
	}
}

func (ll Level) Tracef(format string, a ...interface{}) {
	if ll >= Trace {
		log.Printf(format, a...)
	}
}

func (ll Level) Traceln(a ...interface{}) {
	if ll >= Trace {
		log.Println(a...)
	}
}
func (ll Level) Logln(a ...interface{}) {
	log.Println(a...)
}

func (ll Level) Logf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func (ll Level) Fatalf(format string, a ...interface{}) {
	log.Printf(format, a...)
	os.Exit(1)
}

func (ll Level) Fatalln(a ...interface{}) {
	log.Println(a...)
	os.Exit(1)
}
