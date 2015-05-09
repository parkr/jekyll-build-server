# jekyll build server

Taking a webhook approach to building sites in Jekyll. Kind of like
jekyll-hook and others, but with a single binary because Go is nice.

## Installation & Usage

```shell
$ go get byparker.com/go/jekyll-build-server
$ $GOPATH/bin/jekyll-build-server -prefix="your-username-or-org"
```

Daemonize it or put it in a screen session if you want.

Add a webhook to this server's `POST /_github` endpoint.
Each request will be checked and only build if on the `master`
branch and if the repo has the proper prefix. It will clone
the repo into `${src}/${repo_nwo}` and build it into `${dest}/${repo_nwo}`

## Dependencies

You'll need:

- Go
- Ruby
- Bundler
- Git
- A public-facing IP or domain name

## Configuration

All the configuration happens in flags:

- `-prefix="your-username-or-org"` – the string your repos' full names (e.g. "parkr" in "parkr/jekyll-build-server") must start with in order to be authorized to be built
- `-bind=":9090"` – the port/host to bind the server to
- `-src="/tmp"` – the directory to clone the sources into
- `-dest="/var/www"` – the base directory to put built sites into

Run `jekyll-build-server -h` to learn more at any time.

## Server Configuration

I serve my static files with nginx. A very simple HTTP server might look like this:

```nginx
server {
  listen 80;
  server_name example.com;
  root /var/www/parkr/example.com;
  error_page 404 = /404.html;

  location / {
    try_files $uri $uri.html $uri/index.html index.html;
  }
}
```

This would serve the static files built for the `parkr/example.com` repo at `example.com:80`.


... that's really it!
