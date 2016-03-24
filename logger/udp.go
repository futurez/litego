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
	Host     string `json:"host"`
	Port     int    `json:"port"`
	LogLevel int    `json:"loglevel"`
}

type UdpLogWriter struct {
	lg      *log.Logger
	udpAddr *net.UDPAddr
	udpConn *net.UDPConn
	config  UdpLogConfig
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
	err := json.Unmarshal([]byte(jsonconfig), &ulw.config)
	if err != nil {
		log.Panicln(err)
	}

	ulw.udpAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ulw.config.Host, ulw.config.Port))
	if err != nil {
		log.Panicln(err)
	}

	ulw.udpConn, err = net.DialUDP("udp", nil, ulw.udpAddr)
	if err != nil {
		log.Panicln(err)
	}
	return nil
}

func (ulw *UdpLogWriter) SetLogLevel(loglevel int) {
	ulw.config.LogLevel = loglevel
}

func (ulw UdpLogWriter) WriteMsg(msg string, level int) error {
	if level < ulw.config.LogLevel {
		return nil
	}
	ulw.lg.Println(msg)
	return nil
}

func (ulw *UdpLogWriter) Close() {
	ulw.udpConn.Close()
}

func init() {
	Register(UDP_PROTOCOL, &UdpLogAdapter{})
}
