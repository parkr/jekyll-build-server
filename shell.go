package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kballard/go-shellquote"
)

type Execer struct {
	Build *Build
}

func (e *Execer) commandFromArgs(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stdin = os.Stdin
	e.Log("system: running command: %s", e.commandForLogging(cmd))
	return cmd
}

func (e *Execer) commandForLogging(cmd *exec.Cmd) string {
	return shellquote.Join(cmd.Args...)
}

func (e *Execer) scrubLogMsg(msg, secret string) string {
	return strings.Replace(msg, secret, "*******", -1)
}

func (e *Execer) Log(args ...interface{}) {
	if formatStr, ok := args[0].(string); ok {
		// Log to stdout.
		logMsg := fmt.Sprintf("[%s] %s", e.Build.Id, fmt.Sprintf(formatStr, args[1:]...))
		logMsg = e.scrubLogMsg(logMsg, os.Getenv("GITHUB_ACCESS_TOKEN"))
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
	startTime := time.Now()
	done := make(chan bool, 2)

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
		done <- true
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
		done <- true
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
	<-done // stderr
	<-done // stdout

	e.Log("system: completed command in %v: %s", time.Since(startTime), e.commandForLogging(cmd))

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
