package procstats

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var PageSize = os.Getpagesize()

var ErrProcFormat = errors.New("proc file format error")

var states = map[string]ProcessState{
	"R": Run,
	"S": Sleep,
	"D": IO,
	"Z": Zombie,
	"T": Stopped,
	"W": Paging,
}

func (p Pid) Info() (*Info, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", p))
	if err != nil {
		return nil, err
	}

	idx := bytes.Index(data, []byte(")"))
	if idx == -1 {
		return nil, ErrProcFormat
	}

	fields := strings.Fields(string(data[idx+1:]))

	i := &Info{
		Pid:   p,
		State: states[fields[0]],
	}

	if n, err := strconv.Atoi(fields[1]); err != nil {
		return nil, err
	} else {
		i.ParentPid = Pid(n)
	}

	if n, err := strconv.Atoi(fields[2]); err != nil {
		return nil, err
	} else {
		i.ProcessGrp = n
	}

	data, err = ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", p))
	if err == nil {
		i.CmdLine = strings.TrimSpace(string(bytes.Join(bytes.Split(data, []byte{0}), []byte(" "))))
	}

	// Get the memory stats from status because, weirdly, they're better.
	// This is what `ps` does too.

	data, err = ioutil.ReadFile(fmt.Sprintf("/proc/%d/status", p))

	var fin int

	idx = bytes.Index(data, []byte("VmSize:"))
	if idx == -1 {
		goto Done
	}

	fin = bytes.Index(data[idx:], []byte("kB"))
	if fin == -1 {
		goto Done
	}

	if n, err := strconv.Atoi(strings.TrimSpace(string(data[idx+8 : idx+fin]))); err != nil {
		goto Done
	} else {
		i.VirtMem = Bytes(n)
	}

	idx = bytes.Index(data, []byte("VmRSS:"))
	if idx == -1 {
		goto Done
	}

	fin = bytes.Index(data[idx:], []byte("kB"))
	if fin == -1 {
		goto Done
	}

	if n, err := strconv.Atoi(strings.TrimSpace(string(data[idx+7 : idx+fin]))); err != nil {
		return nil, ErrProcFormat
	} else {
		i.RSS = Bytes(n * 1024)
	}

Done:
	return i, nil
}
