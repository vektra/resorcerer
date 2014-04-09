package resorcerer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Event struct {
	Name    string
	Service *Service
	Value   interface{}
}

type ExternalEvent struct {
	Name     string      `json:"name"`
	Service  string      `json:"service"`
	Hostname string      `json:"hostname"`
	Value    interface{} `json:"value"`
}

func (e *Event) ToJson() ([]byte, error) {
	ee := ExternalEvent{
		Name:     e.Name,
		Service:  e.Service.Name,
		Hostname: Hostname,
		Value:    e.Value,
	}

	return json.Marshal(&ee)
}

type EventDispatcher struct {
	Registry map[*Service]map[string][]*Handler
	Global   map[string][]*Handler
	Actions  []Action
}

func NewEventDispatcher(c *Config) *EventDispatcher {
	ev := &EventDispatcher{
		Registry: make(map[*Service]map[string][]*Handler),
		Global:   make(map[string][]*Handler),
	}

	ev.Actions = []Action{
		&processAction{},
		&emailAction{&c.Email},
		&scriptAction{},
		&webhookAction{},
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
	show("Dispatching event '%s' for '%s': %v", ev.Name, ev.Service.Name, ev.Value)

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

type ActionError struct {
	Event *Event
	Error error
}

func (e *EventDispatcher) Process(ev *Event, hs []*Handler) error {
	for _, h := range hs {
		show("Handling event '%s' for '%s': %v => %#v", ev.Name, ev.Service.Name, ev.Value, h)

		for _, a := range e.Actions {
			err := a.Run(ev, h)
			if err != nil && err != ErrNotConfigured {
				if ev.Name == "action/error" {
					fmt.Fprint(os.Stderr, "Recursive error detected: %#v\n")
				} else {
					e.Dispatch(&Event{"action/error", ev.Service, &ActionError{ev, err}})
				}
			}
		}
	}
	return nil
}
