package resorcerer

import (
	"fmt"
	"github.com/vektra/resorcerer/procstats"
	"github.com/vektra/resorcerer/upstart"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Work struct {
	job       *upstart.Job
	pid       procstats.Pid
	srv       *Service
	memMetric *Metric
}

var ErrReload = fmt.Errorf("Reload configuration")

var Debug bool = false

const defaultPollSeconds = 5
const defaultPollSamples = 5

func RunLoop(u *upstart.Conn, c *Config) error {
	if c.Poll.Seconds == 0 {
		c.Poll.Seconds = defaultPollSeconds
	}

	if c.Poll.Samples == 0 {
		c.Poll.Samples = defaultPollSamples
	}

	if c.Poll.Significant == 0 {
		c.Poll.Significant = (c.Poll.Samples / 2) + 1
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGHUP)

	e := NewEventDispatcher(c)

	sm := make(ServiceMetrics)

	var work []*Work

	for _, s := range c.Services {
		j, err := u.Job(s.Name)
		if err != nil {
			return err
		}

		s.action = j

		e.Dispatch(&Event{"monitor/start", s, nil})

		w := &Work{
			job:       j,
			srv:       s,
			memMetric: sm.Add(s, "memory/limit", c.Poll.Samples),
		}

		w.memMetric.Significant = c.Poll.Significant

		if mem := w.srv.Memory; mem != "" {
			bytes, err := mem.Bytes()
			if err != nil {
				continue
			}

			w.memMetric.Limit = procstats.Bytes(bytes)
		}

		work = append(work, w)
	}

	for {
		forest, err := procstats.DiscoverForest()
		if err != nil {
			return err
		}

		for _, w := range work {
			rpid, err := w.job.Pid()
			if err != nil {
				if w.pid == -1 {
					continue
				}

				w.pid = -1

				e.Dispatch(&Event{"monitor/down", w.srv, nil})
				continue
			}

			pid := procstats.Pid(rpid)

			if w.pid == -1 {
				e.Dispatch(&Event{"monitor/up", w.srv, nil})
			} else if w.pid != pid {
				e.Dispatch(&Event{"monitor/pid-change", w.srv, pid})
				w.memMetric.Reset()
			}

			w.pid = pid

			if gs, ok := forest.Processes[pid]; ok {
				rss := gs.TotalRSS()
				e.Dispatch(&Event{"memory/measured", w.srv, rss})
				w.memMetric.Add(e, rss)
			}
		}

		select {
		case <-sig:
			signal.Stop(sig)
			return ErrReload
		case <-time.After(time.Duration(c.Poll.Seconds) * time.Second):
			continue
		}
	}

	return nil
}
