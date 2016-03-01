// Usage:
//
// import "golite/logger"
//
// Use it like this:
//  logger.Fatal("trace")
//	logger.Info("info")
//	logger.Warn("warning")
//	logger.Debug("debug")
package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

const (
	LevelFatal = iota
	LevelPanic
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelNormal
)

const (
	FILE_PROTOCOL_LOG    = "file"
	CONSOLE_PROTOCOL_LOG = "console"
	TCP_PROTOCOL_LOG     = "tcp"
	UDP_PROTOCOL_LOG     = "udp"
)

type LoggerAdapter interface {
	newLoggerInstance() LoggerInterface
}

type LoggerInterface interface {
	Init(config string) error
	WriteMsg(msg string, level int) error
	Destroy()
	Flush()
}

var adapters = make(map[string]LoggerAdapter)

func Register(name string, adapter LoggerAdapter) {
	if adapter == nil {
		panic("logger: Register adapter is nil")
	}

	if _, dup := adapters[name]; dup {
		panic("logger: Register called twice for provider " + name)
	}
	adapters[name] = adapter
}

type logMsg struct {
	level int
	msg   string
}

type Logger struct {
	sync.Mutex
	level          int
	enableFuncCall bool
	funcCallDepth  int
	asynOutput     bool
	msgQueue       chan *logMsg
	outputs        map[string]LoggerInterface
}

func NewLogger(channellen int64) *Logger {
	return &Logger{
		level:          LevelDebug,
		enableFuncCall: true,
		funcCallDepth:  3,
		asynOutput:     false,
		msgQueue:       make(chan *logMsg, channellen),
		outputs:        make(map[string]LoggerInterface),
	}
}

func (this *Logger) Async() *Logger {
	this.asynOutput = true
	go this.startLogger()
	return this
}

func (this *Logger) SetLogger(name string, config string) error {
	this.Lock()
	defer this.Unlock()
	if adapter, ok := adapters[name]; ok {
		lg := adapter.newLoggerInstance()
		err := lg.Init(config)
		this.outputs[name] = lg
		if err != nil {
			fmt.Println("logger.SetLogger: " + err.Error())
			return err
		}
	} else {
		return fmt.Errorf("logger: unknown adaptername %q", name)
	}
	return nil
}

func (this *Logger) DelLogger(name string) error {
	this.Lock()
	defer this.Unlock()
	if lg, ok := this.outputs[name]; ok {
		lg.Destroy()
		delete(this.outputs, name)
		return nil
	} else {
		return fmt.Errorf("logger: unknown adaptername %q (forgotten Register?)", name)
	}
}

func (this *Logger) writerMsg(loglevel int, msg string) error {
	lm := new(logMsg)
	lm.level = loglevel
	if this.enableFuncCall {
		_, file, line, ok := runtime.Caller(this.funcCallDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)
		lm.msg = fmt.Sprintf("[%s:%d] %s", filename, line, msg)
	} else {
		lm.msg = msg
	}
	if this.asynOutput {
		this.msgQueue <- lm
	} else {
		for name, l := range this.outputs {
			err := l.WriteMsg(lm.msg, lm.level)
			if err != nil {
				fmt.Println("unable to WriteMsg to adapter:", name, err)
				return err
			}

			if lm.level == LevelPanic {
				panic(lm.msg)
			} else if lm.level == LevelFatal {
				os.Exit(1)
			}
		}
	}
	return nil
}

func (this *Logger) SetLogLevel(level int) {
	this.level = level
}

func (this *Logger) SetFuncCallDepth(depth int) {
	this.funcCallDepth = depth
}

func (this *Logger) GetFuncCallDepth() int {
	return this.funcCallDepth
}

func (this *Logger) SetEnableFuncCall(b bool) {
	this.enableFuncCall = b
}

func (this *Logger) GetEnableFuncCall() bool {
	return this.enableFuncCall
}

func (this *Logger) startLogger() {
	for {
		select {
		case lm := <-this.msgQueue:
			for _, l := range this.outputs {
				err := l.WriteMsg(lm.msg, lm.level)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg:", err)
				}

				if lm.level == LevelPanic {
					panic(lm.msg)
				} else if lm.level == LevelFatal {
					os.Exit(1)
				}
			}
		}
	}
}

func (this *Logger) Fatal(format string, v ...interface{}) {
	if LevelFatal > this.level {
		return
	}
	msg := fmt.Sprintf("[F] "+format, v...)
	this.writerMsg(LevelFatal, msg)
}

func (this *Logger) Panic(format string, v ...interface{}) {
	if LevelPanic > this.level {
		return
	}
	msg := fmt.Sprintf("[P] "+format, v...)
	this.writerMsg(LevelPanic, msg)
}

func (this *Logger) Error(format string, v ...interface{}) {
	if LevelError > this.level {
		return
	}
	msg := fmt.Sprintf("[E] "+format, v...)
	this.writerMsg(LevelError, msg)
}

func (this *Logger) Warn(format string, v ...interface{}) {
	if LevelWarn > this.level {
		return
	}
	msg := fmt.Sprintf("[W] "+format, v...)
	this.writerMsg(LevelWarn, msg)
}

func (this *Logger) Info(format string, v ...interface{}) {
	if LevelInfo > this.level {
		return
	}
	msg := fmt.Sprintf("[I] "+format, v...)
	this.writerMsg(LevelInfo, msg)
}

func (this *Logger) Debug(format string, v ...interface{}) {
	if LevelDebug > this.level {
		return
	}
	msg := fmt.Sprintf("[D] "+format, v...)
	this.writerMsg(LevelDebug, msg)
}

func (this *Logger) Normal(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	this.writerMsg(LevelNormal, msg)
}

func (this *Logger) Flush() {
	for _, l := range this.outputs {
		l.Flush()
	}
}

func (this *Logger) Close() {
	for {
		if len(this.msgQueue) > 0 {
			bm := <-this.msgQueue
			for _, l := range this.outputs {
				err := l.WriteMsg(bm.msg, bm.level)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg (while closing logger):", err)
				}
			}
			continue
		}
		break
	}
	for _, l := range this.outputs {
		l.Flush()
		l.Destroy()
	}
}

var DefaultLogger *Logger

func getlogname() string {
	pathfile := os.Args[0]
	curpath, _ := os.Getwd()
	filename := strings.Replace(pathfile, curpath, "", -1)
	filename = filename[1:]
	names := strings.Split(filename, ".")
	name := names[0]
	if name == "" {
		name = string(os.Getpid())
	}
	logname := curpath + "/../log/" + name + ".log"
	if runtime.GOOS == "windows" {
		logname = strings.Replace(logname, "\\", "/", -1)
	}
	return logname
}

func init() {
	DefaultLogger = NewLogger(10000)
	DefaultLogger.SetLogger(CONSOLE_PROTOCOL_LOG, "")

	var fileconf FileLogConfig
	fileconf.LogFlag = (log.Ldate | log.Ltime | log.Lmicroseconds)
	fileconf.FileName = getlogname()
	fileconf.MaxDays = 7
	fileconf.MaxSize = 1 << 30
	fileconfbuf, _ := json.Marshal(fileconf)
	DefaultLogger.SetLogger(FILE_PROTOCOL_LOG, string(fileconfbuf))

	DefaultLogger.SetEnableFuncCall(true)
	DefaultLogger.SetFuncCallDepth(3)
	DefaultLogger.Async()
}

func SetLogLevel(level int) {
	DefaultLogger.SetLogLevel(level)
}

func Fatal(format string, v ...interface{}) {
	DefaultLogger.Fatal(format, v...)
}

func Panic(format string, v ...interface{}) {
	DefaultLogger.Panic(format, v...)
}

func Error(format string, v ...interface{}) {
	DefaultLogger.Error(format, v...)
}

func Warn(format string, v ...interface{}) {
	DefaultLogger.Warn(format, v...)
}

func Info(format string, v ...interface{}) {
	DefaultLogger.Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	DefaultLogger.Debug(format, v...)
}

func Flush() {
	DefaultLogger.Flush()
}

func Close() {
	DefaultLogger.Close()
}
