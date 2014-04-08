package upstart

import (
	"fmt"
	"github.com/guelfey/go.dbus"
	"os"
)

type Conn struct {
	conn *dbus.Conn
}

type Job struct {
	u    *Conn
	path dbus.ObjectPath
}

const BusName = "com.ubuntu.Upstart"

func (u *Conn) object(path dbus.ObjectPath) *dbus.Object {
	return u.conn.Object(BusName, path)
}

func Dial() (*Conn, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	return &Conn{conn}, nil
}

func (u *Conn) Jobs() ([]*Job, error) {
	obj := u.object("/com/ubuntu/Upstart")

	var s []dbus.ObjectPath
	err := obj.Call("com.ubuntu.Upstart0_6.GetAllJobs", 0).Store(&s)
	if err != nil {
		return nil, err
	}

	var out []*Job

	for _, v := range s {
		out = append(out, &Job{u, v})
	}

	return out, nil
}

func (u *Conn) Job(name string) (*Job, error) {
	obj := u.object("/com/ubuntu/Upstart")

	var s dbus.ObjectPath
	err := obj.Call("com.ubuntu.Upstart0_6.GetJobByName", 0, name).Store(&s)
	if err != nil {
		return nil, err
	}

	return &Job{u, s}, nil
}

type Instance struct {
	j    *Job
	path dbus.ObjectPath
}

func (j *Job) obj() *dbus.Object {
	return j.u.object(j.path)
}

func (i *Instance) obj() *dbus.Object {
	return i.j.u.object(i.path)
}

func (j *Job) Instances() ([]*Instance, error) {
	var instances []dbus.ObjectPath

	err := j.obj().Call("com.ubuntu.Upstart0_6.Job.GetAllInstances", 0).Store(&instances)
	if err != nil {
		return nil, err
	}

	var out []*Instance

	for _, inst := range instances {
		out = append(out, &Instance{j, inst})
	}

	return out, nil
}

func (j *Job) Name() (string, error) {
	val, err := j.obj().GetProperty("com.ubuntu.Upstart0_6.Job.name")
	if err != nil {
		return "", err
	}

	if str, ok := val.Value().(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("Name was not a string")
}

func (j *Job) Pid() (int32, error) {
	insts, err := j.Instances()
	if err != nil {
		return 0, err
	}

	switch len(insts) {
	default:
		return 0, fmt.Errorf("More than 1 instances running, no single pid")
	case 0:
		return 0, fmt.Errorf("No instances of job available")
	case 1:
		procs, err := insts[0].Processes()
		if err != nil {
			return 0, err
		}

		switch len(procs) {
		default:
			return 0, fmt.Errorf("More than 1 processes running, no single pid")
		case 0:
			return 0, fmt.Errorf("No process running of any instances")
		case 1:
			return procs[0].Pid, nil
		}
	}
}

func (j *Job) Pids() ([]int32, error) {
	insts, err := j.Instances()
	if err != nil {
		return nil, err
	}

	var pids []int32

	for _, inst := range insts {
		procs, err := inst.Processes()
		if err != nil {
			return nil, err
		}

		for _, proc := range procs {
			pids = append(pids, proc.Pid)
		}
	}

	return pids, nil
}

func (j *Job) Restart() error {
	wait := false
	c := j.obj().Call("com.ubuntu.Upstart0_6.Job.Restart", 0, []string{}, wait)

	var inst dbus.ObjectPath
	return c.Store(&inst)
}

type Process struct {
	Name string
	Pid  int32
}

func (i *Instance) Processes() ([]Process, error) {
	val, err := i.obj().GetProperty("com.ubuntu.Upstart0_6.Instance.processes")

	if err != nil {
		return nil, err
	}

	var out []Process

	if ary, ok := val.Value().([][]interface{}); ok {
		for _, elem := range ary {
			out = append(out, Process{elem[0].(string), elem[1].(int32)})
		}
	} else {
		return nil, fmt.Errorf("Unable to decode processes property")
	}

	return out, nil
}

func main2() {
	u, err := Dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}

	obj := u.object("/com/ubuntu/Upstart")

	var s []dbus.ObjectPath
	err = obj.Call("com.ubuntu.Upstart0_6.GetAllJobs", 0).Store(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get list of owned names:", err)
		os.Exit(1)
	}

	fmt.Println("jobs on machine:")
	for _, v := range s {
		fmt.Println(v)

		j := u.object(v)
		var instances []dbus.ObjectPath

		err = j.Call("com.ubuntu.Upstart0_6.Job.GetAllInstances", 0).Store(&instances)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to get list of instances:", err)
			os.Exit(1)
		}

		for _, inst := range instances {
			fmt.Printf("  %s\n", inst)
			val, err := u.object(inst).GetProperty("com.ubuntu.Upstart0_6.Instance.processes")

			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to get list of processes:", err)
				os.Exit(1)
			}

			if ary, ok := val.Value().([][]interface{}); ok {
				for _, elem := range ary {
					fmt.Printf("  %v: %v\n", elem[0], elem[1])
				}
			} else {
				fmt.Printf("Bad type of prop\n")
			}

		}
	}
}

func all() {
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}

	var s []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get list of owned names:", err)
		os.Exit(1)
	}

	fmt.Println("Currently owned names on the session bus:")
	for _, v := range s {
		fmt.Println(v)
	}
}
