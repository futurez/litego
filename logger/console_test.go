package logger

import (
	"testing"
)

func TestConsole(t *testing.T) {
	lg := NewLogger(10000)
	lg.SetFuncDepth(2)
	lg.SetPrefix("consoletest")
	lg.SetLogger(CONSOLE_PROTOCOL, "")
	lg.Panic("Panic")
	lg.Error("error")
	lg.Warn("warn")
	lg.Info("info")
	lg.Debug("debug")
	t.Log("console test success.")
}
