package procstats

type ProcessState int

const (
	Run ProcessState = iota
	Sleep
	IO
	Zombie
	Stopped
	Paging
)

type Info struct {
	Pid        Pid
	State      ProcessState
	ParentPid  Pid
	ProcessGrp int
	VirtMem    Bytes
	RSS        Bytes
}
