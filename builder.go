package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
)

var (
	sourceBase string
	destBase   string
)

func init() {
	flag.StringVar(&sourceBase, "src", "/tmp", "Where the repo clones go.")
	flag.StringVar(&destBase, "dest", "/var/www", "Where the jekyll builds go.")
}

func formattedTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func buildJekyllSite(payload github.WebHookPayload) {
	buildId := fmt.Sprintf("%s-%s", *payload.Repo.FullName, (*payload.After)[0:10])
	execer := &Execer{&Build{
		Id:        buildId,
		Success:   false,
		CreatedAt: formattedTime(),
	}}

	if execer.Build.Exists() {
		log.Printf("[%s] system: build already exists. re-running...", execer.Build.Id)
		execer.Build.CreatedAt = formattedTime()
		execer.Build.Output = ""
		execer.Build.CompletedAt = ""
		execer.Build.Success = false
		execer.Build.Save()
	}

	src, err := clone(execer, &payload)
	if err != nil {
		execer.Fail("system: encountered an error cloning %s: %v", *payload.Repo.FullName, err)
		return
	}
	err = build(execer, src, destination(payload.Repo))
	if err != nil {
		execer.Fail("system: encountered an error building %s: %v", *payload.Repo.FullName, err)
		return
	}

	execer.Complete()
}

func clone(e *Execer, payload *github.WebHookPayload) (src string, err error) {
	src = source(payload.Repo)

	f, err := os.Stat(src)
	if f != nil && !f.IsDir() {
		e.Log("system: Removing pre-existing non-directory %s", src)
		err = os.RemoveAll(src)
		if err != nil {
			return src, err
		}
	}

	if git, err := os.Stat(fmt.Sprintf("%s/.git", src)); (git != nil && !git.IsDir()) || err != nil {
		e.Log("system: Cloning the repo to %s ...", src)
		err = e.Execute("git", "clone", *payload.Repo.SSHURL, src)
		if err != nil {
			return src, err
		}
	}

	err = e.ExecInDir(src, "git", "fetch", "origin")
	if err != nil {
		return
	}

	e.Log("system: Checking out revision %s ...", *payload.After)
	err = e.ExecInDir(src, "git", "checkout", "--force", *payload.After)
	if err != nil {
		return
	}

	return
}

func build(e *Execer, src, dest string) (err error) {
	e.Log("system: Building from %s to %s", src, dest)

	err = e.ExecInDir(src, "script/bootstrap", "--without development test")
	if err != nil {
		return
	}

	err = e.ExecInDir(src, "script/build", "-s", src, "-d", dest, "--full-rebuild")
	if err != nil {
		return
	}

	return
}

func source(repo *github.Repository) string {
	return fmt.Sprintf("%s/%s", sourceBase, *repo.FullName)
}

func destination(repo *github.Repository) string {
	return fmt.Sprintf("%s/%s", destBase, *repo.FullName)
}
