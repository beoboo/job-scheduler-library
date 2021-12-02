package stream

import (
	"fmt"
	"time"
)

type Line struct {
	Time time.Time
	Type StreamType
	Text []byte
}

type Lines = []Line

func (l *Line) String() string {
	return fmt.Sprintf("[%s][%s] %s", l.Time.Format("15:04:05.000"), l.Type, l.Text)
}
