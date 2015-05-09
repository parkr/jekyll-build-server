package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
)

func commandFromArgs(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stdin = os.Stdin
	return cmd
}

func runCommand(cmd *exec.Cmd) error {
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Error creating stdout pipe for command %v - %v", cmd, err)
		return err
	}

	go func() {
		scanner := bufio.NewScanner(cmdReader)
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()

	cmdErrReader, err := cmd.StderrPipe()
	if err != nil {
		log.Println("Error creating stderr pipe for command %v - %v", cmd, err)
		return err
	}

	go func() {
		scanner := bufio.NewScanner(cmdErrReader)
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Println("Error starting command %v - %v", cmd, err)
		return err
	}

	err = cmd.Wait()
	if err != nil {
		log.Println("Error waiting for command %v - %v", cmd, err)
		return err
	}

	return nil
}

func execute(args ...string) error {
	log.Println(args)
	return runCommand(commandFromArgs(args...))
}

func execInDir(dir string, args ...string) error {
	log.Println(args)
	cmd := commandFromArgs(args...)
	cmd.Dir = dir
	return runCommand(cmd)
}
