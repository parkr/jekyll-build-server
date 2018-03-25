# First, build
FROM golang:1.10.0 as builder
WORKDIR /go/src/github.com/parkr/jekyll-build-server
ADD . .
RUN go version
RUN CGO_ENABLED=0 GOOS=linux go install github.com/parkr/jekyll-build-server/...

# I use Jekyll, so the resulting Docker image is built with Ruby installed.
FROM ruby:2.5.0
COPY --from=builder /go/bin/jekyll-build-server /bin/jekyll-build-server
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8
CMD ["/bin/jekyll-build-server"]
