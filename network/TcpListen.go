package network

import (
	"litego/logger"
	"fmt"
	"net"
)

func ListenTcp(ip string, port uint16, accept BaseListen) error {
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		logger.Error(err)
		panic(err)
	}
	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	go func() {
		defer func() {
			logger.Info("Close listen address: %s", listener.Addr().String())
			listener.Close()
		}()

		logger.Info("listen address: ", listener.Addr().String())
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				continue
			}
			logger.Debugf("%d accept new connect, remote address: %s.", port, conn.RemoteAddr().String())

			pClient := NewTcpClient(false)
			pClient.status = STATUS_CONNECTED
			pClient.conn = conn
			pClient.base = accept.OnAccept(pClient)
			ch := make(chan int)
			pClient.asyncSend(ch)
			<-ch
			go pClient.handleTcpClient()
		}
	}()
	return nil
}
