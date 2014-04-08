package procstats

import (
	"fmt"
)

type Bytes int
type Pid int

func (b Bytes) String() string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%dB", b)
	case b < 1024*1024:
		return fmt.Sprintf("%dKB", b/1024)
	case b < 1024*1024*1024:
		return fmt.Sprintf("%dMB", b/(1024*1024))
	default:
		return fmt.Sprintf("%dGB", b/(1024*1024*1024))
	}
}
