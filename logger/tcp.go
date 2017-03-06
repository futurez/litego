package logger

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type TcpLogAdapter struct {
}

func (adapter TcpLogAdapter) newLoggerInstance() LoggerInterface {
	tlw := &TcpLogWriter{}
	tlw.lg = log.New(tlw, "", (log.Ldate | log.Ltime | log.Lmicroseconds))
	return tlw
}

type TcpLogConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	LogLevel int    `json:"loglevel"`
}

type TcpLogWriter struct {
	lg      *log.Logger
	tcpAddr *net.TCPAddr
	tcpConn *net.TCPConn
	config  TcpLogConfig
}

func (tlw TcpLogWriter) Write(b []byte) (int, error) {
	buflen := int16(len(b))
	if buflen == 0 {
		return 0, nil
	}

	var writeBuf bytes.Buffer
	binary.Write(&writeBuf, binary.LittleEndian, buflen)
	binary.Write(&writeBuf, binary.LittleEndian, b[0:buflen])

	return tlw.write(writeBuf.Bytes())
}

func (tlw *TcpLogWriter) write(b []byte) (int, error) {
	if tlw.tcpConn == nil {
		if err := tlw.connect(); err != nil {
			return 0, err
		}
	}

	n, err := tlw.tcpConn.Write(b)
	if err != nil {
		tlw.Close()
		log.Println(err.Error())
		return 0, err
	}
	return n, nil
}

func (tlw *TcpLogWriter) connect() error {
	tlw.Close()

	var err error
	tlw.tcpConn, err = net.DialTCP("tcp", nil, tlw.tcpAddr)
	if err != nil {
		log.Println(tlw.tcpAddr.Network(), tlw.tcpAddr.String(), err.Error())
		return err
	}
	return tlw.tcpConn.SetKeepAlive(true)
}

func (tlw *TcpLogWriter) Init(jsonconfig string) error {
	err := json.Unmarshal([]byte(jsonconfig), &tlw.config)
	if err != nil {
		log.Panicln(err.Error())
	}

	tlw.tcpAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", tlw.config.Host, tlw.config.Port))
	if err != nil {
		log.Panicln(err.Error())
	}

	return tlw.connect()
}

func (tlw *TcpLogWriter) SetLogLevel(loglevel int) {
	tlw.config.LogLevel = loglevel
}

func (tlw TcpLogWriter) WriteMsg(msg string, level int) error {
	if level < tlw.config.LogLevel {
		return nil
	}
	tlw.lg.Print(msg)
	return nil
}

func (tlw *TcpLogWriter) Close() {
	if tlw.tcpConn != nil {
		tlw.tcpConn.Close()
		tlw.tcpConn = nil
	}
}

func init() {
	Register(TCP_PROTOCOL, &TcpLogAdapter{})
}
