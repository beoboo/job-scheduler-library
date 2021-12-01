package stream

type StreamType int

const (
	Output StreamType = 1
	Error  StreamType = 2
)

func (c StreamType) String() string {
	switch c {
	case Output:
		return "output"
	case Error:
		return "error"
	}

	return "undefined"
}
