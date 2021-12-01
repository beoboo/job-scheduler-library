package stream

import (
	"github.com/beoboo/job-scheduler/library/logsync"
	"io"
	"sync"
)

type Stream struct {
	lines     Lines
	pos       int
	closed    bool
	m         logsync.Mutex
	cond      sync.Cond
	listeners int
}

// New creates a new Stream.
func New() *Stream {
	s := &Stream{
		lines: Lines{},
		m:     logsync.New("Stream"),
	}

	s.cond.L = &s.m
	return s
}

// Read returns a channel of available Lines, if it's not been read, or blocks until the next one is written.
func (s *Stream) Read() <-chan *Line {
	next := make(chan *Line)
	pos := 0

	go func() {
		defer close(next)

		for {
			if s.hasData(pos) {
				next <- s.readNext(pos)
				pos += 1
			} else {
				s.m.Lock()
				s.cond.Wait()

				if s.closed {
					s.m.Unlock()
					break
				}

				s.m.Unlock()
			}
		}
	}()

	return next
}

// Write adds a new Line, or returns io.ErrClosedPipe if the stream is closed.
func (s *Stream) Write(line Line) error {
	s.m.WLock("Write")
	defer s.m.WUnlock("Write")

	if s.closed {
		return io.ErrClosedPipe
	}

	s.lines = append(s.lines, line)

	s.cond.Broadcast()

	return nil
}

// IsClosed returns if the stream is closed.
func (s *Stream) IsClosed() bool {
	s.m.RLock("IsClosed")
	defer s.m.RUnlock("IsClosed")

	return s.closed
}

// Close closes the stream, so that no new writes can be added to it.
func (s *Stream) Close() {
	s.m.WLock("Close")
	defer s.m.WUnlock("Close")

	if s.closed {
		return
	}
	s.cond.Broadcast()

	s.closed = true
}

func (s *Stream) hasData(pos int) bool {
	s.m.RLock("hasData")
	defer s.m.RUnlock("hasData")

	return pos < len(s.lines)
}

func (s *Stream) readNext(pos int) *Line {
	s.m.RLock("readNext")
	defer s.m.RUnlock("readNext")

	return &s.lines[pos]
}
