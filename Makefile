PKG=github.com/parkr/jekyll-build-server
RELEASE=$(shell git rev-parse HEAD)

all: build test

build: deps
	go install $(PKG)/...

test: deps
	go test $(PKG)/...

deps:
	dep ensure
	dep prune

docker-build:
	docker build -t parkr/jekyll-build-server:$(RELEASE) .

docker-release: docker-build
	docker push parkr/jekyll-build-server:$(RELEASE)
