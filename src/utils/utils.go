package utils

import (
	"sync"
)

type IDAllocator struct {
	curAvaId uint32
	ids      map[uint32]struct{}
	maxIds   uint32
	mu       sync.Mutex
}

func NewIDAllocator() *IDAllocator {
	return &IDAllocator{
		curAvaId: 1,
		ids:      make(map[uint32]struct{}),
		maxIds:   65536,
	}
}

func (p *IDAllocator) GetId() uint32 {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id := range p.ids {
		delete(p.ids, id)
		return id
	}

	if p.curAvaId <= p.maxIds {
		id := p.curAvaId
		p.curAvaId++
		return id
	}

	return 0 // 或其他表示无可用 ID 的值
}

func (p *IDAllocator) PutId(id uint32) {
	if id < 1 || id > p.maxIds {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.ids[id] = struct{}{}
}

