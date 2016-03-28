package httplite

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/futurez/litego/logger"
)

//http server config
type Config struct {
	Host           string
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
}

//type FuncHandler func(w http.ResponseWriter, req *http.Request)

type HttpServer struct {
	config     Config
	handlerMux *http.ServeMux
}

func NewServer(cfg Config) *HttpServer {
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	if cfg.MaxHeaderBytes == 0 {
		cfg.MaxHeaderBytes = http.DefaultMaxHeaderBytes
	}
	return &HttpServer{
		config:     cfg,
		handlerMux: http.NewServeMux()}
}

func (hs *HttpServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	hs.handlerMux.HandleFunc(pattern, handler)
}

func (hs *HttpServer) ListenAndServe() {
	addr := fmt.Sprintf("%s:%d", hs.config.Host, hs.config.Port)
	logger.Info("Start listen ", addr)
	go func() {
		server := &http.Server{
			Addr:           addr,
			Handler:        hs.handlerMux,
			ReadTimeout:    time.Duration(hs.config.ReadTimeout),
			WriteTimeout:   time.Duration(hs.config.WriteTimeout),
			MaxHeaderBytes: hs.config.MaxHeaderBytes}

		err := server.ListenAndServe()
		if err != nil {
			logger.Panic(err.Error)
			os.Exit(1)
		}
	}()
}

func WriteResponseJson(w http.ResponseWriter, code int, resp interface{}) {
	w.Header().Set("Accept", "application/json")
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(code)
	jsonBytes, _ := json.Marshal(resp)
	w.Write(jsonBytes)
}

func WriteResponse(w http.ResponseWriter, code int, contentType string, respData []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	w.Write(respData)
}
