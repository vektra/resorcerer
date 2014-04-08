package resorcerer

import (
	"fmt"
	"os"
	"strings"
)

type Event struct {
	Name    string
	Service *Service
	Value   interface{}
}

type EventDispatcher struct {
	Registry map[*Service]map[string][]*Handler
	Global   map[string][]*Handler
	Actions  []Action
	Debug    bool
}

func NewEventDispatcher(c *Config) *EventDispatcher {
	ev := &EventDispatcher{
		Registry: make(map[*Service]map[string][]*Handler),
		Global:   make(map[string][]*Handler),
	}

	ev.Actions = []Action{
		&processAction{},
		&emailAction{&c.Email},
	}

	for _, s := range c.Services {
		for _, h := range s.Handlers {
			ev.Add(s, h)
		}
	}

	for _, h := range c.Handlers {
		ev.Global[h.Event] = append(ev.Global[h.Event], h)
	}

	return ev
}

func (e *EventDispatcher) Add(s *Service, h *Handler) {
	p := e.Registry[s]
	if p == nil {
		p = make(map[string][]*Handler)
		e.Registry[s] = p
	}

	p[h.Event] = append(p[h.Event], h)
}

func (e *EventDispatcher) Dispatch(ev *Event) error {
	if e.Debug {
		fmt.Printf("Dispatching event '%s' for '%s': %v\n", ev.Name, ev.Service.Name, ev.Value)
	}

	parts := strings.Split(ev.Name, "/")

	regs := []map[string][]*Handler{e.Global}

	if s, ok := e.Registry[ev.Service]; ok {
		regs = append(regs, s)
	}

	for _, s := range regs {
		cur := ""

		for _, part := range parts {
			if cur == "" {
				cur = part
			} else {
				cur = cur + "/" + part
			}

			if h, ok := s[cur]; ok {
				err := e.Process(ev, h)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (e *EventDispatcher) Process(ev *Event, hs []*Handler) error {
	for _, h := range hs {
		if e.Debug {
			fmt.Printf("Handling event '%s' for '%s': %v => %#v\n", ev.Name, ev.Service.Name, ev.Value, h)
		}

		for _, a := range e.Actions {
			err := a.Run(ev, h)
			if err != nil && err != ErrNotConfigured {
				fmt.Fprintf(os.Stderr, "Error executing action: %s\n", err)
			}
		}
	}
	return nil
}
