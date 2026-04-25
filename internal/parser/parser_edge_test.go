package parser

import (
	"strings"
	"testing"

	"github.com/WJQSERVER/go2zig/internal/model"
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

func TestParseRecognizesStreamTypes(t *testing.T) {
	t.Parallel()

	api, err := Parse(`
        pub extern fn consume(reader: GoReader, writer: GoWriter) void;
    `)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := api.Funcs[0].Params[0].Type.Kind; got != model.TypeGoReader {
		t.Fatalf("consume reader kind = %v, want GoReader", got)
	}
	if got := api.Funcs[0].Params[1].Type.Kind; got != model.TypeGoWriter {
		t.Fatalf("consume writer kind = %v, want GoWriter", got)
	}
}

func TestParseRejectsNestedStreamTypes(t *testing.T) {
	t.Parallel()

	_, err := Parse(`
        pub const Payload = extern struct {
            reader: GoReader,
        };
    `)
	if err == nil {
		t.Fatal("Parse() error = nil, want nested stream rejection")
	}
	if !strings.Contains(err.Error(), "unsupported stream type") {
		t.Fatalf("Parse() error = %q, want unsupported stream type message", err)
	}
}

func TestParseCodegenDirectives(t *testing.T) {
	t.Parallel()

	api, err := Parse(`
        pub const String = extern struct { ptr: [*]const u8, len: usize, };
        //go2zig:bridge-call inline
        //go2zig:go-noinline
        pub extern fn login(name: String) String;
    `)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := api.Funcs[0].Codegen.BridgeCall; got != model.CallHintInline {
		t.Fatalf("login bridge call hint = %q, want inline", got)
	}
	if !api.Funcs[0].Codegen.GoNoInline {
		t.Fatal("login should enable go noinline hint")
	}
}

func TestParseCodegenDirectivesAcrossMultilineFunctionDecl(t *testing.T) {
	t.Parallel()

	api, err := Parse(`
        pub const String = extern struct { ptr: [*]const u8, len: usize, };
        //go2zig:bridge-call noinline
        pub extern fn
        login(name: String) String;
    `)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := api.Funcs[0].Name; got != "login" {
		t.Fatalf("function name = %q, want login", got)
	}
	if got := api.Funcs[0].Codegen.BridgeCall; got != model.CallHintNoInline {
		t.Fatalf("login bridge call hint = %q, want noinline", got)
	}
}

func TestParseRejectsUnknownCodegenDirective(t *testing.T) {
	t.Parallel()

	_, err := Parse(`
        //go2zig:unknown value
        pub extern fn login() void;
    `)
	if err == nil {
		t.Fatal("Parse() error = nil, want codegen directive error")
	}
	if !strings.Contains(err.Error(), "unsupported codegen directive") {
		t.Fatalf("Parse() error = %q, want unsupported codegen directive message", err)
	}
}

func TestParseRejectsDetachedCodegenDirective(t *testing.T) {
	t.Parallel()

	_, err := Parse(`
        //go2zig:bridge-call inline
        pub const String = extern struct { ptr: [*]const u8, len: usize, };
    `)
	if err == nil {
		t.Fatal("Parse() error = nil, want detached codegen directive error")
	}
	if !strings.Contains(err.Error(), "must be attached to a function declaration") {
		t.Fatalf("Parse() error = %q, want detached directive message", err)
	}
}
