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
