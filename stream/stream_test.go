package stream

import (
	"sync"
	"testing"
	"time"
)

func TestStreamNew(t *testing.T) {
	s := New()
	defer s.Close()

	if s.IsClosed() {
		t.Fatalf("Stream should be open")
	}
}

func TestStreamWrite(t *testing.T) {
	s := New()
	defer s.Close()

	expected := "line"
	println("1")
	write(s, expected)
	println("2")

	assertRead(t, s, expected)
	println("3")
}

func TestStreamResetPos(t *testing.T) {
	s := New()
	defer s.Close()

	expected := "line"
	write(s, expected)

	assertRead(t, s, expected)

	assertRead(t, s, expected)
}

func TestStreamAllLines(t *testing.T) {
	s := New()
	defer s.Close()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		_ = s.Write(buildLine("line1"))
		wg.Done()
	}()

	// This is required to avoid write goroutines to overlap (could be solved in other ways too)
	time.Sleep(100 * time.Millisecond)

	go func() {
		_ = s.Write(buildLine("line2"))
		wg.Done()
	}()

	wg.Wait()

	l := s.Read()

	assertLine(t, <-l, "line1")
	assertLine(t, <-l, "line2")
}

func TestStreamCannotWriteToClosedStream(t *testing.T) {
	s := New()
	s.Close()

	errCh := make(chan error, 1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		err := s.Write(buildLine("invalid"))

		errCh <- err
	}()

	if <-errCh == nil {
		t.Fatal("Should not be able to write if stream is closed")
	}
}

func TestStreamConcurrentReads(t *testing.T) {
	s := New()
	res1 := ""
	res2 := ""

	var wg sync.WaitGroup
	wg.Add(4)

	// Starts writing delayed
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = s.Write(buildLine("#1"))
		time.Sleep(100 * time.Millisecond)
		_ = s.Write(buildLine("#2"))
		wg.Done()
	}()

	// Starts before the writes
	go func() {
		for l := range s.Read() {
			res1 += string(l.Text)
		}
		wg.Done()
	}()

	// Starts after the first write, but before the second
	go func() {
		time.Sleep(50 * time.Millisecond)
		for l := range s.Read() {
			res2 += string(l.Text)
		}
		wg.Done()
	}()

	// Closes the channel
	go func() {
		// Starts after the first write, but before the second
		time.Sleep(400 * time.Millisecond)
		s.Close()
		wg.Done()
	}()

	wg.Wait()

	if res1 != res2 {
		t.Fatalf("Expected outputs to be equal: %s != %s", res1, res2)
	}
}

func write(s *Stream, t string) {
	go func() {
		_ = s.Write(buildLine(t))
	}()
}

func buildLine(t string) Line {
	return Line{
		Text: []byte(t),
		Time: time.Now(),
	}
}

func assertRead(t *testing.T, s *Stream, expected string) {
	l := <-s.Read()

	if string(l.Text) != expected {
		t.Fatalf("Didn't read successfully, expected \"%s\", got \"%s\"", expected, l.Text)
	}
}

func assertLine(t *testing.T, l *Line, expected string) {
	if string(l.Text) != expected {
		t.Fatalf("Didn't read successfully, expected \"%s\", got \"%s\"", expected, l.Text)
	}
}
