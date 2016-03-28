package tcplib

import (
	"errors"
	"litego/logger"
)

type BaseSocket struct {
	Reconnect bool //true : is accept connect object.
	HostPort  string
	linkid    LinkID
	sendChan  chan<- *Packet
	closeChan chan<- bool
}

func (b *BaseSocket) IsReconnect() bool {
	return b.Reconnect
}

func (b *BaseSocket) SetSendAndCloseChan(sendChan chan<- *Packet, closeChan chan<- bool) {
	b.sendChan = sendChan
	b.closeChan = closeChan
}

func (b *BaseSocket) Close() {
	b.Reconnect = false
	if b.closeChan != nil {
		b.closeChan <- true
		b.closeChan = nil
	}
}

func (b *BaseSocket) GetLinkID() LinkID {
	return b.linkid
}

func (b *BaseSocket) SetLinkID(id LinkID) {
	b.linkid = id
}

func (b *BaseSocket) Send(id uint32, buf []byte) (int, error) {
	logger.Debugf("%d Send msgid %d, len %d ", b.linkid, id, len(buf))
	if len(buf) <= 0 {
		return 0, errors.New("send buffer is zero or conn is closed.")
	}
	b.sendChan <- &Packet{id, buf}
	return len(buf), nil
}
