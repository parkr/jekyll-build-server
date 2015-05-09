package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Execer struct {
	BuildId string
}

func (e *Execer) commandFromArgs(args ...string) *exec.Cmd {
	e.Log("system: running command: %s", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stdin = os.Stdin
	return cmd
}

func (e *Execer) Log(args ...interface{}) {
	if formatStr, ok := args[0].(string); ok {
		log.Printf("[%s] %s", e.BuildId, fmt.Sprintf(formatStr, args[1:]...))
	} else {
		log.Printf("invalid format string: %v", formatStr)
	}
}

func (e *Execer) Execute(args ...string) error {
	return e.runCommand(e.commandFromArgs(args...))
}

func (e *Execer) ExecInDir(dir string, args ...string) error {
	cmd := e.commandFromArgs(args...)
	cmd.Dir = dir
	return e.runCommand(cmd)
}

func (e *Execer) runCommand(cmd *exec.Cmd) error {
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		e.Log("system: Error creating stdout pipe for command %v - %v", cmd, err)
		return err
	}

	go func() {
		scanner := bufio.NewScanner(cmdReader)
		for scanner.Scan() {
			e.Log("stdout: %s", scanner.Text())
		}
	}()

	cmdErrReader, err := cmd.StderrPipe()
	if err != nil {
		e.Log("system: Error creating stderr pipe for command %v - %v", cmd, err)
		return err
	}

	go func() {
		scanner := bufio.NewScanner(cmdErrReader)
		for scanner.Scan() {
			e.Log("stderr: %s", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		e.Log("system: error starting command %v - %v", cmd, err)
		return err
	}

	err = cmd.Wait()
	if err != nil {
		e.Log("system: error waiting for command %v - %v", cmd, err)
		return err
	}

	return nil
}
