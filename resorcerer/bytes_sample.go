package resorcerer

import (
	"github.com/vektra/resorcerer/procstats"
	"sort"
)

type BytesSamples struct {
	pos   int
	slice []procstats.Bytes
}

func NewBytesSamples(c int) *BytesSamples {
	bs := &BytesSamples{
		pos:   0,
		slice: make([]procstats.Bytes, c),
	}

	return bs
}

func (bs *BytesSamples) Reset() {
	for i := 0; i < len(bs.slice); i++ {
		bs.slice[i] = 0
	}

	bs.pos = 0
}

func (bs *BytesSamples) Add(s procstats.Bytes) {
	bs.slice[bs.pos] = s

	bs.pos++
	if bs.pos >= len(bs.slice) {
		bs.pos = 0
	}
}

func (bs *BytesSamples) CountOver(thresh procstats.Bytes) int {
	count := 0

	for _, s := range bs.slice {
		if s >= thresh {
			count++
		}
	}

	return count
}

func (bs *BytesSamples) Average() procstats.Bytes {
	bytes := procstats.Bytes(0)

	valid := 0

	for _, s := range bs.slice {
		if s > 0 {
			bytes += s
			valid++
		}
	}

	return bytes / procstats.Bytes(valid)
}

func (bs *BytesSamples) Median() procstats.Bytes {
	is := make([]int, len(bs.slice))

	for idx, val := range bs.slice {
		is[idx] = int(val)
	}

	sort.Ints(is)

	return procstats.Bytes(is[len(is)/2])
}
