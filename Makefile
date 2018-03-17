PKG=github.com/parkr/jekyll-build-server
RELEASE=$(shell git rev-parse HEAD)

all: build test

build: deps
	go install $(PKG)/...

test: deps
	go test $(PKG)/...

deps: godep
	godep save $(PKG)/... github.com/stretchr/testify/assert

godep:
	go get github.com/tools/godep

docker-build:
	docker build -t parkr/jekyll-build-server:$(RELEASE) .

docker-release: docker-build
	docker push parkr/jekyll-build-server:$(RELEASE)
