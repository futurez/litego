package httplite

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	METHOD_GET = iota
	METHOD_POST
)

func HttpRequest(url string, method int, headers map[string]string, data []byte) ([]byte, error) {
	var req *http.Request
	var err error
	switch method {
	case METHOD_GET:
		url = url + "?" + string(data)
		req, err = http.NewRequest("GET", url, nil)
	case METHOD_POST:
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	}

	if err != nil {
		return []byte(""), err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// 设置 TimeOut
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return []byte(""), err
	}

	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return []byte(""), err
	}
	return resp_body, nil
}
