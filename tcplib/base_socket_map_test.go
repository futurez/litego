package tcplib

import (
	"github.com/futurez/litego/logger"
)

type ServerManager struct {
	tcplib.BaseSocketMap
}

var SP_SvrMgr = NewServerManager()

func NewServerManager() *ServerManager {
	return &ServerManager{}
}

func (s *ServerManager) AcceptClient(chan<- *tcplib.Packet, chan<- bool) (tcplib.Socketer, error) {
	logger.Panic("AcceptClient : not support this function.")
	return nil, nil
}
