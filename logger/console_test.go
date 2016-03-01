package logger

import (
	"testing"
)

// Try each log level in decreasing order of priority.
func testConsoleCalls(bl *Logger) {

}

// Test console logging by visually comparing the lines being output with and
// without a log level specification.
func TestConsole(t *testing.T) {
	lg := NewLogger(10000)
	lg.SetEnableFuncCall(true)
	lg.SetFuncCallDepth(2)
	lg.SetLogger(CONSOLE_PROTOCOL_LOG, "")
	lg.Error("error")
	lg.Warn("warn")
	lg.Info("info")
	lg.Debug("debug")
}
