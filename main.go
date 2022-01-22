package main

import (
	"encoding/json"
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
	if err != nil {
		exit1(err)
	}
	u, err := client.ListUsers()
	if err != nil {
		exit1(err)
	}
	j, err := json.Marshal(u)
	if err != nil {
		exit1(noclist.ErrFailed)
	}
	fmt.Printf("%s\n", j)
}

func exit1(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
