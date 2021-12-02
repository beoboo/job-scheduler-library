package logsync

import (
	"github.com/beoboo/job-scheduler/library/log"
	"sync"
)

// Mutex These wraps sync.RWMutexlock and unlock for debugging purposes
type Mutex struct {
	parent string
	m      sync.RWMutex
}

func New(parent string) Mutex {
	return Mutex{
		parent: parent,
	}
}

func (m *Mutex) Lock() {
	m.m.Lock()
}

func (m *Mutex) Unlock() {
	m.m.Unlock()
}

func (m *Mutex) RLock(fn string) {
	log.Tracef("%s read locking %s\n", m.parent, fn)
	m.m.RLock()
}

func (m *Mutex) RUnlock(fn string) {
	log.Tracef("%s read unlocking %s\n", m.parent, fn)
	m.m.RUnlock()
}

func (m *Mutex) WLock(fn string) {
	log.Tracef("%s write locking %s\n", m.parent, fn)
	m.m.Lock()
}

func (m *Mutex) WUnlock(fn string) {
	log.Tracef("%s write unlocking %s\n", m.parent, fn)
	m.m.Unlock()
}
