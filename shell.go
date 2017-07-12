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
	Build *Build
}

func (e *Execer) commandFromArgs(args ...string) *exec.Cmd {
	e.Log("system: running command: %s", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stdin = os.Stdin
	return cmd
}

func (e *Execer) Log(args ...interface{}) {
	if formatStr, ok := args[0].(string); ok {
		// Log to stdout.
		logMsg := fmt.Sprintf("[%s] %s", e.Build.Id, fmt.Sprintf(formatStr, args[1:]...))
		log.Printf(logMsg)

		// Log to the database
		e.Build.Log(logMsg)
	} else {
		log.Printf("[%s] invalid format string: %v", e.Build.Id, formatStr)
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

func (e *Execer) Fail(args ...interface{}) {
	e.Log(args...)
	e.Build.Success = false
	e.Log("system: build terminated")
	e.End()
}

func (e *Execer) Complete() {
	e.Build.Success = true
	e.Log("system: build complete")
	e.End()
}

func (e *Execer) End() {
	e.Build.CompletedAt = mySQLFormattedTime()
	if err := e.Build.Save(); err != nil {
		log.Printf("[%s] exec: error updating db: %v", e.Build.Id, err)
	}
}
