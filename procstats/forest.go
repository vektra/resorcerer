package procstats

import (
	"fmt"
	"io/ioutil"
	"strconv"
)

type GroupStats struct {
	Process  *Info
	ChildRSS Bytes
}

func (g *GroupStats) TotalRSS() Bytes {
	return g.Process.RSS + g.ChildRSS
}

type Forest struct {
	Processes map[Pid]*GroupStats
}

func DiscoverForest() (*Forest, error) {
	f := &Forest{
		Processes: make(map[Pid]*GroupStats),
	}

	fis, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		n, err := strconv.Atoi(fi.Name())
		if err != nil {
			continue
		}

		pid := Pid(n)

		info, err := pid.Info()
		if err != nil {
			fmt.Printf("Unable to load %d, %s\n", n, err)
			continue
		}

		f.Processes[pid] = &GroupStats{Process: info}
	}

	for _, info := range f.Processes {
		if parent, ok := f.Processes[info.Process.ParentPid]; ok {
			parent.ChildRSS += info.Process.RSS
		}
	}

	return f, nil
}
