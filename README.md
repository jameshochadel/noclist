# Noclist

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

# Vet the code
go vet ./...
```

## Notes

Future improvements could include:

* "Politely" ask for to user list by adding backoff logic to retries
* Test for handling of dropped connections, not just non-200 response codes
* Test for presence of `x-request-checksum` header in test server
* Add end-to-end tests that validate JSON output of CLI tool
* Create `context.Context` in `main()`, pass to client's `ListUsers` and `New` methods, watch for OS cancellation signals in main
* Initialize all HTTP requests, or at least paths, just once in `New()` so they're not allocated every time `authenticate()` and `ListUsers()` are called
* Validate `ServerURL` more robustly; `ParseRequestURI` doesn't mandate a protocol, but requests fail without one
