//go:build (amd64 || arm64) && (windows || linux || darwin)

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
			ID:     7,
			Kind:   UserKindMember,
			Name:   "alice",
			Email:  "alice@example.com",
			Scores: [3]uint16{3, 5, 8},
		},
		Password: "secret-123",
	})

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
		log.Fatal(err)
	}

	if _, err := LoginChecked(LoginRequest{
		User: User{
			ID:     7,
			Kind:   UserKindMember,
			Name:   "alice",
			Email:  "alice@example.com",
			Scores: [3]uint16{3, 5, 8},
		},
		Password: "bad",
	}); err == nil {
		log.Fatal("expected error from login_checked")
	}

	renamed := RenameUser(User{
		ID:     7,
		Kind:   UserKindMember,
		Name:   "alice",
		Email:  "alice@example.com",
		Scores: [3]uint16{3, 5, 8},
	}, "ally")
	promoted := PromoteUser(User{
		ID:     7,
		Kind:   UserKindMember,
		Name:   "alice",
		Email:  "alice@example.com",
		Scores: [3]uint16{3, 5, 8},
	}, UserKindAdmin, [3]uint16{13, 21, 34})
	digest := DigestName("alice")
	scaled := ScaleScores(ScoreList{2, 4, 6}, 3)
	history := MirrorKindHistory(UserKindList{UserKindGuest, UserKindAdmin})
	duplicates := DuplicateDigest("alice")
	metrics := MirrorMetrics(MetricList{{Kind: UserKindMember, Scores: [3]uint16{3, 5, 8}}, {Kind: UserKindAdmin, Scores: [3]uint16{13, 21, 34}}})
	users := MirrorUsers(UserList{{ID: 7, Kind: UserKindMember, Name: "alice", Email: "alice@example.com", Scores: [3]uint16{3, 5, 8}}, {ID: 8, Kind: UserKindAdmin, Name: "bob", Email: "bob@example.com", Scores: [3]uint16{13, 21, 34}}})
	buckets := MirrorBuckets(BucketList{{Kind: UserKindMember, Scores: ScoreList{2, 4, 6}}, {Kind: UserKindAdmin, Scores: ScoreList{3, 6, 9}}})
	kind := MaybeKind(true)
	noKind := MaybeKind(false)
	digestPtr := MaybeDigest(true)
	base := uint32(9)
	limit := ChooseLimit(true, &base)
	defaultLimit := ChooseLimit(true, nil)
	groups := MirrorScoreGroups(ScoreGroupList{ScoreList{1, 2}, ScoreList{3, 6, 9}})

	fmt.Printf("login ok=%v message=%q token=%q\n", resp.OK, resp.Message, string(resp.Token))
	fmt.Printf("checked login ok=%v message=%q\n", checked.OK, checked.Message)
	fmt.Printf("renamed user=%+v\n", renamed)
	fmt.Printf("promoted kind=%d scores=%v digest=%v scaled=%v history=%v duplicates=%v metrics=%v users=%v buckets=%v optionalKind=%d noKind=%v optionalDigest=%d optionalLimit=%d groups=%v\n", promoted.Kind, promoted.Scores, digest, scaled, history, duplicates, metrics, users, buckets, *kind, noKind == nil, (*digestPtr)[1], *limit+*defaultLimit, groups)
}
