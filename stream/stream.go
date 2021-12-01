package stream

import (
	"github.com/beoboo/job-scheduler/library/logsync"
	"io"
	"time"
)

type Stream struct {
	lines  Lines
	pos    int
	close  chan bool
	closed bool
	m      logsync.Mutex
}

// New creates a new Stream.
func New() *Stream {
	s := &Stream{
		lines: Lines{},
		close: make(chan bool, 1),
		m:     logsync.New("Stream"),
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
	s.m.WLock("Write")
	defer s.m.WUnlock("Write")

	if s.closed {
		return io.ErrClosedPipe
	}

	s.lines = append(s.lines, line)

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

	close(s.close)

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
