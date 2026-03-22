package parser

import (
	"strings"
	"testing"

	"go2zig/internal/model"
)

const sampleAPI = `
// builtin aliases used by go2zig
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const Bytes = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const ScoreList = extern struct {
    ptr: ?[*]const u16,
    len: usize,
};

pub const UserKindList = extern struct {
    ptr: ?[*]const UserKind,
    len: usize,
};

pub const DigestList = extern struct {
    ptr: ?[*]const [4]u8,
    len: usize,
};

pub const MetricList = extern struct {
    ptr: ?[*]const Metric,
    len: usize,
};

pub const UserKind = enum(u8) {
    guest,
    member,
    admin,
};

/* user models */
pub const User = extern struct {
    id: u64,
    kind: UserKind,
    name: String,
    email: String,
    scores: [3]u16,
};

pub const Metric = extern struct {
    kind: UserKind,
    scores: [3]u16,
};

pub const LoginRequest = extern struct {
    user: User,
    password: String,
};

pub const LoginResponse = extern struct {
    ok: bool,
    message: String,
    token: Bytes,
    digest: [4]u8,
};

pub const LoginError = error{
    InvalidPassword,
};

pub extern fn health() bool;
pub export fn login(req: LoginRequest) LoginResponse {
    unreachable;
}
pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
pub extern fn rename_user(user: User, next_name: String) User;
pub extern fn promote_user(user: User, next_kind: UserKind, next_scores: [3]u16) User;
pub extern fn digest_name(name: String) [4]u8;
pub extern fn scale_scores(scores: ScoreList, factor: u16) ScoreList;
pub extern fn mirror_kind_history(history: UserKindList) UserKindList;
pub extern fn duplicate_digest(seed: String) DigestList;
pub extern fn mirror_metrics(metrics: MetricList) MetricList;
`

func TestParse(t *testing.T) {
	t.Parallel()

	api, err := Parse(sampleAPI)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(api.Structs) != 4 {
		t.Fatalf("Parse() structs = %d, want 4", len(api.Structs))
	}
	if len(api.Enums) != 1 {
		t.Fatalf("Parse() enums = %d, want 1", len(api.Enums))
	}
	if len(api.Slices) != 4 {
		t.Fatalf("Parse() slices = %d, want 4", len(api.Slices))
	}
	if len(api.Funcs) != 10 {
		t.Fatalf("Parse() funcs = %d, want 10", len(api.Funcs))
	}

	if api.Struct("String") != nil || api.Struct("Bytes") != nil {
		t.Fatal("builtin String/Bytes should not be treated as user structs")
	}

	loginReq := api.Struct("LoginRequest")
	if loginReq == nil {
		t.Fatal("LoginRequest struct missing")
	}
	if got := loginReq.Fields[0].Type.Kind; got != model.TypeStruct {
		t.Fatalf("LoginRequest.user kind = %v, want struct", got)
	}
	if got := loginReq.Fields[1].Type.Kind; got != model.TypeString {
		t.Fatalf("LoginRequest.password kind = %v, want string", got)
	}
	if got := api.Struct("User").Fields[1].Type.Kind; got != model.TypeEnum {
		t.Fatalf("User.kind kind = %v, want enum", got)
	}
	if got := api.Struct("User").Fields[4].Type.Kind; got != model.TypeArray {
		t.Fatalf("User.scores kind = %v, want array", got)
	}

	if got := api.Funcs[0].Name; got != "health" {
		t.Fatalf("first function = %q, want health", got)
	}
	if got := api.Funcs[1].Return.Name; got != "LoginResponse" {
		t.Fatalf("login return = %q, want LoginResponse", got)
	}
	if !api.Funcs[2].CanErr {
		t.Fatal("login_checked should be marked CanErr")
	}
	if got := api.Funcs[2].Return.Name; got != "LoginResponse" {
		t.Fatalf("login_checked payload = %q, want LoginResponse", got)
	}
	if got := api.Funcs[3].Params[1].Type.Kind; got != model.TypeString {
		t.Fatalf("rename_user second param kind = %v, want string", got)
	}
	if got := api.Funcs[4].Params[1].Type.Kind; got != model.TypeEnum {
		t.Fatalf("promote_user second param kind = %v, want enum", got)
	}
	if got := api.Funcs[5].Return.Kind; got != model.TypeArray {
		t.Fatalf("digest_name return kind = %v, want array", got)
	}
	if got := api.Funcs[6].Params[0].Type.Kind; got != model.TypeSlice {
		t.Fatalf("scale_scores first param kind = %v, want slice", got)
	}
	if got := api.Funcs[7].Params[0].Type.Elem.Kind; got != model.TypeEnum {
		t.Fatalf("mirror_kind_history elem kind = %v, want enum", got)
	}
	if got := api.Funcs[8].Return.Elem.Kind; got != model.TypeArray {
		t.Fatalf("duplicate_digest elem kind = %v, want array", got)
	}
	if got := api.Funcs[9].Params[0].Type.Elem.Kind; got != model.TypeStruct {
		t.Fatalf("mirror_metrics elem kind = %v, want struct", got)
	}
}

func TestParseRejectsUnknownType(t *testing.T) {
	t.Parallel()

	_, err := Parse(`
        pub const Broken = extern struct {
            bad: Missing,
        };
    `)
	if err == nil {
		t.Fatal("Parse() error = nil, want unknown type error")
	}
	if !strings.Contains(err.Error(), "unknown type") {
		t.Fatalf("Parse() error = %q, want unknown type message", err)
	}
}

func TestParseSupportsVoidAndInlineErrorUnion(t *testing.T) {
	t.Parallel()

	api, err := Parse(`
        pub const String = extern struct { ptr: [*]const u8, len: usize, };
        pub extern fn ping() void;
        pub extern fn risky() error{Boom}!String;
    `)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(api.Funcs) != 2 {
		t.Fatalf("Parse() funcs = %d, want 2", len(api.Funcs))
	}
	if got := api.Funcs[0].Return.Kind; got != model.TypeVoid {
		t.Fatalf("ping return kind = %v, want void", got)
	}
	if !api.Funcs[1].CanErr {
		t.Fatal("risky should be marked CanErr")
	}
	if got := api.Funcs[1].Return.Kind; got != model.TypeString {
		t.Fatalf("risky payload kind = %v, want string", got)
	}
}

func TestParseRejectsDuplicateField(t *testing.T) {
	t.Parallel()

	_, err := Parse(`
        pub const User = extern struct {
            id: u64,
            id: u32,
        };
    `)
	if err == nil {
		t.Fatal("Parse() error = nil, want duplicate field error")
	}
	if !strings.Contains(err.Error(), "duplicate field") {
		t.Fatalf("Parse() error = %q, want duplicate field message", err)
	}
}
