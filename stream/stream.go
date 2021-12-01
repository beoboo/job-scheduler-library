package stream

import (
	"github.com/beoboo/job-scheduler/library/logsync"
	"io"
	"log"
)

type Stream struct {
	lines     Lines
	pos       int
	dataCh    chan bool
	closeCh   chan bool
	closed    bool
	m         logsync.Mutex
	ml        logsync.Mutex
	listeners int
}

// New creates a new Stream.
func New() *Stream {
	return &Stream{
		lines:   Lines{},
		dataCh:  make(chan bool),
		closeCh: make(chan bool, 1),
		m:       logsync.New("Stream"),
		ml:      logsync.New("Stream listeners"),
	}
}

// Read returns a channel of available Lines, if it's not been read, or blocks until the next one is written.
func (s *Stream) Read() <-chan *Line {
	next := make(chan *Line)
	pos := 0

	s.updateListeners(1)

	go func() {
		for {
			if s.hasData(pos) {
				next <- s.readNext(pos)
				pos += 1
			} else {
				select {
				case <-s.dataCh:
					continue
				case <-s.closeCh:
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

	go s.notifyNewData()

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

	close(s.closeCh)

	s.closed = true
}

func (s *Stream) Unsubscribe() {
	s.updateListeners(-1)
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

func (s *Stream) notifyNewData() {
	s.ml.RLock("updateListeners")
	defer s.ml.RUnlock("updateListeners")
	for i := 0; i < s.listeners; i++ {
		s.dataCh <- true
	}
}

func (s *Stream) updateListeners(num int) {
	s.ml.WLock("updateListeners")
	defer s.ml.WUnlock("updateListeners")

	s.listeners += num

	if s.listeners < 0 {
		log.Fatalf("Cannot have negative listeners")
	}
}
