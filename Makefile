all: build test

build: deps
	go build

test: deps
	go test

deps:
	go get github.com/zenazn/goji \
		github.com/go-sql-driver/mysql \
		github.com/jmoiron/sqlx \
		github.com/google/go-github/github \
		github.com/stretchr/testify/assert
