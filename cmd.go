package state

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func execCmd(command string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), stderr.String())
	}

	if len(stderr.String()) != 0 {
		fmt.Println(stderr.String())
	}

	return nil
}
