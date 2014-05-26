// Foked from https://github.com/llimllib/loglevel

package module

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
)

// Priority used for identifying the severity of an event for Logger
const (
	Poff = iota
	Pfatal
	Perror
	Pwarn
	Pinfo
	Pdebug
	Ptrace
	Pall
)

var priorityName = []string{
	Poff:   "OFF",
	Pfatal: "FATAL",
	Perror: "ERROR",
	Pwarn:  "WARN",
	Pinfo:  "INFO",
	Pdebug: "DEBUG",
	Ptrace: "TRACE",
	Pall:   "ALL",
}

// Logger flags used for identifying the format of an event. They are
// or'ed together to control what's printed. There is no control over the
// order they appear (the order listed here) or the format they present (as
// described in the comments). A colon appears after these items:
//	2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
const (
	Ldate         = log.Ldate         // the date: 2012/01/23
	Ltime         = log.Ltime         // the time: 01:23:23
	Lmicroseconds = log.Lmicroseconds // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile     = log.Llongfile     // full file name and line number: /a/b/c/d.go:23
	Lshortfile    = log.Lshortfile    // final file name element and line number: d.go:23. overrides Llongfile
	LstdFlags     = log.LstdFlags     // initial values for the standard logger
	Lpriority     = 1 << 5            // the priority: Debug
)

type priorityMessage struct {
	priority int32
	message  string
}

// Logger defines our wrapper around the system logger
type Logger struct {
	priority int32
	prefix   string
	logger   *log.Logger
	logs     []string

	prioMsg  chan *priorityMessage
	quitWait chan bool
	mut      sync.RWMutex
}

// New creates a new Logger.
func newLogger(out io.Writer, prefix string, flag int, priority int32) *Logger {
	log := &Logger{
		priority: priority,
		prefix:   prefix,
		logger:   log.New(out, prefix, flag),
		logs:     make([]string, 0, 5),
		prioMsg:  make(chan *priorityMessage, 5),
		quitWait: make(chan bool),
	}
	go log.start()

	return log
}

// SetPrefix sets the output prefix for the logger.
func (me *Logger) SetPrefix(prefix string) {
	me.mut.Lock()
	defer me.mut.Unlock()

	me.prefix = prefix
	me.logger.SetPrefix(prefix)
}

// Prefix returns the current logger prefix
func (me *Logger) Prefix() string {
	me.mut.RLock()
	defer me.mut.RUnlock()

	return me.prefix
}

func (me *Logger) setFullPrefix(priority int32) {
	if me.logger.Flags()&Lpriority != 0 {
		me.logger.SetPrefix(fmt.Sprintf("%v %v", priorityName[priority], me.prefix))
	}
}

// Calls Output to print to the logger and append message to logs slice
func (me *Logger) print(priority int32, v ...interface{}) {
	if priority <= me.Priority() {
		me.prioMsg <- &priorityMessage{priority, fmt.Sprint(v...)}
	}
}

// Calls Output to printf to the logger and append message to logs slice
func (me *Logger) printf(priority int32, format string, v ...interface{}) {
	if priority <= me.Priority() {
		me.prioMsg <- &priorityMessage{priority, fmt.Sprintf(format, v...)}
	}
}

// Calls Output to println to the logger and append message to logs slice
func (me *Logger) println(priority int32, v ...interface{}) {
	if priority <= me.Priority() {
		me.prioMsg <- &priorityMessage{priority, fmt.Sprintln(v...)}
	}
}

func (me *Logger) start() {
	defer func() { me.quitWait <- true }()

	for msg := range me.prioMsg {
		me.setFullPrefix(msg.priority)
		me.logger.Print(msg.message)
		me.addLog(msg)
	}
}

func (me *Logger) exit() {
	close(me.prioMsg)
	<-me.quitWait
}

func (me *Logger) addLog(msg *priorityMessage) {
	msg.message = strings.TrimRight(msg.message, "\n")

	me.mut.Lock()
	defer me.mut.Unlock()

	me.logs = append(me.logs,
		fmt.Sprintf("[%5v] %v", priorityName[msg.priority], msg.message))
}

// Priority returns the output priority for the logger.
func (me *Logger) Priority() int32 {
	return atomic.LoadInt32(&me.priority)
}

// SetPriority sets the output priority for the logger.
func (me *Logger) SetPriority(priority int32) {
	atomic.StoreInt32(&me.priority, priority)
}

// SetPriorityString sets the output priority by the name of a debug level
func (me *Logger) SetPriorityString(s string) error {
	s = strings.ToUpper(s)

	// Lock unnecessary b/c only reads from priorityName
	for i, name := range priorityName {
		if name == s {
			me.SetPriority(int32(i))
			return nil
		}
	}
	return fmt.Errorf("Unable to find priority %v", s)
}

// Flags returns the output layouts for the logger.
func (me *Logger) Flags() int {
	me.mut.RLock()
	defer me.mut.RUnlock()

	return me.logger.Flags()
}

// SetFlags sets the output layouts for the logger.
func (me *Logger) SetFlags(layouts int) {
	me.mut.Lock()
	defer me.mut.Unlock()

	me.logger.SetFlags(layouts)
}

// Fatal prints the message it's given and quits the program
func (me *Logger) Fatal(v ...interface{}) {
	me.mut.Lock()
	defer me.mut.Unlock()

	me.setFullPrefix(Pfatal)
	me.logger.Fatal(v...)
}

// Fatalf prints the message it's given and quits the program
func (me *Logger) Fatalf(format string, v ...interface{}) {
	me.mut.Lock()
	defer me.mut.Unlock()

	me.setFullPrefix(Pfatal)
	me.logger.Fatalf(format, v...)
}

// Fatalln prints the message it's given and quits the program
func (me *Logger) Fatalln(v ...interface{}) {
	me.mut.Lock()
	defer me.mut.Unlock()

	me.setFullPrefix(Pfatal)
	me.logger.Fatalln(v...)
}

// Panic prints the message it's given and panic()s the program
func (me *Logger) Panic(v ...interface{}) {
	me.mut.Lock()
	defer me.mut.Unlock()

	me.setFullPrefix(Pfatal)
	me.logger.Panic(v...)
}

// Panicf prints the message it's given and panic()s the program
func (me *Logger) Panicf(format string, v ...interface{}) {
	me.mut.Lock()
	defer me.mut.Unlock()

	me.setFullPrefix(Pfatal)
	me.logger.Panicf(format, v...)
}

// Panicln prints the message it's given and panic()s the program
func (me *Logger) Panicln(v ...interface{}) {
	me.mut.Lock()
	defer me.mut.Unlock()

	me.setFullPrefix(Pfatal)
	me.logger.Panicln(v...)
}

// Error prints to the standard logger with the Error level.
func (me *Logger) Error(v ...interface{}) {
	me.print(Perror, v...)
}

// Errorf prints to the standard logger with the Error level.
func (me *Logger) Errorf(format string, v ...interface{}) {
	me.printf(Perror, format, v...)
}

// Errorln prints to the standard logger with the Error level.
func (me *Logger) Errorln(v ...interface{}) {
	me.println(Perror, v...)
}

// Warn prints to the standard logger with the Warn level.
func (me *Logger) Warn(v ...interface{}) {
	me.print(Pwarn, v...)
}

// Warnf prints to the standard logger with the Warn level.
func (me *Logger) Warnf(format string, v ...interface{}) {
	me.printf(Pwarn, format, v...)
}

// Warnln prints to the standard logger with the Warn level.
func (me *Logger) Warnln(v ...interface{}) {
	me.println(Pwarn, v...)
}

// Info prints to the standard logger with the Info level.
func (me *Logger) Info(v ...interface{}) {
	me.print(Pinfo, v...)
}

// Infof prints to the standard logger with the Info level.
func (me *Logger) Infof(format string, v ...interface{}) {
	me.printf(Pinfo, format, v...)
}

// Infoln prints to the standard logger with the Info level.
func (me *Logger) Infoln(v ...interface{}) {
	me.println(Pinfo, v...)
}

// Debug prints to the standard logger with the Debug level.
func (me *Logger) Debug(v ...interface{}) {
	me.print(Pdebug, v...)
}

// Debugf prints to the standard logger with the Debug level.
func (me *Logger) Debugf(format string, v ...interface{}) {
	me.printf(Pdebug, format, v...)
}

// Debugln prints to the standard logger with the Debug level.
func (me *Logger) Debugln(v ...interface{}) {
	me.println(Pdebug, v...)
}

// Trace prints to the standard logger with the Trace level.
func (me *Logger) Trace(v ...interface{}) {
	me.print(Ptrace, v...)
}

// Tracef prints to the standard logger with the Trace level.
func (me *Logger) Tracef(format string, v ...interface{}) {
	me.printf(Ptrace, format, v...)
}

// Traceln prints to the standard logger with the Trace level.
func (me *Logger) Traceln(v ...interface{}) {
	me.println(Ptrace, v...)
}

// Return a copy of logs[:min(n, len(logs))]. If n is 0 a copy of logs is returned.
// Calling Logs() with n < 0 is the same as calling TailLogs(-n)
func (me *Logger) Logs(n int) []string {
	if n < 0 {
		return me.TailLogs(-n)
	}

	me.mut.RLock()
	defer me.mut.RUnlock()

	if n == 0 || n > len(me.logs) {
		n = len(me.logs)
	}

	return copySlice(me.logs[:n])
}

// Returns a copy of logs[len(logs) - min(n, len(logs):]. Calling TailLogs(n)
// where n < 0 is equivalent to calling Logs(-n)
func (me *Logger) TailLogs(n int) []string {
	if n == 0 {
		return make([]string, 0)
	} else if n < 0 {
		return me.Logs(-n)
	}

	me.mut.RLock()
	defer me.mut.RUnlock()

	if n >= len(me.logs) {
		n = len(me.logs)
	}

	return copySlice(me.logs[len(me.logs)-n:])
}

func (me *Logger) LenLogs() int {
	me.mut.RLock()
	defer me.mut.RUnlock()

	return len(me.logs)
}

// Clears saved logs slice, not those stored to disk
func (me *Logger) ClearLogs() {
	me.mut.Lock()
	defer me.mut.Unlock()

	clear(&me.logs)
}
