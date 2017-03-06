package network

import (
	"sync"
)

type ClientManager struct {
	sync.Mutex
	Map map[uint64]interface{}
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		Map: make(map[uint64]interface{}),
	}
}

func (this *ClientManager) AddClient(linkID uint64, base interface{}) {
	this.Lock()
	defer this.Unlock()
	this.Map[linkID] = base
}

func (this *ClientManager) DelClient(linkID uint64) {
	this.Lock()
	defer this.Unlock()
	delete(this.Map, linkID)
}

func (this *ClientManager) FindClient(linkID uint64) interface{} {
	this.Lock()
	defer this.Unlock()
	v, ok := this.Map[linkID]
	if !ok {
		return nil
	}
	return v
}

func (this *ClientManager) Clear() {
	this.Lock()
	defer this.Unlock()
	this.Map = make(map[uint64]interface{})
}

func (this *ClientManager) Count() int {
	this.Lock()
	defer this.Unlock()
	return len(this.Map)
}
