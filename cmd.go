package state

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func execCmd(name string, arg ...string) error {
	var errBuf bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s %s", err.Error(), errBuf.String())
	}
	if len(errBuf.String()) != 0 {
		fmt.Println(errBuf.String())
	}
	return nil
}

func execSplitCmd(cmd string) error {
	parts := strings.Split(cmd, " ")
	return execCmd(parts[0], parts[1:]...)
}
