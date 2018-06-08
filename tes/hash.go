package tes

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

// Hash returns a hash of the task, for caching.
// The hash is calculated only from fields which affect execution,
// so fields such as Name, Description, Tags, etc. are ignored.
func Hash(task *Task) (string, error) {
	h := md5.New()
	var err error

	write := func(d interface{}) {
		if err != nil {
			err = binary.Write(h, binary.LittleEndian, d)
		}
	}

	for _, in := range task.Inputs {
		write(in.Url)
		write(in.Path)
		write(in.Content)
	}

	for _, out := range task.Outputs {
		write(out.Url)
		write(out.Path)
	}

	for _, exec := range task.Executors {
		write(exec.Image)
		for _, arg := range exec.Command {
			write(arg)
		}
		write(exec.Workdir)
		write(exec.Stdin)
		write(exec.Stdout)
		write(exec.Stderr)
		for k, v := range exec.Env {
			write(k)
			write(v)
		}
	}

	if task.Resources != nil {
		r := task.Resources
		write(r.CpuCores)
		write(r.Preemptible)
		write(r.RamGb)
		write(r.DiskGb)
		for _, zone := range r.Zones {
			write(zone)
		}
	}

	for _, vol := range task.Volumes {
		write(vol)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), err
}
