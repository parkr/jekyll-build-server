package main

import (
	"log"
	"os"
	"os/exec"
)

func commandFromArgs(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func execute(args ...string) error {
	log.Println(args)
	return commandFromArgs(args...).Run()
}

func execInDir(dir string, args ...string) error {
	log.Println(args)
	cmd := commandFromArgs(args...)
	cmd.Dir = dir
	return cmd.Run()
}
