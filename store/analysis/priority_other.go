//go:build !unix

package analysis

import "os"

func lowerPriority(*os.Process) {}
