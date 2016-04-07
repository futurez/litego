package tcplib

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/futurez/litego/logger"
)

type Socketer interface {
	IsReconnect() bool // if reconnect.

	Connected(chan<- *Packet, chan<- bool) error //send chan, close socket chan

	DataIn(uint32, []byte) error // process receive data.

	Send(uint32, []byte) (int, error)

	Close() //please send true to chan<- bool by connected param

	Clearup() error

	GetLinkID() LinkID

	SetLinkID(LinkID)
}

type Listener interface {
	AcceptClient(chan<- *Packet, chan<- bool) (Socketer, error)

	AddClient(Socketer) LinkID

	DeleteClient(LinkID)

	FindClient(LinkID) (Socketer, bool)

	Close()
}

func handleRead(conn *net.TCPConn, socketer Socketer) {
	buf := make([]byte, 2048)
	var recvBuf []byte
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				logger.Info("Read data End.", conn.RemoteAddr().String())
				socketer.Clearup() //数据应该要被处理的
			} else {
				logger.Warn("Read Error: %s", err.Error())
				socketer.Clearup()
				return
			}
		}
		recvBuf = append(recvBuf, buf[:n]...)
		for {
			id, data, ok := decodePacket(&recvBuf)
			if ok {
				socketer.DataIn(id, data)
			} else {
				break
			}
		}
	}
}

func handleSend(conn *net.TCPConn) (chan<- *Packet, chan<- bool) {
	packetChan := make(chan *Packet, 64)
	closeChan := make(chan bool)

	go func() {
		defer conn.Close()

		for {
			select {
			case packet, ok := <-packetChan:
				{
					if !ok { //close chan
						logger.Warn("close conn when have send all data to remote.")
						return
					}

					buf := encodePacket(packet.MsgId, &packet.Data)
					_, err := conn.Write(buf[:len(buf)])
					if err != nil {
						logger.Warnf("%s write buffer error.", conn.RemoteAddr().String())
						return
					}
				}

			case <-closeChan:
				logger.Info("close this connect.")
				close(packetChan)
			}
		}
	}()
	return packetChan, closeChan
}

func ListenTcp(host string, port int, listener Listener) error {
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		logger.Error(err)
		return err
	}
	listConn, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		logger.Error(err)
		return err
	}
	go func() {
		defer func() {
			logger.Info("Close listen address: %s", listConn.Addr().String())
			listener.Close()
			listConn.Close()
		}()

		logger.Info("listen address: ", listConn.Addr().String())
		for {
			conn, err := listConn.AcceptTCP()
			if err != nil {
				continue
			}
			logger.Debugf("%s accept new connect, remote address: %s.", listConn.Addr().String(), conn.RemoteAddr().String())

			sendChan, closeChan := handleSend(conn)

			client, err := listener.AcceptClient(sendChan, closeChan)
			if err != nil {
				logger.Warnf("appect failed, close remote address: %s.", conn.RemoteAddr().String())
				closeChan <- true
				continue
			}
			go handleRead(conn, client)
		}
	}()
	return nil
}

func ConnectTcp(hostport string, socketer Socketer) error {
	addr, err := net.ResolveTCPAddr("tcp", hostport)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	go func() {
		for {
			logger.Infof("Connect remote %s", addr.String())
			conn, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				logger.Warn(err.Error())
				if !socketer.IsReconnect() {
					break
				}
				time.Sleep(time.Second * 3)
				continue
			}
			conn.SetKeepAlive(true)

			sendChan, closeChan := handleSend(conn)

			socketer.Connected(sendChan, closeChan)

			handleRead(conn, socketer) //if disconnect, return this call.
			if !socketer.IsReconnect() {
				break
			}
			time.Sleep(time.Second * 2)
		}
	}()
	return nil
}
