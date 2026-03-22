//go:build amd64 && (windows || linux)

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

	checked, err := LoginChecked(LoginRequest{
		User: User{
			ID:    7,
			Name:  "alice",
			Email: "alice@example.com",
		},
		Password: "secret-123",
	})
	if err != nil {
		t.Fatalf("LoginChecked() error = %v", err)
	}
	if !checked.OK {
		t.Fatal("LoginChecked() returned not ok")
	}
	if checked.Message != "welcome alice" {
		t.Fatalf("LoginChecked() message = %q, want %q", checked.Message, "welcome alice")
	}
	if _, err := LoginChecked(LoginRequest{
		User:     User{ID: 7, Name: "alice", Email: "alice@example.com"},
		Password: "bad",
	}); err == nil {
		t.Fatal("LoginChecked() error = nil, want error")
	}

	renamed := RenameUser(User{ID: 7, Name: "alice", Email: "alice@example.com"}, "ally")
	if renamed.Name != "ally" {
		t.Fatalf("RenameUser() name = %q, want %q", renamed.Name, "ally")
	}
	if renamed.Email != "alice@example.com" {
		t.Fatalf("RenameUser() email = %q, want %q", renamed.Email, "alice@example.com")
	}
}
