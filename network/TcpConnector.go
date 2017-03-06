package network

import (
	"litego/logger"
	"litego/mylist"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	STATUS_NULL = iota
	STATUS_CONNECTED
	STATUS_CLOSEING
)

type TcpClient struct {
	send_buff *mylist.MyList
	cond      *sync.Cond
	linkID    uint64
	status    int
	reconnect bool
	conn      *net.TCPConn
	base      BaseSocket
}

func NewTcpClient(b bool) *TcpClient {
	locker := new(sync.Mutex)
	ID = ID + 1
	return &TcpClient{
		send_buff: mylist.NewList("sendbuf"),
		cond:      sync.NewCond(locker),
		linkID:    ID,
		status:    STATUS_NULL,
		reconnect: b,
	}
}

func (c *TcpClient) LinkID() uint64 {
	return c.linkID
}

func (c *TcpClient) Status() int {
	return c.status
}

func (c *TcpClient) Remote() string {
	return c.conn.RemoteAddr().String()
}

func (c *TcpClient) Local() string {
	return c.conn.LocalAddr().String()
}

func (c *TcpClient) ConnectTcp(host string, port uint16, base BaseSocket) error {
	c.base = base
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		logger.Panic(err.Error())
		return err
	}
	go func() {
		for {
			logger.Infof("ConnectTcp : Connect remote %s", addr.String())
			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				logger.Warn(err.Error())
				if !c.reconnect {
					logger.Infof("ConnectTcp : not reconnect, quit remote %s", addr.String())
					break
				}
				time.Sleep(time.Second * 5)
				continue
			}
			conn.SetKeepAlive(true)
			c.conn = conn
			c.status = STATUS_CONNECTED
			ch := make(chan int)
			c.asyncSend(ch)
			<-ch
			c.base.OnConnect(c)
			c.handleTcpClient()
			if !c.reconnect {
				logger.Infof("ConnectTcp : not reconnect, quit remote %s", addr.String())
				break
			} else {
				logger.Infof("ConnectTcp : reconnect remote %s", addr.String())
				time.Sleep(time.Second * 5)
			}
		}
	}()
	return nil
}

func (c *TcpClient) notifyClose() {
	if c.status != STATUS_NULL {
		c.status = STATUS_NULL
		c.conn.Close()
		c.base.OnClose()
		c.send_buff.Clean()
		c.cond.Broadcast()
	}
}

func (c *TcpClient) asyncSend(ch chan int) {
	go func() {
		defer func() {
			c.notifyClose()
		}()

		ch <- 1
		for {
			c.cond.L.Lock()
			c.cond.Wait()
			c.cond.L.Unlock()

			for {
				p := c.send_buff.PopFront()
				if p == nil {
					if c.status == STATUS_CONNECTED {
						break
					} else {
						logger.Info("asyncSend : quit linkid=", c.linkID, " status=", c.status)
						return
					}
				}
				msg, ok := p.(*[]byte)
				if !ok {
					logger.Error("asyncSend: convert msg error")
					continue
				}
				_, err := c.conn.Write(*msg)
				if err != nil {
					logger.Warnf("asyncSend: %s write buffer error.", c.conn.RemoteAddr().String(), err)
					return
				}
			}
		}
	}()
}

func (c *TcpClient) handleTcpClient() {
	defer func() {
		c.notifyClose()
	}()

	buf := make([]byte, 2048)
	var recvBuf []byte
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				logger.Info("handleTcpClient : connection is closed.", c.conn.RemoteAddr().String())
			} else {
				logger.Warn("handleTcpClient : read error: ", err.Error())
			}
			return
		}
		recvBuf = append(recvBuf, buf[:n]...)
		for {
			ok, data := c.base.OnCheckPackage(&recvBuf)
			if ok {
				c.base.OnDataIn(data)
			} else {
				break
			}
		}
	}
}

func (c *TcpClient) Send(data []byte) error {
	logger.Debug("TcpSend : len=", len(data))
	if c.status == STATUS_CONNECTED {
		c.send_buff.PushBack(&data)
		c.cond.Broadcast()
		return nil
	}
	logger.Warn("Send: Linkid=", c.linkID, " status=", c.status)
	return fmt.Errorf("Send: Linkid=", c.linkID, " status=", c.status)
}

func (c *TcpClient) Close() {
	c.status = STATUS_CLOSEING
	c.reconnect = false
	c.cond.Broadcast()
}
