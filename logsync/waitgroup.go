package logsync

import (
	"github.com/beoboo/job-scheduler/library/log"
	"sync"
)

// WaitGroup wraps sync.WaitGroup for debugging purposes
type WaitGroup struct {
	parent string
	wg     sync.WaitGroup
}

func NewWaitGroup(parent string) WaitGroup {
	return WaitGroup{
		parent: parent,
	}
}

func (wg *WaitGroup) Add(id string, delta int) {
	log.Tracef("[%s] Adding %d %s\n", wg.parent, delta, id)
	wg.wg.Add(delta)
}

func (wg *WaitGroup) Done(id string) {
	log.Tracef("[%s] Done %s\n", wg.parent, id)
	wg.wg.Done()
}

func (wg *WaitGroup) Wait() {
	log.Tracef("[%s] Waiting\n", wg.parent)
	wg.wg.Wait()
}
