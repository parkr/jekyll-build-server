package main

import (
	"database/sql"
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
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var requiredOwner string

func buildsIndexHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	if db == nil {
		fmt.Fprintf(w, "database logging not enabled")
		return
	}

	builds := []Build{}
	err := db.Select(&builds, "SELECT * FROM builds ORDER BY created_at DESC")
	if err != nil {
		log.Printf("[%s] error listing builds: %v", middleware.GetReqID(c), err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if err = templates.ExecuteTemplate(w, "index.html", builds); err != nil {
		log.Printf("[%s] failed to render index.html: %v", middleware.GetReqID(c), err)
		w.Write([]byte(`FAILED`))
	}
}

func buildsShowHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	if db == nil {
		fmt.Fprintf(w, "database logging not enabled")
		return
	}

	id := fmt.Sprintf("%s/%s", c.URLParams["name"], c.URLParams["repo_tag"])
	var build Build
	if err := build.Get(id); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		switch err {
		case sql.ErrNoRows:
			log.Printf("[%s] build with id='%s' doesn't exist", middleware.GetReqID(c), id)
			http.Error(w, fmt.Sprintf("404 build %s not found", id), 404)
		default:
			log.Printf("[%s] error fetching build with id='%s': %v", middleware.GetReqID(c), id, err)
			http.Error(w, err.Error(), 500)
		}
		return
	} else {
		w.Header().Set("Content-Type", "text/html")
		if err = templates.ExecuteTemplate(w, "build.show.html", &build); err != nil {
			log.Printf("[%s] failed to render build.show.html: %v", middleware.GetReqID(c), err)
			w.Write([]byte(`FAILED`))
		}
	}
}

func isAuthorizedBuild(payload github.WebHookPayload) bool {
	return strings.HasPrefix(*payload.Repo.FullName, requiredOwner+"/")
}

func shouldBuild(payload github.WebHookPayload) bool {
	return isAuthorizedBuild(payload) &&
		(*payload.Ref == "refs/heads/master" || *payload.Ref == "refs/heads/gh-pages") &&
		*payload.After != "0000000000000000000000000000000000000000"
}

func postReceiveHook(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request %v", r)

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

	if &payload == nil || payload.After == nil {
		log.Println("received a ping request, must be... did you just setup a repo?")
		fmt.Fprintf(w, "no revision given...")
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
	flag.StringVar(&requiredOwner, "owner", "", "The owner required in order to build.")
	flag.Parse()

	if requiredOwner == "" {
		log.Fatal("Specify an owner to look for in the repo names with -owner='name'")
	}

	if f, err := os.Stat(sourceBase); f == nil || os.IsNotExist(err) {
		log.Fatalf("The -src folder, %s, doesn't exist.", sourceBase)
	}

	if f, err := os.Stat(destBase); f == nil || os.IsNotExist(err) {
		log.Fatalf("The -dest folder, %s, doesn't exist.", destBase)
	}

	if dbConnString != "" {
		InitDatabase()
	}

	goji.Get("/", buildsIndexHandler)
	goji.Get("/:name/:repo_tag", buildsShowHandler)
	goji.Post("/_github", postReceiveHook)
	goji.Serve()
}
