package logger

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type TcpLogAdapter struct {
}

func (this *TcpLogAdapter) newLoggerInstance() LoggerInterface {
	tw := &TcpLogWriter{}
	tw.lg = log.New(tw, "", (log.Ldate | log.Ltime | log.Lmicroseconds))
	return tw

}

type TcpLogConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type TcpLogWriter struct {
	lg     *log.Logger
	conn   net.Conn
	addr   string
	config TcpLogConfig
}

func (this *TcpLogWriter) Write(b []byte) (int, error) {
	var writeBuf bytes.Buffer
	buflen := uint16(len(b) - 1) //'\n'
	if buflen == 0 {
		return 0, nil
	}
	binary.Write(&writeBuf, binary.LittleEndian, buflen)
	binary.Write(&writeBuf, binary.LittleEndian, b[0:buflen])

	if this.conn == nil {
		if err := this.connect(); err != nil {
			return 0, err
		}
	}

	n, err := this.conn.Write(writeBuf.Bytes())
	if err != nil {
		this.conn.Close()
		this.conn = nil
	}
	return n, err
}

func (this *TcpLogWriter) connect() error {
	if this.conn != nil {
		this.conn.Close()
	}
	var err error
	this.conn, err = net.DialTimeout("tcp", this.addr, 5*time.Second)
	if err != nil {
		fmt.Printf("TcpLog : connect %s failed, %s\n", this.addr, err.Error())
		return err
	}
	return nil
}

func (this *TcpLogWriter) Init(config string) error {
	err := json.Unmarshal([]byte(config), &this.config)
	if err != nil {
		fmt.Printf("TcpLog : unmarshal json config failed, %s\n", err.Error())
		return err
	}
	this.addr = fmt.Sprintf("%s:%d", this.config.Host, this.config.Port)
	return this.connect()
}

func (this *TcpLogWriter) WriteMsg(msg string, level int) error {
	this.lg.Print(msg)
	return nil
}

func (this *TcpLogWriter) Destroy() {
	this.conn.Close()
	this.conn = nil
}

func (this *TcpLogWriter) Flush() {

}

func init() {
	Register(TCP_PROTOCOL_LOG, &TcpLogAdapter{})
}
