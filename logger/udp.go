package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type UdpLogAdapter struct {
}

func (adapter UdpLogAdapter) newLoggerInstance() LoggerInterface {
	ulw := &UdpLogWriter{}
	ulw.lg = log.New(ulw, "", (log.Ldate | log.Ltime | log.Lmicroseconds))
	return ulw
}

type UdpLogConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type UdpLogWriter struct {
	lg      *log.Logger
	udpAddr *net.UDPAddr
	udpConn *net.UDPConn
}

func (ulw UdpLogWriter) Write(b []byte) (int, error) {
	buflen := len(b)
	if buflen <= 0 {
		return 0, nil
	}
	for sendLen := buflen; buflen > 0; buflen -= sendLen {
		if sendLen > 512 {
			sendLen = 512
		}
		ulw.udpConn.Write(b[0:sendLen])
	}
	return len(b), nil
}

func (ulw *UdpLogWriter) Init(jsonconfig string) error {
	var config UdpLogConfig
	err := json.Unmarshal([]byte(jsonconfig), &config)
	if err != nil {
		log.Panicln(err)
	}

	ulw.udpAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		log.Panicln(err)
	}

	ulw.udpConn, err = net.DialUDP("udp", nil, ulw.udpAddr)
	if err != nil {
		log.Panicln(err)
	}
	return nil
}

func (ulw UdpLogWriter) WriteMsg(msg string, level int) error {
	ulw.lg.Println(msg)
	return nil
}

func (ulw *UdpLogWriter) Close() {
	ulw.udpConn.Close()
}

func init() {
	Register(UDP_PROTOCOL, &UdpLogAdapter{})
}
