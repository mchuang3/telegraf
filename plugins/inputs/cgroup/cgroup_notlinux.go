// +build !linux

package cgroup

import (
	"github.com/mchuang3/telegraf"
)

func (g *CGroup) Gather(acc telegraf.Accumulator) error {
	return nil
}
