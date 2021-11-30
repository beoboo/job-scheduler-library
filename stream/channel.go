package stream

type Channel int

const (
	Output Channel = 1
	Error  Channel = 2
)

func (c Channel) String() string {
	switch c {
	case Output:
		return "output"
	case Error:
		return "error"
	}

	return "undefined"
}
