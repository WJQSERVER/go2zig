//go:build windows && amd64

package main

import "testing"

func TestExampleAPI(t *testing.T) {
	if err := Default.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !Health() {
		t.Fatal("Health() = false, want true")
	}

	resp := Login(LoginRequest{
		User: User{
			ID:    7,
			Name:  "alice",
			Email: "alice@example.com",
		},
		Password: "secret-123",
	})
	if !resp.OK {
		t.Fatal("Login() returned not ok")
	}
	if resp.Message != "welcome alice" {
		t.Fatalf("Login() message = %q, want %q", resp.Message, "welcome alice")
	}
	if string(resp.Token) != "token-123" {
		t.Fatalf("Login() token = %q, want %q", string(resp.Token), "token-123")
	}

	renamed := RenameUser(User{ID: 7, Name: "alice", Email: "alice@example.com"}, "ally")
	if renamed.Name != "ally" {
		t.Fatalf("RenameUser() name = %q, want %q", renamed.Name, "ally")
	}
	if renamed.Email != "alice@example.com" {
		t.Fatalf("RenameUser() email = %q, want %q", renamed.Email, "alice@example.com")
	}
}
