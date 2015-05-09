package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

func buildJekyllSite(payload github.WebHookPayload) {
	log.Println("Building payload:", payload)
	log.Println("")
	src, err := clone(&payload)
	if err != nil {
		log.Printf("Encountered an error cloning %s: %v", *payload.Repo.FullName, err)
		return
	}
	err = build(src, destination(payload.Repo))
	if err != nil {
		log.Printf("Encountered an error building %s: %v", *payload.Repo.FullName, err)
		return
	}

	log.Printf("Finished build for %s!", *payload.Repo.FullName)
}

func clone(payload *github.WebHookPayload) (src string, err error) {
	src = source(payload.Repo)
	f, err := os.Stat(src)
	if !f.IsDir() || os.IsNotExist(err) {
		log.Printf("Removing pre-existing non-directory %s", src)
		err = os.RemoveAll(src)
		if err != nil {
			return src, err
		}
	}

	if git, err := os.Stat(fmt.Sprintf("%s/.git", src)); !git.IsDir() || err != nil {
		log.Printf("Cloning the repo to %s ...", src)
		err = execute("git", "clone", *payload.Repo.SSHURL, src)
		if err != nil {
			return src, err
		}
	}

	err = execInDir(src, "git", "fetch", "origin")
	if err != nil {
		return
	}

	log.Printf("Checking out revision %s ...", *payload.After)
	err = execInDir(src, "git", "checkout", "--force", *payload.After)
	if err != nil {
		return
	}

	return
}

func build(src, dest string) (err error) {
	log.Printf("Building from %s to %s", src, dest)

	err = execInDir(src, "script/bootstrap", "--without development test")
	if err != nil {
		return
	}

	err = execInDir(src, "script/build", "-s", src, "-d", dest, "--full-rebuild")
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
