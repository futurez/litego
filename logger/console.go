package logger

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
)

type brush func(string) string

func newBrush(color string) brush {
	reset := "\033[0m"
	return func(text string) string {
		return color + text + reset
	}
}

var colors = []brush{
	newBrush("\033[0m"),     // Debue  endcolor
	newBrush("\033[01;32m"), // Info   green
	newBrush("\033[01;33m"), // Warn   yellow
	newBrush("\033[22;31m"), // Error  red
	newBrush("\033[22;35m"), // Painc  magenta
}

type ConsoleLogAdapter struct {
}

// create ConsoleWriter returning as LoggerInterface.
func (this *ConsoleLogAdapter) newLoggerInstance() LoggerInterface {
	cw := &ConsoleWriter{
		lg:     log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds),
		config: ConsoleLogConfig{LogLevel: LevelDebug},
	}
	return cw
}

type ConsoleLogConfig struct {
	LogLevel int `json:"loglevel"`
}

// ConsoleWriter implements LoggerInterface and writes messages to terminal.
type ConsoleWriter struct {
	lg     *log.Logger
	config ConsoleLogConfig
}

// init console logger.
// jsonconfig like '{"loglevel":LevelTrace}'.
func (c *ConsoleWriter) Init(jsonconfig string) error {
	if len(jsonconfig) > 0 {
		err := json.Unmarshal([]byte(jsonconfig), &c.config)
		if err != nil {
			log.Panicln(err.Error())
		}
	}
	return nil
}

func (c *ConsoleWriter) SetLogLevel(loglevel int) {
	c.config.LogLevel = loglevel
}

// write message in console.
func (c *ConsoleWriter) WriteMsg(msg string, level int) error {
	if level < c.config.LogLevel {
		return nil
	}
	if goos := runtime.GOOS; goos == "windows" {
		c.lg.Println(msg)
		return nil
	}
	c.lg.Println(colors[level](msg))
	return nil
}

// implementing method. empty.
func (c *ConsoleWriter) Close() {

}

func init() {
	Register(CONSOLE_PROTOCOL, &ConsoleLogAdapter{})
}
