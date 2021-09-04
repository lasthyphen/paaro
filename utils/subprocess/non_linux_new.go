// +build !linux

// (c) 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package subprocess

import "os/exec"

func New(path string, args ...string) *exec.Cmd {
	return exec.Command(path, args...)
}
