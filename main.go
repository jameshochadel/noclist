package main

import (
	"fmt"
	"os"

	"github.com/jameshochadel/noclist/internal/noclist"
)

func main() {
	client, err := noclist.New(
		noclist.Config{
			ServerURL: "http://localhost:8888",
		},
	)
	u, err := client.ListUsers()
	if err != nil {
		os.Exit(1)
	}
	fmt.Printf("%v\n", u)
}
