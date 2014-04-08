package resorcerer

import (
	"github.com/vektra/resorcerer/procstats"
)

type Metric struct {
	Name        string
	Service     *Service
	Samples     *BytesSamples
	Significant int

	Errored bool
	Limit   procstats.Bytes
}

func (m *Metric) Add(e *EventDispatcher, b procstats.Bytes) {
	m.Samples.Add(b)
	m.Check(e)
}

func (m *Metric) Reset() {
	m.Samples.Reset()
	m.Errored = false
}

func (m *Metric) Check(e *EventDispatcher) {
	if m.Limit == 0 {
		return
	}

	if m.Errored {
		over := m.Samples.CountOver(m.Limit)

		if over == 0 {
			m.Errored = false
			e.Dispatch(&Event{m.Name + "/recover", m.Service, m.Samples.Median()})
		}
	} else {
		over := m.Samples.CountOver(m.Limit)

		if over >= m.Significant {
			m.Errored = true
			e.Dispatch(&Event{m.Name + "/over", m.Service, m.Samples.Median()})
		}
	}
}

type Metrics map[string]*Metric

type ServiceMetrics map[*Service]Metrics

func (sm ServiceMetrics) Add(s *Service, metric string, samples int) *Metric {
	ms, ok := sm[s]
	if !ok {
		ms = make(map[string]*Metric)
		sm[s] = ms
	}

	m, ok := ms[metric]
	if !ok {
		m = &Metric{
			Name:    metric,
			Service: s,
			Samples: NewBytesSamples(samples),
		}

		ms[metric] = m
	}

	return m
}
