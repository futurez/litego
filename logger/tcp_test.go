package logger

import (
	"encoding/json"
	"testing"
	"time"
)

func TestConn(t *testing.T) {
	lg := NewLogger(10000)
	var config TcpLogConfig
	config.Host = "192.168.1.199"
	config.Port = 30000
	confbuf, _ := json.Marshal(config)
	lg.SetLogger(CONSOLE_PROTOCOL_LOG, "")
	lg.SetLogger(TCP_PROTOCOL_LOG, string(confbuf))
	lg.SetEnableFuncCall(true)
	lg.SetFuncCallDepth(2)

	for i := 0; i < 10000; i++ {
		lg.Debug("DEBUG, %d", i)
		lg.Warn("Warn %d", i)
		lg.Error("Error %d", i)
	}
	time.Sleep(time.Second * 4)
	lg.Close()
}
