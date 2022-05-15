package core

import (
	"fmt"
	"os"
	"os/exec"
)

func Clone(src, dest string) error {
	cmd := exec.Command("ssh-agent", "bash", "-c", fmt.Sprintf("ssh-add %s; git clone %s %s", os.Getenv("SSH_KEY_PATH"), src, dest))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
