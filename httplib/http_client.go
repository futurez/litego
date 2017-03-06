package httplib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"base/logger"
)

const (
	METHOD_GET = iota
	METHOD_POST
)

var JsonHeaders = map[string]string{
	"Accept":       "application/json",
	"Content-Type": "application/json;charset=utf-8",
}

var FormHeaders = map[string]string{
	"Content-Type": "application/x-www-form-urlencoded",
}

var MultiFormHeaders = map[string]string{
	"Content-Type": "multipart/form-data",
}

func HttpRequest(url string, method int, headers map[string]string, data []byte) ([]byte, error) {
	if headers == nil {
		headers = FormHeaders
		logger.Warn("Should be set HTTP headers")
	}

	var req *http.Request
	var err error
	switch method {
	case METHOD_GET:
		url = url + "?" + string(data)
		req, err = http.NewRequest("GET", url, nil)
	case METHOD_POST:
		//logger.Info("data =", data)
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
		Timeout: 60 * time.Second,
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

func HttpRequestData(url string, data []byte) ([]byte, error) {
	respData, err := HttpRequest(url, METHOD_POST, JsonHeaders, data)
	if err != nil {
		logger.Warn("HttpRequestJsonData : http request failed.")
		return nil, err
	}
	return respData, nil
}

func HttpRequestJsonData(url string, req interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		logger.Warn("HttpRequestJsonData : encode req to json failed.")
		return nil, err
	}

	respData, err := HttpRequest(url, METHOD_POST, JsonHeaders, jsonBytes)
	if err != nil {
		logger.Warn("HttpRequestJsonData : http request failed.")
		return nil, err
	}
	return respData, nil
}

func HttpRequestJson(url string, req, resp interface{}) error {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		logger.Warn("HttpRequestJson : encode req to json failed.")
		return err
	}

	respData, err := HttpRequest(url, METHOD_POST, JsonHeaders, jsonBytes)
	if err != nil {
		logger.Warn("HttpRequestJson : http request failed.")
		return err
	}

	err = json.Unmarshal(respData, resp)
	if err != nil {
		logger.Warn("HttpRequestJson : decode resp json data failed.")
		return err
	}
	return nil
}

func HttpRequestJsonToken(url string, headers map[string]string, req, resp interface{}) error {
	if headers == nil {
		return HttpRequestJson(url, req, resp)
	}

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		logger.Warn("HttpRequestJson : encode req to json failed.")
		return err
	}

	for k, v := range JsonHeaders {
		headers[k] = v
	}
	//	logger.Debug("HttpRequestJsonToken : req=", string(jsonBytes))
	respData, err := HttpRequest(url, METHOD_POST, headers, jsonBytes)
	if err != nil {
		logger.Warn("HttpRequestJson : http request failed.")
		return err
	}
	//	logger.Debug("HttpRequestJsonToken : resp=", string(respData))

	err = json.Unmarshal(respData, resp)
	if err != nil {
		logger.Warn("HttpRequestJson : decode resp json data failed.")
		return err
	}
	return nil
}
