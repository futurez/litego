package logger

import (
	"log"
	"os"
	"runtime"
)

type Brush func(string) string

func NewBrush(color string) Brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []Brush{
	NewBrush("1;36"), // Fatal	    cyan
	NewBrush("1;35"), // Painc      magenta
	NewBrush("1;31"), // Error      red
	NewBrush("1;33"), // Warn	    yellow
	NewBrush("1;32"), // Info		green
	NewBrush("1;34"), // Debug      blue
}

type ConsoleLogAdapter struct {
}

// create ConsoleWriter returning as LoggerInterface.
func (this *ConsoleLogAdapter) newLoggerInstance() LoggerInterface {
	cw := &ConsoleWriter{
		lg: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds),
	}
	return cw
}

// ConsoleWriter implements LoggerInterface and writes messages to terminal.
type ConsoleWriter struct {
	lg *log.Logger
}

// init console logger.
// jsonconfig like '{"level":LevelTrace}'.
func (c *ConsoleWriter) Init(jsonconfig string) error {
	return nil
}

// write message in console.
func (c *ConsoleWriter) WriteMsg(msg string, level int) error {
	if goos := runtime.GOOS; goos == "windows" {
		c.lg.Println(msg)
		return nil
	}
	c.lg.Println(colors[level](msg))
	return nil
}

// implementing method. empty.
func (c *ConsoleWriter) Destroy() {

}

// implementing method. empty.
func (c *ConsoleWriter) Flush() {

}

func init() {
	Register(CONSOLE_PROTOCOL_LOG, &ConsoleLogAdapter{})
}
