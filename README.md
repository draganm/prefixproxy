# PrefixProxy

Simplest HTTP proxy that will match requests by path prefix and forward them to the configured backends.

Main reason for writing this proxy is to be used in the situation where both front-end and backend are being developed at the same time.
For example, when front-end is served with Rollup.js, but certain paths have to be served by a Golang backend.

## Installation

PrefixProxy is written in Golang. At the moment the only way to get the latest release is to install it using `go get`.

```bash
$ go get -u github.com/draganm/prefixproxy
```

## Usage

PrefixProxy is configured using command line arguments.
Each argument has the form `<path>=<url>` where `<path>` is the path matcher and `<url>` is the URL of the backend where the request will be proxied to.

```bash
$ prefixproxy --port=8000 /=http://localhost:5000 /api/=http://localhost:5100
```

