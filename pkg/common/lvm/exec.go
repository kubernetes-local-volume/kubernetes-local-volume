package lvm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
)

func run(cmd string, v interface{}, extraArgs ...string) error {
	logger := logging.GetLogger()
	var args []string
	if v != nil {
		args = append(args, "--reportformat=json")
		args = append(args, "--units=M")
		args = append(args, "--nosuffix")
	}
	args = append(args, extraArgs...)
	cmd = cmd + " " + strings.Join(args, " ")

	c := exec.Command("sh", "-c", cmd)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	c.Stdout = stdout
	c.Stderr = stderr
	if err := c.Run(); err != nil {
		logger.Errorf("run cmd(%s) error = %s", cmd, err.Error())
		return err
	}

	stdoutbuf := stdout.Bytes()
	if v != nil {
		if err := json.Unmarshal(stdoutbuf, v); err != nil {
			return fmt.Errorf("%v: [%v]", err, string(stdoutbuf))
		}
	}
	return nil
}
