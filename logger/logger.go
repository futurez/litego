// Usage:
//
// import "github.com/zhoufuture/golite/logger"
//
// Use it like this:
//  logger.Fatal("fatal")
//  logger.Panic("panic")
//  logger.Error("error")
//	logger.Info("info")
//	logger.Warn("warn")
//	logger.Debug("debug")
package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"runtime"
	"sync"

	"github.com/zhoufuture/golite/util"
)

const (
	LevelPanic = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

const (
	CONSOLE_PROTOCOL = "console"
	FILE_PROTOCOL    = "file"
	TCP_PROTOCOL     = "tcp"
	UDP_PROTOCOL     = "udp"
)

type LoggerAdapter interface {
	newLoggerInstance() LoggerInterface
}

type LoggerInterface interface {
	Init(config string) error
	WriteMsg(msg string, level int) error
	Close()
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
	funcdepth int
	localip   string
	appname   string
	bprefix   bool
	prefix    string
	syncClose chan bool
	msgQueue  chan *logMsg
	outputs   map[string]LoggerInterface
}

func NewLogger(channellen int64) *Logger {
	lg := &Logger{
		funcdepth: 3,
		localip:   util.GetIntranetIP(),
		appname:   util.GetAppName(),
		syncClose: make(chan bool),
		msgQueue:  make(chan *logMsg, channellen),
		outputs:   make(map[string]LoggerInterface),
	}
	go lg.save()
	return lg
}

func (lg *Logger) SetLogger(name, config string) error {
	lg.Lock()
	defer lg.Unlock()
	if adapter, ok := adapters[name]; ok {
		output := adapter.newLoggerInstance()
		err := output.Init(config)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		lg.outputs[name] = output
	} else {
		log.Printf("unknown adaptername %s\n", name)
		return fmt.Errorf("unknown adaptername %s", name)
	}
	return nil
}

func (lg *Logger) DelLogger(name string) error {
	lg.Lock()
	defer lg.Unlock()
	if output, ok := lg.outputs[name]; ok {
		output.Close()
		delete(lg.outputs, name)
		return nil
	} else {
		return fmt.Errorf("logger: unknown adaptername %q", name)
	}
}

func (lg *Logger) write(loglevel int, msg string) {
	lm := &logMsg{level: loglevel}
	if lg.funcdepth > 0 {
		_, file, line, ok := runtime.Caller(lg.funcdepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)
		if lg.bprefix {
			lm.msg = fmt.Sprintf("%s/%s/%s/%s:%d %s", lg.localip, lg.appname, lg.prefix, filename, line, msg)
		} else {
			lm.msg = fmt.Sprintf("%s/%s/%s:%d %s", lg.localip, lg.appname, filename, line, msg)
		}
	} else {
		if lg.bprefix {
			lm.msg = fmt.Sprintf("%s/%s/%s %s", lg.localip, lg.appname, lg.prefix, msg)
		} else {
			lm.msg = fmt.Sprintf("%s/%s %s", lg.localip, lg.appname, msg)
		}
	}
	lg.msgQueue <- lm
	if lm.level == LevelPanic {
		panic(lm.msg)
	}
}

func (lg *Logger) save() {
	for lm := range lg.msgQueue {
		for _, output := range lg.outputs {
			err := output.WriteMsg(lm.msg, lm.level)
			if err != nil {
				log.Println("ERROR, unable to WriteMsg:", err)
			}
		}
	}
	for _, output := range lg.outputs {
		output.Close()
	}
	lg.syncClose <- true
}

func (lg *Logger) SetFuncDepth(depth int) {
	lg.funcdepth = depth
}

func (lg Logger) GetFuncDepth() int {
	return lg.funcdepth
}

func (lg *Logger) SetPrefix(prefix string) {
	if len(prefix) > 0 {
		lg.prefix = prefix
		lg.bprefix = true
	}
}

func (lg Logger) GetPrefix() string {
	return lg.prefix
}

func (lg *Logger) Panic(format string, v ...interface{}) {
	msg := fmt.Sprintf("[P] "+format, v...)
	lg.write(LevelPanic, msg)
}

func (lg *Logger) Error(format string, v ...interface{}) {
	msg := fmt.Sprintf("[E] "+format, v...)
	lg.write(LevelError, msg)
}

func (lg *Logger) Warn(format string, v ...interface{}) {
	msg := fmt.Sprintf("[W] "+format, v...)
	lg.write(LevelWarn, msg)
}

func (lg *Logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf("[I] "+format, v...)
	lg.write(LevelInfo, msg)
}

func (lg *Logger) Debug(format string, v ...interface{}) {
	msg := fmt.Sprintf("[D] "+format, v...)
	lg.write(LevelDebug, msg)
}

func (lg *Logger) Close() {
	if lg.msgQueue != nil {
		close(lg.msgQueue)
		lg.msgQueue = nil
	}
	<-lg.syncClose
}

var stdLogger *Logger

func getlogname() string {
	return util.GetCurrentPath() + "/../log/" + util.GetAppName() + ".log"
}

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	stdLogger = NewLogger(10000)
	stdLogger.SetFuncDepth(3)

	var consoleconf ConsoleLogConfig
	consoleconf.LogLevel = LevelDebug
	consoleconfbuf, _ := json.Marshal(consoleconf)
	stdLogger.SetLogger(CONSOLE_PROTOCOL, string(consoleconfbuf))

	var fileconf FileLogConfig
	fileconf.LogFlag = (log.Ldate | log.Ltime | log.Lmicroseconds)
	fileconf.FileName = getlogname()
	fileconf.MaxDays = 7
	fileconf.MaxSize = 1 << 30
	fileconf.LogLevel = LevelDebug
	fileconfbuf, _ := json.Marshal(fileconf)
	stdLogger.SetLogger(FILE_PROTOCOL, string(fileconfbuf))
}

func SetTcpLog(jsonconfig string) {
	stdLogger.SetLogger(TCP_PROTOCOL, jsonconfig)
}

func SetUdpLog(jsonconfig string) {
	stdLogger.SetLogger(UDP_PROTOCOL, jsonconfig)
}

func SetPrefix(prefix string) {
	stdLogger.SetPrefix(prefix)
}

func GetPrefix() string {
	return stdLogger.GetPrefix()
}

func Panic(format string, v ...interface{}) {
	stdLogger.Panic(format, v...)
}

func Error(format string, v ...interface{}) {
	stdLogger.Error(format, v...)
}

func Warn(format string, v ...interface{}) {
	stdLogger.Warn(format, v...)
}

func Info(format string, v ...interface{}) {
	stdLogger.Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	stdLogger.Debug(format, v...)
}

func Close() {
	stdLogger.Close()
}
