FROM golang

WORKDIR /go/src/github.com/parkr/jekyll-build-server

ADD . .

RUN go version

# Compile a standalone executable
RUN CGO_ENABLED=0 go install github.com/parkr/jekyll-build-server/...
