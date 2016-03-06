package logger

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
)

type brush func(string) string

func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + text + reset
	}
}

var colors = []brush{
	newBrush("1;35m"), // Painc     magenta
	newBrush("1;31m"), // Error     red
	newBrush("1;33m"), // Warn	    yellow
	newBrush("1;32m"), // Info		green
	newBrush("1;34m"), // Debug     blue
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

// write message in console.
func (c *ConsoleWriter) WriteMsg(msg string, level int) error {
	if level > c.config.LogLevel {
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
