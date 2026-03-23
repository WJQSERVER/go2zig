//go:build (windows || darwin) && (amd64 || arm64)

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

	scaled := ScaleScores(ScoreList{2, 4, 6}, 3)
	if len(scaled) != 3 || scaled[0] != 6 || scaled[2] != 18 {
		t.Fatalf("ScaleScores() result = %v, want [6 12 18]", scaled)
	}

	history := MirrorKindHistory(UserKindList{UserKindGuest, UserKindAdmin})
	if len(history) != 2 || history[0] != UserKindGuest || history[1] != UserKindAdmin {
		t.Fatalf("MirrorKindHistory() result = %v, want [guest admin]", history)
	}

	duplicates := DuplicateDigest("alice")
	if len(duplicates) != 2 || duplicates[0] != [4]uint8{'a', 5, 0xAB, 0xCD} || duplicates[1][1] != 6 {
		t.Fatalf("DuplicateDigest() result = %v, want two digests", duplicates)
	}

	metrics := MirrorMetrics(MetricList{{Kind: UserKindMember, Scores: [3]uint16{3, 5, 8}}, {Kind: UserKindAdmin, Scores: [3]uint16{13, 21, 34}}})
	if len(metrics) != 2 || metrics[0].Kind != UserKindMember || metrics[1].Scores[0] != 13 {
		t.Fatalf("MirrorMetrics() result = %v, want mirrored metrics", metrics)
	}

	users := MirrorUsers(UserList{{ID: 7, Kind: UserKindMember, Name: "alice", Email: "alice@example.com", Scores: [3]uint16{3, 5, 8}}, {ID: 8, Kind: UserKindAdmin, Name: "bob", Email: "bob@example.com", Scores: [3]uint16{13, 21, 34}}})
	if len(users) != 2 || users[0].Name != "alice" || users[1].Email != "bob@example.com" {
		t.Fatalf("MirrorUsers() result = %v, want mirrored users", users)
	}

	buckets := MirrorBuckets(BucketList{{Kind: UserKindMember, Scores: ScoreList{2, 4, 6}}, {Kind: UserKindAdmin, Scores: ScoreList{3, 6, 9}}})
	if len(buckets) != 2 || len(buckets[0].Scores) != 3 || buckets[1].Scores[2] != 9 {
		t.Fatalf("MirrorBuckets() result = %v, want mirrored buckets", buckets)
	}

	kind := MaybeKind(true)
	if kind == nil || *kind != UserKindAdmin {
		t.Fatalf("MaybeKind(true) = %v, want admin", kind)
	}
	if got := MaybeKind(false); got != nil {
		t.Fatalf("MaybeKind(false) = %v, want nil", got)
	}
	digestPtr := MaybeDigest(true)
	if digestPtr == nil || (*digestPtr)[1] != 8 {
		t.Fatalf("MaybeDigest(true) = %v, want digest with second byte 8", digestPtr)
	}
	base := uint32(9)
	limit := ChooseLimit(true, &base)
	if limit == nil || *limit != 10 {
		t.Fatalf("ChooseLimit(true,&base) = %v, want 10", limit)
	}
	defaultLimit := ChooseLimit(true, nil)
	if defaultLimit == nil || *defaultLimit != 1 {
		t.Fatalf("ChooseLimit(true,nil) = %v, want 1", defaultLimit)
	}
	if got := ChooseLimit(false, &base); got != nil {
		t.Fatalf("ChooseLimit(false,&base) = %v, want nil", got)
	}

	groups := MirrorScoreGroups(ScoreGroupList{ScoreList{1, 2}, ScoreList{3, 6, 9}})
	if len(groups) != 2 || len(groups[1]) != 3 || groups[1][2] != 9 {
		t.Fatalf("MirrorScoreGroups() result = %v, want mirrored score groups", groups)
	}
}
