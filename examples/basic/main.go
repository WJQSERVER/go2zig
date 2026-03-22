//go:build windows && amd64

package main

import (
	"fmt"
	"log"
)

func main() {
	if err := Default.Load(); err != nil {
		log.Fatal(err)
	}

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

	checked, err := LoginChecked(LoginRequest{
		User: User{
			ID:    7,
			Name:  "alice",
			Email: "alice@example.com",
		},
		Password: "secret-123",
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := LoginChecked(LoginRequest{
		User: User{
			ID:    7,
			Name:  "alice",
			Email: "alice@example.com",
		},
		Password: "bad",
	}); err == nil {
		log.Fatal("expected error from login_checked")
	}

	renamed := RenameUser(User{
		ID:    7,
		Name:  "alice",
		Email: "alice@example.com",
	}, "ally")

	fmt.Printf("login ok=%v message=%q token=%q\n", resp.OK, resp.Message, string(resp.Token))
	fmt.Printf("checked login ok=%v message=%q\n", checked.OK, checked.Message)
	fmt.Printf("renamed user=%+v\n", renamed)
}
