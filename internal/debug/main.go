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
type errorTrace struct {
	err   error
	trace string
}

func newErrorTrace(args ...interface{}) error {
	formatstring := ""
	for i := len(args); i >= 1; i-- {
		formatstring += " |%+v|"
	}

	msg := fmt.Sprintf(formatstring, args...)
	buf := bytes.Buffer{}
	skip := 2

addtrace:
	pc, file, line, ok := runtime.Caller(skip)
	if ok && skip < 6 { // print a max of 6 lines of trace
		fun := runtime.FuncForPC(pc)
		buf.WriteString(fmt.Sprint(fun.Name(), " -- ", file, ":", line, "\n"))
		skip++
		goto addtrace
	}

	if buf.Len() > 0 {
		trace := buf.String()
		return errorTrace{err: errors.New(msg), trace: trace}
	}

	return errors.New("error generating error")
}

func (et errorTrace) Error() string {
	return et.err.Error() + "\n  " + et.trace
}

// BT prints a backtrace
func BT(args ...interface{}) {
	ts := time.Now().Format("2006-02-01 15:04:05: ")
	println(ts, newErrorTrace(args...).Error())
}

// FBT prints a backtrace and then panics (fatal backtrace)
func FBT(args ...interface{}) {
	ts := time.Now().Format("2006-02-01 15:04:05: ")
	println(ts, newErrorTrace(args...).Error())
	panic("-----")
}

// F calls panic with a formatted string based on its arguments
func F(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
