package proc

import (
	"errors"

	"github.com/shirou/gopsutil/v3/process"
)

func GetPid(processName string) (int, error) {
	processes, err := process.Processes()
	if err != nil {
		return 0, err
	}
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}
		if name == processName {
			return int(proc.Pid), nil
		}
	}
	return 0, errors.New("no process with that name exists")
}
