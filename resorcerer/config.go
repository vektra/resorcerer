package resorcerer

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"strconv"
	"strings"
)

func show(str string, args ...interface{}) {
	if DryRun || Debug {
		fmt.Fprintf(os.Stderr, "=> "+str+"\n", args...)
	}
}

type MemoryAmount string

var Hostname string = "(unknown)"

func init() {
	if hostname, err := os.Hostname(); err == nil {
		Hostname = hostname
	}
}

func (m MemoryAmount) Bytes() (int, error) {
	s := string(m)

	n, err := strconv.Atoi(s)
	if err == nil {
		return n, nil
	}

	if strings.ToLower(s[len(s)-2:len(s)]) == "mb" {
		n, err := strconv.Atoi(strings.TrimSpace(s[0 : len(s)-2]))
		if err != nil {
			return 0, err
		}

		return n * (1024 * 1024), nil
	}

	return 0, fmt.Errorf("Unsupported memory ammount '%s'", m)
}

type Email struct {
	Address string `yaml:"address"`
	Subject string `yaml:"subject"`
}

type Handler struct {
	Event   string `yaml:"event"`
	Process string `yaml:"process"`
	Email   Email  `yaml:"email"`
	Script  string `yaml:"script"`
	WebHook string `yaml:"webhook"`
}

type ServiceAction interface {
	Restart() error
	Stop() error
}

type Service struct {
	Name     string       `yaml:"name"`
	Memory   MemoryAmount `yaml:"memory"`
	Handlers []*Handler   `yaml:"on"`
	action   ServiceAction
}

type EmailSettings struct {
	Server   string
	Username string
	Password string
	From     string
}

type PollSettings struct {
	Seconds     int
	Samples     int
	Significant int
}

type Config struct {
	Mode     string        `yaml:"mode"`
	Poll     PollSettings  `yaml:"poll"`
	Email    EmailSettings `yaml:"email"`
	Services []*Service    `yaml:"services"`
	Handlers []*Handler    `yaml:"on"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	c := &Config{}

	err = goyaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
