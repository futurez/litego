package monitor

import (
	"base/httplib"
	"base/logger"
	"io"
	"net/http"
)

var (
	monitor *httplib.HttpServer
)

func Init(port uint16) {
	if port == 0 {
		port = 54438
	}

	cfg := httplib.Config{
		Host: "0.0.0.0",
		Port: port,
	}
	monitor := httplib.NewServer(cfg)
	monitor.HandleFunc("/SetLogLevel", setLogLevel)
	monitor.ListenAndServe()
}

func AddManageFunc(name string, f func(w http.ResponseWriter, r *http.Request)) {
	if monitor != nil {
		monitor.HandleFunc(name, f)
	}
}

func setLogLevel(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var level int
	var pro string
	switch r.FormValue("level") {
	case "debug":
		level = 0
	case "info":
		level = 1
	case "warn":
		level = 2
	case "error":
		level = 3
	case "panic":
		level = 4
	default:
		level = 5
	}

	switch r.FormValue("type") {
	case "file":
		pro = logger.FILE_PROTOCOL
	case "console":
		pro = logger.CONSOLE_PROTOCOL
	case "":
		pro = logger.ALL_PROTOCOL
	}
	if level > 4 {
		io.WriteString(w, string("Set Log Level Faild"))
	}
	logger.Debug(pro, " log set ", level)
	logger.SetLogLevel(pro, level)
	io.WriteString(w, string("Set Log Level Success"))
}
