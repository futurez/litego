// socket_map.go
package tcplib

import (
	"sync"
)

type BaseSocketMap struct {
	sync.RWMutex
	linkIDGen     LinkIDGen
	BaseSocketMap map[LinkID]Socketer
}

func (s *BaseSocketMap) AddClient(socket Socketer) LinkID {
	s.Lock()
	defer s.Unlock()
	linkID := s.linkIDGen.NewID()
	s.BaseSocketMap[linkID] = socket
	socket.SetLinkID(linkID)
	return linkID
}

func (s *BaseSocketMap) DeleteClient(id LinkID) {
	s.Lock()
	defer s.Unlock()
	delete(s.BaseSocketMap, id)
}

func (s *BaseSocketMap) FindClient(id LinkID) (Socketer, bool) {
	s.RLock()
	defer s.RUnlock()
	socket, ok := s.BaseSocketMap[id]
	return socket, ok
}

func (s *BaseSocketMap) Close() {
	//	logger.Warn("ServerMange Close!!!")
	s.RLock()
	defer s.RUnlock()
	for _, socket := range s.BaseSocketMap {
		socket.Close()
	}
}
