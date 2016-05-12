package httplib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/futurez/litego/logger"
)

type ServerHandler func(w http.ResponseWriter, r *http.Request)

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
	config   Config
	handlers map[string]ServerHandler
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
		config:   cfg,
		handlers: make(map[string]ServerHandler),
	}
}

func (hs *HttpServer) HandleFunc(pattern string, handler ServerHandler) {
	hs.handlers[pattern] = handler
}

func (hs *HttpServer) ListenAndServe() {
	addr := fmt.Sprintf("%s:%d", hs.config.Host, hs.config.Port)
	logger.Info("Start listen ", addr)
	go func() {
		server := &http.Server{
			Addr:           addr,
			Handler:        hs,
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

func (hs *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if handler, ok := hs.handlers[path]; ok {
		start := time.Now()
		addr := r.Header.Get("X-Real-IP")
		if addr == "" {
			addr = r.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = r.RemoteAddr
			}
		}
		logger.Infof("Start %s %s for %s", r.Method, r.URL.Path, addr)
		handler(w, r)
		logger.Infof("End %s %s for %s in %v\n", r.Method, r.URL.Path, addr, time.Since(start))
	} else {
		hs.serveNotFound(w, r)
	}
}

func (this *HttpServer) serveNotFound(w http.ResponseWriter, r *http.Request) {
	logger.Error("serveNotFound", r.Method, r.RequestURI, http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error":"serivce not found"}`))
}

func HttpResponse(w http.ResponseWriter, code int, contentType string, respData []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	logger.Info("HttpResponse : resp=", string(respData))
	w.Write(respData)
}

func HttpResponseJson(w http.ResponseWriter, code int, resp interface{}) {
	w.Header().Set("Accept", "application/json")
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(code)
	jsonBytes, _ := json.Marshal(resp)
	logger.Debug("HttpResponseJson : resp=", string(jsonBytes))
	w.Write(jsonBytes)
}

func HttpResponseImage(w http.ResponseWriter, picData []byte) {
	w.Header().Set("Content-Type", "image")
	w.WriteHeader(http.StatusOK)
	w.Write(picData)
}

func MakeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		addr := r.Header.Get("X-Real-IP")
		if addr == "" {
			addr = r.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = r.RemoteAddr
			}
		}

		logger.Infof("=>Start %s %s for %s", r.Method, r.URL.Path, addr)
		fn(w, r)
		logger.Infof("=>Finish %s %s for %s in %v\n", r.Method, r.URL.Path, addr, time.Since(start))
	}
}
