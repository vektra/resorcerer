package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/vektra/resorcerer/resorcerer"
	"github.com/vektra/resorcerer/upstart"
	"os"
)

type Options struct {
	Debug  bool `long:"debug" description:"Show debugging output"`
	DryRun bool `long:"dryrun" description:"Print out actions to take, but don't take them"`
}

var options Options

func listProcesses() {
	u, err := upstart.Dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}

	jobs, err := u.Jobs()
	if err != nil {
		panic(err)
	}

	fmt.Println("jobs on machine:")
	for _, job := range jobs {
		name, err := job.Name()
		if err != nil {
			panic(err)
		}

		instances, err := job.Instances()
		if err != nil {
			panic(err)
		}

		if len(instances) == 0 {
			continue
		}

		fmt.Println(name)

		for _, inst := range instances {
			procs, err := inst.Processes()
			if err != nil {
				continue
			}

			for _, p := range procs {
				fmt.Printf("  %v: %v\n", p.Name, p.Pid)
			}
		}
	}
}

func main() {
	rest, err := flags.Parse(&options)
	if err != nil {
		if se, ok := err.(*flags.Error); ok && se.Type != flags.ErrHelp {
			fmt.Fprintf(os.Stderr, "Error parsing options: %s\n", err)
		}
		os.Exit(1)
	}

	if options.Debug {
		fmt.Printf("Loading config: %s\n", rest[0])
	}

	resorcerer.Debug = options.Debug
	resorcerer.DryRun = options.DryRun

	for {
		c, err := resorcerer.LoadConfig(rest[0])
		if err != nil {
			fmt.Printf("Unable to load config: %s\n", err)
			os.Exit(1)
		}

		if c.Mode != "upstart" {
			fmt.Printf("Unsupported mode '%s'\n", c.Mode)
			os.Exit(1)
		}

		u, err := upstart.Dial()
		if err != nil {
			fmt.Printf("Unable to connect to Upstart: %s\n", err)
			os.Exit(1)
		}

		if resorcerer.RunLoop(u, c) == resorcerer.ErrReload {
			fmt.Fprint(os.Stderr, "! Reloading configuration\n")
			continue
		}
	}
}
