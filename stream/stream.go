package stream

import (
	"github.com/beoboo/job-scheduler/library/log"
	"io"
	"sync"
	"time"
)

type Stream struct {
	lines  Lines
	pos    int
	close  chan bool
	closed bool
	m      sync.RWMutex
}

// New creates a new Stream.
func New() *Stream {
	s := &Stream{
		lines: Lines{},
		close: make(chan bool, 1),
	}

	return s
}

// Read returns a channel of available Lines, if it's not been read, or blocks until the next one is written.
func (s *Stream) Read() <-chan *Line {
	next := make(chan *Line)
	pos := 0

	go func() {
		for {
			if s.hasData(pos) {
				next <- s.readNext(pos)
				pos += 1
			} else {
				select {
				case <-time.After(100 * time.Millisecond):
					continue
				case <-s.close:
					close(next)
				}
				break
			}
		}
	}()

	return next
}

// Write adds a new Line, or returns io.ErrClosedPipe if the stream is closed.
func (s *Stream) Write(line Line) error {
	if s.IsClosed() {
		return io.ErrClosedPipe
	}

	s.wlock("Write")
	defer s.wunlock("Write")

	s.lines = append(s.lines, line)

	return nil
}

// IsClosed returns if the stream is closed.
func (s *Stream) IsClosed() bool {
	s.rlock("IsClosed")
	defer s.runlock("IsClosed")

	return s.closed
}

// Close closes the stream, so that no new writes can be added to it.
func (s *Stream) Close() {
	s.wlock("Close")
	defer s.wunlock("Close")

	if s.closed {
		return
	}

	close(s.close)

	s.closed = true
}

func (s *Stream) hasData(pos int) bool {
	s.rlock("hasData")
	defer s.runlock("hasData")

	return pos < len(s.lines)
}

func (s *Stream) readNext(pos int) *Line {
	s.wlock("readNext")
	defer s.wunlock("readNext")

	return &s.lines[pos]
}

// These wraps mutex R/W lock and unlock for debugging purposes
func (s *Stream) rlock(id string) {
	log.Tracef("Stream read locking %s\n", id)
	s.m.RLock()
}

func (s *Stream) runlock(id string) {
	log.Tracef("Stream read unlocking %s\n", id)
	s.m.RUnlock()
}

func (s *Stream) wlock(id string) {
	log.Tracef("Stream write locking %s\n", id)
	s.m.Lock()
}

func (s *Stream) wunlock(id string) {
	log.Tracef("Stream write unlocking %s\n", id)
	s.m.Unlock()
}
