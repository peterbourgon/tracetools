# tracetools

Tools to process Go trace logs into various profiles. Complement for "go tool trace".

## Usage

```
cd $(mktemp -d)
env GOPATH=$(pwd) GO111MODULE=off go get -u -v github.com/peterbourgon/tracetools/...
```

The binaries are now available in the `bin/` subdirectory.
