package parser

import (
	"strings"
	"testing"

	"go2zig/internal/model"
)

func TestParseRejectsMalformedOptionalType(t *testing.T) {
	t.Parallel()

	_, err := Parse(`pub const Broken = extern struct { value: ?, };`)
	if err == nil {
		t.Fatal("Parse() error = nil, want malformed optional error")
	}
	if !strings.Contains(err.Error(), "type is empty") {
		t.Fatalf("Parse() error = %q, want empty type message", err)
	}
}

func TestParseArrayAliasAndOptionalField(t *testing.T) {
	t.Parallel()

	api, err := Parse(`
        pub const Digest = [4]u8;
        pub const Item = extern struct {
            digest: ?Digest,
        };
    `)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(api.Arrays) != 1 {
		t.Fatalf("Parse() arrays = %d, want 1", len(api.Arrays))
	}
	if got := api.Struct("Item").Fields[0].Type.Kind; got != model.TypeOptional {
		t.Fatalf("Item.digest kind = %v, want optional", got)
	}
}

func TestParseRejectsDuplicateEnumValue(t *testing.T) {
	t.Parallel()

	_, err := Parse(`
        pub const State = enum(u8) {
            idle,
            idle,
        };
    `)
	if err == nil {
		t.Fatal("Parse() error = nil, want duplicate enum value error")
	}
	if !strings.Contains(err.Error(), "duplicate enum value") {
		t.Fatalf("Parse() error = %q, want duplicate enum value message", err)
	}
}

func TestParseRejectsUnsupportedEnumBase(t *testing.T) {
	t.Parallel()

	_, err := Parse(`pub const State = enum(bool) { off, on };`)
	if err == nil {
		t.Fatal("Parse() error = nil, want unsupported enum base error")
	}
	if !strings.Contains(err.Error(), "unsupported base type") {
		t.Fatalf("Parse() error = %q, want unsupported base type message", err)
	}
}

func TestParseResolvesArrayAliasInsideSlice(t *testing.T) {
	t.Parallel()

	api, err := Parse(`
        pub const Digest = [4]u8;
        pub const DigestList = extern struct {
            ptr: ?[*]const Digest,
            len: usize,
        };
        pub extern fn duplicate(seed: Digest) DigestList;
    `)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(api.Arrays) != 1 {
		t.Fatalf("Parse() arrays = %d, want 1", len(api.Arrays))
	}
	if len(api.Slices) != 1 {
		t.Fatalf("Parse() slices = %d, want 1", len(api.Slices))
	}
	if got := api.Funcs[0].Return.Kind; got != model.TypeSlice {
		t.Fatalf("duplicate return kind = %v, want slice", got)
	}
	if got := api.Funcs[0].Return.Elem.Alias; got != "Digest" {
		t.Fatalf("duplicate return element alias = %q, want Digest", got)
	}
}

func TestParseRejectsStringSliceAlias(t *testing.T) {
	t.Parallel()

	_, err := Parse(`
        pub const StringList = extern struct {
            ptr: ?[*]const String,
            len: usize,
        };
    `)
	if err == nil {
		t.Fatal("Parse() error = nil, want unsupported slice element error")
	}
	if !strings.Contains(err.Error(), "unsupported element type") {
		t.Fatalf("Parse() error = %q, want unsupported element type message", err)
	}
}

func TestParseRecognizesCRLFSliceAlias(t *testing.T) {
	t.Parallel()

	content := strings.ReplaceAll(`
        pub const ScoreList = extern struct {
            ptr: ?[*]const u16,
            len: usize,
        };
        pub extern fn mirror(scores: ScoreList) ScoreList;
    `, "\n", "\r\n")

	api, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(api.Slices) != 1 {
		t.Fatalf("Parse() slices = %d, want 1", len(api.Slices))
	}
	if len(api.Structs) != 0 {
		t.Fatalf("Parse() structs = %d, want 0", len(api.Structs))
	}
	if got := api.Funcs[0].Params[0].Type.Kind; got != model.TypeSlice {
		t.Fatalf("mirror param kind = %v, want slice", got)
	}
}
