package model

import (
	"fmt"
	"litego/logger"
	"litego/network/tcplib"
	"protocol"
)

type ServerConnect struct {
	tcplib.BaseSocket
	serverType int
}

func NewServerConnect(serverType int, host string, port int) *ServerConnect {
	sc := &ServerConnect{}
	sc.serverType = serverType
	sc.Reconnect = true
	sc.HostPort = fmt.Sprintf("%s:%d", host, port)
	logger.Infof("NewServerConnect : connect to %s(%s)", protocol.ServerType[serverType], sc.HostPort)
	tcplib.ConnectTcp(sc.HostPort, sc)
	return sc
}

func (s *ServerConnect) Connected(sendChan chan<- *tcplib.Packet, closeChan chan<- bool) error {
	s.SetSendAndCloseChan(sendChan, closeChan)

	SP_SvrMgr.AddClient(s)
	return nil
}

func (s *ServerConnect) DataIn(msgId uint32, data []byte) error {
	return nil
}

func (s *ServerConnect) Clearup() error {
	SP_SvrMgr.DeleteClient(s.GetLinkID())
	return nil
}
