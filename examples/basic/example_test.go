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
			ID:     7,
			Kind:   UserKindMember,
			Name:   "alice",
			Email:  "alice@example.com",
			Scores: [3]uint16{3, 5, 8},
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
	if resp.Digest != [4]uint8{1, 2, 3, 4} {
		t.Fatalf("Login() digest = %v, want %v", resp.Digest, [4]uint8{1, 2, 3, 4})
	}

	checked, err := LoginChecked(LoginRequest{
		User: User{
			ID:     7,
			Kind:   UserKindMember,
			Name:   "alice",
			Email:  "alice@example.com",
			Scores: [3]uint16{3, 5, 8},
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
		User:     User{ID: 7, Kind: UserKindMember, Name: "alice", Email: "alice@example.com", Scores: [3]uint16{3, 5, 8}},
		Password: "bad",
	}); err == nil {
		t.Fatal("LoginChecked() error = nil, want error")
	}

	renamed := RenameUser(User{ID: 7, Kind: UserKindMember, Name: "alice", Email: "alice@example.com", Scores: [3]uint16{3, 5, 8}}, "ally")
	if renamed.Name != "ally" {
		t.Fatalf("RenameUser() name = %q, want %q", renamed.Name, "ally")
	}
	if renamed.Email != "alice@example.com" {
		t.Fatalf("RenameUser() email = %q, want %q", renamed.Email, "alice@example.com")
	}
	if renamed.Kind != UserKindMember {
		t.Fatalf("RenameUser() kind = %d, want %d", renamed.Kind, UserKindMember)
	}

	promoted := PromoteUser(User{ID: 7, Kind: UserKindMember, Name: "alice", Email: "alice@example.com", Scores: [3]uint16{3, 5, 8}}, UserKindAdmin, [3]uint16{13, 21, 34})
	if promoted.Kind != UserKindAdmin {
		t.Fatalf("PromoteUser() kind = %d, want %d", promoted.Kind, UserKindAdmin)
	}
	if promoted.Scores != [3]uint16{13, 21, 34} {
		t.Fatalf("PromoteUser() scores = %v, want %v", promoted.Scores, [3]uint16{13, 21, 34})
	}

	digest := DigestName("alice")
	if digest != [4]uint8{'a', 5, 0xAB, 0xCD} {
		t.Fatalf("DigestName() digest = %v, want %v", digest, [4]uint8{'a', 5, 0xAB, 0xCD})
	}
}
