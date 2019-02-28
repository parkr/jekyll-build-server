PKG=github.com/parkr/jekyll-build-server
RELEASE=$(shell git rev-parse HEAD)

all: build test

build:
	go install $(PKG)/...

test:
	go test $(PKG)/...

deps:
	dep ensure
	dep prune

dive: docker-build
	dive parkr/jekyll-build-server:$(RELEASE)

docker-build:
	docker build -t parkr/jekyll-build-server:$(RELEASE) .

docker-release: docker-build
	docker push parkr/jekyll-build-server:$(RELEASE)
