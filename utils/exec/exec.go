/*
 * @Author: bill
 * @Date: 2021-10-28 18:27:16
 * @LastEditors: bill
 * @LastEditTime: 2021-11-02 10:21:17
 * @Description:
 * @FilePath: /ars-device-server/utils/exec/exec.go
 */
package exec

import (
	"bytes"
	"errors"
	"os/exec"
)

var ErrNoBinary = errors.New("No Binary!")
var ErrExecCommand = errors.New("Exec command error!")

func BackgroundExec() error {
	return nil
}

func GetOutByNoninteractionExec(binary string, args ...string) ([]byte, error) {
	binaryPath, err := exec.LookPath(binary)
	if err != nil {
		return nil, ErrNoBinary
	}

	cmd := exec.Command(binaryPath, args...)
	out, cmd_err := cmd.CombinedOutput()
	if cmd_err != nil {
		return out, ErrExecCommand
	}

	return out, nil
}

func GetOutByShellExec(command string) ([]byte, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	out, cmd_err := cmd.CombinedOutput()
	if cmd_err != nil {
		return nil, cmd_err
	}
	return out, nil
}

func InteractionExec(command string, args string, interactions ...string) error {
	input := bytes.NewBuffer(nil)
	//cmd := exec.Command(command, args)
	command += " " + args
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stdin = input
	go func() {
		for _, interaction := range interactions {
			input.WriteString(interaction + "\n")
		}
	}()
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func NoninteractionExec(binary string, args ...string) error {
	cmd := exec.Command(binary, args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
