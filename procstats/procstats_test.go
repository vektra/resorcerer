package procstats

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestInfo(t *testing.T) {
	c := exec.Command("cat")
	x, err := c.StdinPipe()

	if err != nil {
		panic(err)
	}

	err = c.Start()
	if err != nil {
		panic(err)
	}

	defer c.Wait()
	defer x.Close()

	pid := Pid(c.Process.Pid)

	out, err := exec.Command("ps", "-p", strconv.Itoa(int(pid)), "v").CombinedOutput()
	if err != nil {
		panic(err)
	}

	info, err := pid.Info()
	if err != nil {
		panic(err)
	}

	if info.ParentPid != Pid(os.Getpid()) {
		t.Errorf("Did not decode parent properly")
	}

	lines := bytes.Split(out, []byte("\n"))

	fields := bytes.Fields(lines[1])

	rss, err := strconv.Atoi(string(fields[7]))
	if err != nil {
		panic(err)
	}

	rss = rss * 1024

	if Bytes(rss) != info.RSS {
		t.Errorf("did not decode RSS properly: %d != %d", rss, info.RSS)
	}
}
