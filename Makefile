all: build test

build: deps
	go install github.com/parkr/jekyll-build-server/...

test: deps
	go test github.com/parkr/jekyll-build-server/...

deps: godep
	godep save github.com/parkr/jekyll-build-server/... github.com/stretchr/testify/assert

godep:
	go get github.com/tools/godep
