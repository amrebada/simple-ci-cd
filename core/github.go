package core

import (
	"os"
	"os/exec"
)

func Clone(src, dest string) error {
	cmd := exec.Command("git", "clone", src, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
