package d

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/destinygg/website2/internal/config"
	"golang.org/x/net/context"
)

const (
	EnableDebug  = true
	DisableDebug = false
)

var (
	mu               sync.RWMutex
	debuggingenabled bool
)

// Init initializes the printing of debugging information based on its arg
func Init(ctx context.Context) context.Context {
	cfg := ctx.Value("appconfig").(*config.AppConfig)

	mu.Lock()
	debuggingenabled = cfg.Debug.Debug
	mu.Unlock()

	logfile := cfg.Debug.Logfile
	if logfile == "" {
		logfile = "logs/debug.txt"
	}

	w, err := os.OpenFile(logfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		panic(logfile + err.Error())
	}
	mw := io.MultiWriter(os.Stderr, w)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	return ctx
}

func shouldprint() bool {
	mu.RLock()
	defer mu.RUnlock()
	return debuggingenabled
}

func getformatstr(numargs int, addlineinfo bool) string {
	var formatstring string
	if addlineinfo {
		formatstring = "%+v:%+v \n  "
	} else {
		formatstring = "\n  "
	}
	for i := numargs; i >= 1; i-- {
		formatstring += "|%+v|\n  "
	}
	formatstring += "\n\n"
	return formatstring
}

func logwithcaller(file string, line int, args ...interface{}) {
	pos := strings.LastIndex(file, "website2/") + len("website2/")
	newargs := make([]interface{}, 0, len(args)+2)
	newargs = append(newargs, file[pos:], line)
	newargs = append(newargs, args...)
	log.Printf(getformatstr(len(args), true), newargs...)
}

// D prints debug info about its arguments depending on whether debug printing
// is enabled or not
func D(args ...interface{}) {
	if shouldprint() {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			logwithcaller(file, line, args...)
		} else {
			log.Printf(getformatstr(len(args), false), args...)
		}
	}
}

// P prints debug info about its arguments always
func P(args ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		logwithcaller(file, line, args...)
	} else {
		log.Printf(getformatstr(len(args), false), args...)
	}
}

// source https://groups.google.com/forum/?fromgroups#!topic/golang-nuts/C24fRw8HDmI
// from David Wright
type ErrorTrace struct {
	trace bytes.Buffer
}

func NewErrorTrace(skip int, args ...interface{}) error {
	buf := bytes.Buffer{}

	formatstring := "\n  "
	for i := len(args); i >= 1; i-- {
		formatstring += "|%+v|\n  "
	}
	formatstring += "\n"

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
