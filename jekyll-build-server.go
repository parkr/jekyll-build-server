package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/zenazn/goji"
)

var (
	repoPrefix string
)

func isAuthorizedBuild(payload github.WebHookPayload) bool {
	return strings.HasPrefix(*payload.Repo.FullName, repoPrefix)
}

func shouldBuild(payload github.WebHookPayload) bool {
	return isAuthorizedBuild(payload) &&
		*payload.Ref == "refs/heads/master" &&
		*payload.After != "0000000000000000000000000000000000000000"
}

func postReceiveHook(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request %v", r)

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var payload github.WebHookPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Printf("Request to build %s", *payload.Repo.FullName)
	if isAuthorizedBuild(payload) {
		if shouldBuild(payload) {
			go buildJekyllSite(payload)
			fmt.Fprintf(w, "Building repo %s at revision %s", *payload.Repo.FullName, *payload.After)
		} else {
			log.Printf("Skipping build for %s at revision %s", *payload.Repo.FullName, *payload.After)
			fmt.Fprintf(w, "Skipping build for %s at revision %s", *payload.Repo.FullName, *payload.After)
		}
	} else {
		http.Error(w, "user not allowed to build here", 403)
		return
	}
}

func main() {
	flag.StringVar(&repoPrefix, "prefix", "", "The repo name prefix required in order to build.")
	flag.Parse()

	if repoPrefix == "" {
		log.Fatal("Specify a prefix to look for in the repo names with -prefix='name'")
	}

	if f, err := os.Stat(sourceBase); f == nil || os.IsNotExist(err) {
		log.Fatalf("The -src folder, %s, doesn't exist.", sourceBase)
	}

	if f, err := os.Stat(destBase); f == nil || os.IsNotExist(err) {
		log.Fatalf("The -dest folder, %s, doesn't exist.", destBase)
	}

	goji.Post("/_github", postReceiveHook)
	goji.Serve()
}
