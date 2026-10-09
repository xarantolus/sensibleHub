//go:build unix

package analysis

import (
	"os"
	"syscall"
)

// lowerPriority makes an analysis process yield the CPU to everything else,
// so a small server (e.g. a Raspberry Pi) stays responsive while it works
// through a library. Failure only means the analysis competes normally.
func lowerPriority(p *os.Process) {
	_ = syscall.Setpriority(syscall.PRIO_PROCESS, p.Pid, 19)
}
