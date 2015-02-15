package d

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

var (
	mu               sync.RWMutex
	debuggingenabled bool
)

// SetDebugPrint switches the printing of debugging information based on its arg
func SetDebugPrint(enable bool) {
	mu.Lock()
	debuggingenabled = enable
	mu.Unlock()
}

func shouldprint() bool {
	mu.RLock()
	defer mu.RUnlock()
	return debuggingenabled
}

// D prints debug info about its arguments depending on whether debug printing
// is enabled or not
func D(args ...interface{}) {
	if shouldprint() {
		formatstring := ""
		for i := len(args); i >= 1; i-- {
			formatstring += " |%+v|"
		}
		log.Printf(formatstring, args...)
	}
}

// P prints debug info about its arguments always
func P(args ...interface{}) {
	formatstring := ""
	for i := len(args); i >= 1; i-- {
		formatstring += " |%+v|"
	}
	log.Printf(formatstring, args...)
}

// source https://groups.google.com/forum/?fromgroups#!topic/golang-nuts/C24fRw8HDmI
// from David Wright
type ErrorTrace struct {
	trace bytes.Buffer
}

func NewErrorTrace(skip int, args ...interface{}) error {
	buf := bytes.Buffer{}

	formatstring := ""
	for i := len(args); i >= 1; i-- {
		formatstring += " |%+v|"
	}
	formatstring += "\n  "

	if len(formatstring) != 0 {
		buf.WriteString(fmt.Sprintf(formatstring, args...))
	}

addtrace:
	pc, file, line, ok := runtime.Caller(skip)
	if ok && skip < 15 { // print a max of 15 lines of trace
		fun := runtime.FuncForPC(pc)
		buf.WriteString(fmt.Sprint(fun.Name(), " -- ", file, ":", line, "\n"))
		skip++
		goto addtrace
	}

	if buf.Len() > 0 {
		return ErrorTrace{trace: buf}
	}

	return errors.New("error generating error")
}

func (et ErrorTrace) Error() string {
	return et.trace.String()
}

// BT prints a backtrace
func BT(args ...interface{}) {
	ts := time.Now().Format("2006-02-01 15:04:05: ")
	println(ts, NewErrorTrace(2, args...).Error())
}

// FBT prints a backtrace and then panics (fatal backtrace)
func FBT(args ...interface{}) {
	ts := time.Now().Format("2006-02-01 15:04:05: ")
	println(ts, NewErrorTrace(2, args...).Error())
	panic("-----")
}

// F calls panic with a formatted string based on its arguments
func F(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
