package resorcerer

import (
	"fmt"
	"github.com/scorredoira/email"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

var ErrNotConfigured = fmt.Errorf("No configuration set")

type Action interface {
	Run(ev *Event, h *Handler) error
}

var DryRun bool = false

type processAction struct{}

func (_ *processAction) Run(ev *Event, h *Handler) error {
	switch h.Process {
	case "":
		return ErrNotConfigured
	case "restart":
		if DryRun {
			fmt.Printf("%s restart\n", ev.Service.Name)
			return nil
		} else {
			return ev.Service.action.Restart()
		}
	default:
		return fmt.Errorf("Undefined process action: %s", h.Process)
	}
}

type emailAction struct {
	settings *EmailSettings
}

const defaultSubject = "[resorcerer] Event %s on %s on %s"
const defaultFrom = "resorcerer@vektra.io"

func (a *emailAction) Run(ev *Event, h *Handler) error {
	if h.Email.Address == "" {
		return ErrNotConfigured
	}

	template := h.Email.Subject
	if template == "" {
		template = defaultSubject
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "(unknown)"
	}

	subject := fmt.Sprintf(template, ev.Name, ev.Service.Name, hostname)
	body := fmt.Sprintf(`
Resorcerer has detected an noteworthy event.

Time: %s
Host: %s
Service: %s
Event: %s

Value: %v
`, time.Now(), hostname, ev.Service.Name, ev.Name, ev.Value)

	m := email.NewMessage(subject, body)

	if a.settings.From == "" {
		m.From = defaultFrom
	} else {
		m.From = a.settings.From
	}

	m.To = strings.Split(h.Email.Address, ",")

	if DryRun {
		fmt.Printf("Sending email:\n%s\n", string(m.Bytes()))
		return nil
	} else {
		host, _, _ := net.SplitHostPort(a.settings.Server)

		auth := smtp.PlainAuth("", a.settings.Username, a.settings.Password, host)
		return email.Send(a.settings.Server, auth, m)
	}
}
