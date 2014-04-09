package resorcerer

import (
	"bytes"
	"fmt"
	"github.com/scorredoira/email"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
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
		show("%s restart", ev.Service.Name)

		if DryRun {
			return nil
		}
		return ev.Service.action.Restart()
	case "stop":
		show("%s stop", ev.Service.Name)

		if DryRun {
			return nil
		}
		return ev.Service.action.Stop()
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

	subject := fmt.Sprintf(template, ev.Name, ev.Service.Name, Hostname)
	body := fmt.Sprintf(`
Resorcerer has detected an noteworthy event.

Time: %s
Host: %s
Service: %s
Event: %s

Value: %v
`, time.Now(), Hostname, ev.Service.Name, ev.Name, ev.Value)

	m := email.NewMessage(subject, body)

	if a.settings.From == "" {
		m.From = defaultFrom
	} else {
		m.From = a.settings.From
	}

	m.To = strings.Split(h.Email.Address, ",")

	show("Sending email:\n%s", string(m.Bytes()))

	if DryRun {
		return nil
	}

	host, _, _ := net.SplitHostPort(a.settings.Server)

	auth := smtp.PlainAuth("", a.settings.Username, a.settings.Password, host)
	return email.Send(a.settings.Server, auth, m)
}

type scriptAction struct{}

func (_ *scriptAction) Run(ev *Event, h *Handler) error {
	script := h.Script
	if script == "" {
		return ErrNotConfigured
	}

	json, err := ev.ToJson()
	if err != nil {
		return err
	}

	show("Run '%s', passing JSON: %s", script, string(json))

	if DryRun {
		return nil
	}

	c := exec.Command("bash", "-c", script)
	c.Stdin = bytes.NewReader(json)

	if Debug {
		c.Stdout = os.Stdout
	}

	return c.Run()
}

type webhookAction struct{}

func (_ *webhookAction) Run(ev *Event, h *Handler) error {
	url := h.WebHook
	if url == "" {
		return ErrNotConfigured
	}

	json, err := ev.ToJson()
	if err != nil {
		return err
	}

	show("POSTing to '%s', passing JSON: %s", url, string(json))

	if DryRun {
		return nil
	}

	_, err = http.Post(url, "application/json", bytes.NewReader(json))
	return err
}
