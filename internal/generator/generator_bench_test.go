package generator

import (
	"testing"

	"go2zig/internal/parser"
)

func BenchmarkRenderSampleAPI(b *testing.B) {
	api, err := parser.Parse(`
        pub const String = extern struct { ptr: [*]const u8, len: usize, };
        pub const Digest = [4]u8;
        pub const Bytes = extern struct { ptr: [*]const u8, len: usize, };
        pub const ScoreList = extern struct { ptr: ?[*]const u16, len: usize, };
        pub const UserKind = enum(u8) { guest, member, admin };
        pub const User = extern struct {
            id: u64,
            kind: UserKind,
            name: String,
            email: String,
            scores: [3]u16,
        };
        pub extern fn digest_name(name: String) Digest;
        pub extern fn scale_scores(scores: ScoreList, factor: u16) ScoreList;
        pub extern fn maybe_kind(flag: bool) ?UserKind;
        pub extern fn choose_limit(flag: bool, value: ?u32) ?u32;
    `)
	if err != nil {
		b.Fatalf("Parse() error = %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Render(api, Config{PackageName: "sample", LibraryName: "sample", APIModule: "api.zig", ImplModule: "lib.zig"}); err != nil {
			b.Fatalf("Render() error = %v", err)
		}
	}
}
