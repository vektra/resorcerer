package procstats

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func TestDiscoverForest(t *testing.T) {
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

	f, err := DiscoverForest()

	if err != nil {
		panic(err)
	}

	out, err := exec.Command("ps", "ax").CombinedOutput()
	if err != nil {
		panic(err)
	}

	ps_procs := len(bytes.Split(out, []byte("\n"))) - 1

	// Total hack, weird slop on the number of processes because we're reading
	// the number at different times and it can change in between.
	if len(f.Processes) < ps_procs-5 {
		t.Fatalf("Didn't capture all the processes: %d != %d", len(f.Processes), ps_procs)
	}

	child := f.Processes[Pid(c.Process.Pid)]
	self := f.Processes[Pid(os.Getpid())]

	if self.TotalRSS() != self.Process.RSS+child.Process.RSS {
		t.Error("didn't calculate the child RSS properly")
	}

}
