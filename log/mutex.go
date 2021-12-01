package log

import "sync"

type LogMutex struct {
	m sync.RWMutex
}
