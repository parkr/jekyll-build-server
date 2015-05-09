package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/libgit2/git2go"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

func postReceiveHook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", "github")
}

func main() {
	goji.Get("/_github", postReceiveHook)
	goji.Serve()
}
