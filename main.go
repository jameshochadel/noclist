package main

import (
	"fmt"

	"github.com/jameshochadel/noclist/internal/noclist"
)

func main() {
	client, err := noclist.New(
		noclist.Config{
			ServerURL: "http://localhost:8888",
		},
	)
	u, err := client.ListUsers()
	fmt.Printf("err: %v\n", err)
	fmt.Printf("client: %v\n", client)
	fmt.Printf("u: %v\n", u)
}
