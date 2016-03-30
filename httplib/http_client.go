package httplib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/futurez/litego/logger"
)

const (
	METHOD_GET = iota
	METHOD_POST
)

var jsonHeaders = map[string]string{
	"Accept":       "application/json",
	"Content-Type": "application/json;charset=utf-8",
}

func HttpRequest(url string, method int, headers map[string]string, data []byte) ([]byte, error) {
	var req *http.Request
	var err error
	switch method {
	case METHOD_GET:
		url = url + "?" + string(data)
		req, err = http.NewRequest("GET", url, nil)
	case METHOD_POST:
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	default:
		logger.Warn("unknown http method = ", method)
		return []byte(""), errors.New("unknown http method")
	}

	if err != nil {
		logger.Warn(err.Error())
		return []byte(""), err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// set TimeOut
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn(err.Error())
		return []byte(""), err
	}

	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warn(err.Error())
		return []byte(""), err
	}
	return resp_body, nil
}

func HttpRequestJson(url string, req interface{}) ([]byte, error) {
	jsonBytes, _ := json.Marshal(req)
	return HttpRequest(url, METHOD_POST, jsonHeaders, jsonBytes)
}

func HttpRequestJsonToken(url string, headers map[string]string, req interface{}) ([]byte, error) {
	if headers == nil {
		return HttpRequestJson(url, req)
	}

	for k, v := range jsonHeaders {
		headers[k] = v
	}
	jsonBytes, _ := json.Marshal(req)
	return HttpRequest(url, METHOD_POST, headers, jsonBytes)
}
