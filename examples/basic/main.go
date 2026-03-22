//go:build cgo

package main

import (
	"fmt"
	"log"
)

func main() {
	if !Health() {
		log.Fatal("zig library is not healthy")
	}

	resp := Login(LoginRequest{
		User: User{
			ID:    7,
			Name:  "alice",
			Email: "alice@example.com",
		},
		Password: "secret-123",
	})

	renamed := RenameUser(User{
		ID:    7,
		Name:  "alice",
		Email: "alice@example.com",
	}, "ally")

	fmt.Printf("login ok=%v message=%q token=%q\n", resp.OK, resp.Message, string(resp.Token))
	fmt.Printf("renamed user=%+v\n", renamed)
}
