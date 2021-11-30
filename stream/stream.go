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
	m      sync.Mutex
}

// New creates a new Stream.
func New() *Stream {
	s := &Stream{
		lines: Lines{},
		close: make(chan bool, 1),
	}

	return s
}

// Read return an available Line, if it's not been read, or blocks until the next one is written.
// Returns io.EOF if the stream is closed.
func (s *Stream) Read() (*Line, error) {
	for {
		if s.hasData() {
			return s.readNext(), nil
		} else {
			select {
			case <-time.After(100 * time.Millisecond):
				continue
			case <-s.close:
			}
			break
		}
	}

	return nil, io.EOF
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

// Rewind sets the next read to be the first stored line.
func (s *Stream) Rewind() {
	s.wlock("Rewind")
	defer s.wunlock("Rewind")

	s.pos = 0
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

func (s *Stream) hasData() bool {
	s.rlock("hasData")
	defer s.runlock("hasData")

	return s.pos < len(s.lines)
}

func (s *Stream) readNext() *Line {
	s.wlock("readNext")
	defer s.wunlock("readNext")

	pos := s.pos
	s.pos += 1

	return &s.lines[pos]
}

// These wraps mutex R/W lock and unlock for debugging purposes
func (s *Stream) rlock(id string) {
	log.Tracef("Stream locking %s\n", id)
	s.m.Lock()
}

func (s *Stream) runlock(id string) {
	log.Tracef("Stream unlocking %s\n", id)
	s.m.Unlock()
}

func (s *Stream) wlock(id string) {
	log.Tracef("Stream locking %s\n", id)
	s.m.Lock()
}

func (s *Stream) wunlock(id string) {
	log.Tracef("Stream write unlocking %s\n", id)
	s.m.Unlock()
}
