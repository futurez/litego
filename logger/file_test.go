package logger

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	lg := NewLogger(10000)
	var config FileLogConfig
	config.FileName = "test"
	config.LogFlag = log.Ldate | log.Ltime | log.Lmicroseconds
	config.MaxSize = 1 << 30
	config.MaxDays = 7
	confbuf, _ := json.Marshal(config)
	lg.SetLogger(FILE_PROTOCOL_LOG, string(confbuf))
	lg.SetEnableFuncCall(true)
	lg.SetFuncCallDepth(2)

	lg.Normal("Normal")
	lg.Debug("DEBUG")
	lg.Warn("Warn")
	lg.Error("Error")

	time.Sleep(time.Second * 1)
	f, err := os.Open("test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		lg.Close()
		f.Close()
		os.Remove("test")
	}()

	b := bufio.NewReader(f)
	linenum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			linenum++
		}
	}
	if linenum != 4 {
		t.Fatal(linenum, "not 4 lines")
	}
}
