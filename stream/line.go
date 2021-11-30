package stream

import (
	"fmt"
	"time"
)

type Line struct {
	Time    time.Time
	Channel Channel
	Text    []byte
}

type Lines = []Line

func (l *Line) String() string {
	return fmt.Sprintf("[%s][%s] %s", l.Time.Format("15:04:05.000"), l.Channel, l.Text)
}
