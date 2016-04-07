// Usage:
//
// import "Common/logger"
//
// Use it like this:
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
	"net"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
)

const (
	CONSOLE_PROTOCOL = "console"
	FILE_PROTOCOL    = "file"
	TCP_PROTOCOL     = "tcp"
	UDP_PROTOCOL     = "udp"
	ALL_PROTOCOL     = "all"
)

type LoggerAdapter interface {
	newLoggerInstance() LoggerInterface
}

type LoggerInterface interface {
	Init(config string) error
	SetLogLevel(loglevel int)
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
	async     bool
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
		async:     false,
		localip:   GetIntranetIP(),
		appname:   GetAppName(),
		syncClose: make(chan bool),
		msgQueue:  make(chan *logMsg, channellen),
		outputs:   make(map[string]LoggerInterface),
	}
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

func split(path string) string {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return path
	}
	file := path[i+1:]
	path = path[:i]
	i = strings.LastIndex(path, "/")
	return path[i+1:] + "/" + file
}

func (lg *Logger) write(loglevel int, msg string) {
	lm := &logMsg{level: loglevel}
	if lg.funcdepth > 0 {
		_, file, line, ok := runtime.Caller(lg.funcdepth)
		if !ok {
			file = "???"
			line = 0
		}
		filename := split(file)
		if lg.bprefix {
			lm.msg = fmt.Sprintf("%s %s %s %s:%d %s", lg.localip, lg.appname, lg.prefix, filename, line, msg)
		} else {
			lm.msg = fmt.Sprintf("%s %s %s:%d %s", lg.localip, lg.appname, filename, line, msg)
		}
	} else {
		if lg.bprefix {
			lm.msg = fmt.Sprintf("%s %s %s %s", lg.localip, lg.appname, lg.prefix, msg)
		} else {
			lm.msg = fmt.Sprintf("%s %s %s", lg.localip, lg.appname, msg)
		}
	}

	if loglevel == LevelPanic {
		lg.outputMsg(lm)
		panic(lm.msg)
		return
	}

	if lg.async {
		lg.msgQueue <- lm
	} else {
		lg.outputMsg(lm)
	}
}

func (lg *Logger) outputMsg(lm *logMsg) {
	lg.Lock()
	defer lg.Unlock()
	for _, output := range lg.outputs {
		err := output.WriteMsg(lm.msg, lm.level)
		if err != nil {
			log.Println("ERROR, unable to WriteMsg:", err)
		}
	}
}

func (lg *Logger) save() {
	if !lg.async {
		return
	}
	for lm := range lg.msgQueue {
		lg.outputMsg(lm)
	}
	<-lg.syncClose
}

func (lg *Logger) StartAsyncSave() {
	if !lg.async {
		lg.async = true
		go lg.save()
	}
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

func (lg *Logger) SetLogLevel(protocol string, loglevel int) error {
	switch protocol {
	case CONSOLE_PROTOCOL:
		fallthrough
	case FILE_PROTOCOL:
		fallthrough
	case TCP_PROTOCOL:
		fallthrough
	case UDP_PROTOCOL:
		output := lg.outputs[protocol]
		if output == nil {
			return fmt.Errorf("%s not use...", protocol)
		}
		output.SetLogLevel(loglevel)
	case ALL_PROTOCOL:
		for _, output := range lg.outputs {
			output.SetLogLevel(loglevel)
		}
	default:
		return fmt.Errorf("fuck you, err protocol %s.", protocol)
	}
	return nil
}

func (lg *Logger) Panic(v ...interface{}) {
	lg.write(LevelPanic, "[P] "+fmt.Sprint(v...))
}

func (lg *Logger) Error(v ...interface{}) {
	lg.write(LevelError, "[E] "+fmt.Sprint(v...))
}

func (lg *Logger) Warn(v ...interface{}) {
	lg.write(LevelWarn, "[W] "+fmt.Sprint(v...))
}

func (lg *Logger) Info(v ...interface{}) {
	lg.write(LevelInfo, "[I] "+fmt.Sprint(v...))
}

func (lg *Logger) Debug(v ...interface{}) {
	lg.write(LevelDebug, "[D] "+fmt.Sprint(v...))
}

func (lg *Logger) Panicf(format string, v ...interface{}) {
	lg.write(LevelPanic, fmt.Sprintf("[P] "+format, v...))
}

func (lg *Logger) Errorf(format string, v ...interface{}) {
	lg.write(LevelError, fmt.Sprintf("[E] "+format, v...))
}

func (lg *Logger) Warnf(format string, v ...interface{}) {
	lg.write(LevelWarn, fmt.Sprintf("[W] "+format, v...))
}

func (lg *Logger) Infof(format string, v ...interface{}) {
	lg.write(LevelInfo, fmt.Sprintf("[I] "+format, v...))
}

func (lg *Logger) Debugf(format string, v ...interface{}) {
	lg.write(LevelDebug, fmt.Sprintf("[D] "+format, v...))
}

func (lg *Logger) PrintStack() {
	lm := logMsg{
		level: LevelError,
		msg:   string(debug.Stack()),
	}
	lg.outputMsg(&lm)
}

func (lg *Logger) Close() {
	if lg.async {
		if lg.msgQueue != nil {
			close(lg.msgQueue)
			lg.msgQueue = nil
		}
		lg.syncClose <- true
	}
	for _, output := range lg.outputs {
		output.Close()
	}
}

var stdLogger *Logger

func getlogname() string {
	return GetCurrentPath() + "/../log/" + GetAppName() + ".log"
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

func StartAsyncSave() {
	stdLogger.StartAsyncSave()
}

func SetTcpLog(jsonconfig string) {
	stdLogger.SetLogger(TCP_PROTOCOL, jsonconfig)
}

func SetUdpLog(jsonconfig string) {
	stdLogger.SetLogger(UDP_PROTOCOL, jsonconfig)
}

func SetLogLevel(protocol string, loglevel int) error {
	return stdLogger.SetLogLevel(protocol, loglevel)
}

func SetPrefix(prefix string) {
	stdLogger.SetPrefix(prefix)
}

func GetPrefix() string {
	return stdLogger.GetPrefix()
}

func Panic(v ...interface{}) {
	stdLogger.Panic(v...)
}

func Error(v ...interface{}) {
	stdLogger.Error(v...)
}

func Warn(v ...interface{}) {
	stdLogger.Warn(v...)
}

func Info(v ...interface{}) {
	stdLogger.Info(v...)
}

func Debug(v ...interface{}) {
	stdLogger.Debug(v...)
}

func Panicf(format string, v ...interface{}) {
	stdLogger.Panicf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	stdLogger.Errorf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	stdLogger.Warnf(format, v...)
}

func Infof(format string, v ...interface{}) {
	stdLogger.Infof(format, v...)
}

func Debugf(format string, v ...interface{}) {
	stdLogger.Debugf(format, v...)
}

func PrintStack() {
	stdLogger.PrintStack()
}

func Close() {
	stdLogger.Close()
}

func GetAppName() string {
	execfile := os.Args[0]
	if runtime.GOOS == `windows` {
		execfile = strings.Replace(execfile, "\\", "/", -1)
	}
	_, filename := path.Split(execfile)
	return filename
}

func GetCurrentPath() string {
	curpath, _ := os.Getwd()
	return curpath
}

func GetIntranetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ""
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip == nil {
			continue
		}
		if IsIntranetIP(ip.String()) {
			return ip.String()
		}
	}
	return "127.0.0.1"
}

// 10.0.0.0 ~ 10.255.255.255(A)
// 172.16.0.0 ~ 172.31.255.255(B)
// 192.168.0.0 ~ 192.168.255.255(C)
func IsIntranetIP(ip string) bool {
	if strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.168.") {
		return true
	}
	if strings.HasPrefix(ip, "172.") {
		arr := strings.Split(ip, ".")
		if len(arr) != 4 {
			return false
		}
		second, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return false
		}
		if second >= 16 && second <= 31 {
			return true
		}
	}
	return false
}
