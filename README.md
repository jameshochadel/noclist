# TTS: Noclist

## Prerequisites

You must have [Go 1.17 or later](https://go.dev/doc/install) installed on your system.

## Commands

```sh
# Run the program
go run .

# Build the program
go build .

# Run the tests
go test ./...
```

## TODO

* Create context.Context in main(), pass to client's ListUsers method, watch for OS cancellation signals in main
