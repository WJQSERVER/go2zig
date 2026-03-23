//go:build ((windows || linux) && (amd64 || arm64)) || (darwin && arm64)

package main

import (
	"sync"
	"testing"
)

var loadOnce sync.Once

func ensureLoaded(b testing.TB) {
	b.Helper()
	loadOnce.Do(func() {
		if err := Default.Load(); err != nil {
			b.Fatalf("Load() error = %v", err)
		}
	})
}

func BenchmarkHealth(b *testing.B) {
	ensureLoaded(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Health()
	}
}

func BenchmarkDigestName(b *testing.B) {
	ensureLoaded(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DigestName("alice")
	}
}

func BenchmarkScaleScores(b *testing.B) {
	ensureLoaded(b)
	scores := ScoreList{2, 4, 6, 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScaleScores(scores, 3)
	}
}

func BenchmarkMirrorMetrics(b *testing.B) {
	ensureLoaded(b)
	metrics := MetricList{{Kind: UserKindMember, Scores: [3]uint16{3, 5, 8}}, {Kind: UserKindAdmin, Scores: [3]uint16{13, 21, 34}}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MirrorMetrics(metrics)
	}
}

func BenchmarkMirrorUsers(b *testing.B) {
	ensureLoaded(b)
	users := UserList{{ID: 7, Kind: UserKindMember, Name: "alice", Email: "alice@example.com", Scores: [3]uint16{3, 5, 8}}, {ID: 8, Kind: UserKindAdmin, Name: "bob", Email: "bob@example.com", Scores: [3]uint16{13, 21, 34}}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MirrorUsers(users)
	}
}

func BenchmarkMirrorBuckets(b *testing.B) {
	ensureLoaded(b)
	buckets := BucketList{{Kind: UserKindMember, Scores: ScoreList{2, 4, 6}}, {Kind: UserKindAdmin, Scores: ScoreList{3, 6, 9}}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MirrorBuckets(buckets)
	}
}

func BenchmarkMirrorScoreGroups(b *testing.B) {
	ensureLoaded(b)
	groups := ScoreGroupList{ScoreList{1, 2}, ScoreList{3, 6, 9}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MirrorScoreGroups(groups)
	}
}

func BenchmarkMaybeKind(b *testing.B) {
	ensureLoaded(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MaybeKind(i&1 == 0)
	}
}

func BenchmarkChooseLimit(b *testing.B) {
	ensureLoaded(b)
	base := uint32(9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ChooseLimit(i&1 == 0, &base)
	}
}
