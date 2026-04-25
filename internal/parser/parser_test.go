package parser

import (
	"strings"
	"testing"

	"github.com/WJQSERVER/go2zig/internal/model"
)

const sampleAPI = `
// builtin aliases used by go2zig
pub const String = extern struct {
    ptr: [*]const u8,
    len: usize,
};

pub const Digest = [4]u8;

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
    ptr: ?[*]const Digest,
    len: usize,
};

pub const MetricList = extern struct {
    ptr: ?[*]const Metric,
    len: usize,
};

pub const UserList = extern struct {
    ptr: ?[*]const User,
    len: usize,
};

pub const BucketList = extern struct {
    ptr: ?[*]const Bucket,
    len: usize,
};

pub const ScoreGroupList = extern struct {
    ptr: ?[*]const ScoreList,
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

pub const Bucket = extern struct {
    kind: UserKind,
    scores: ScoreList,
};

pub const LoginRequest = extern struct {
    user: User,
    password: String,
};

pub const LoginResponse = extern struct {
    ok: bool,
    message: String,
    token: Bytes,
    digest: Digest,
};

pub const LoginError = error{
    InvalidPassword,
};

pub extern fn health() bool;
//go2zig:bridge-call noinline
//go2zig:go-noinline
pub export fn login(req: LoginRequest) LoginResponse {
    unreachable;
}
pub extern fn login_checked(req: LoginRequest) LoginError!LoginResponse;
pub extern fn rename_user(user: User, next_name: String) User;
pub extern fn promote_user(user: User, next_kind: UserKind, next_scores: [3]u16) User;
pub extern fn digest_name(name: String) Digest;
pub extern fn scale_scores(scores: ScoreList, factor: u16) ScoreList;
pub extern fn mirror_kind_history(history: UserKindList) UserKindList;
pub extern fn duplicate_digest(seed: String) DigestList;
pub extern fn mirror_metrics(metrics: MetricList) MetricList;
pub extern fn mirror_users(users: UserList) UserList;
pub extern fn mirror_buckets(buckets: BucketList) BucketList;
pub extern fn maybe_kind(flag: bool) ?UserKind;
pub extern fn maybe_digest(flag: bool) ?Digest;
pub extern fn choose_limit(flag: bool, value: ?u32) ?u32;
pub extern fn mirror_score_groups(groups: ScoreGroupList) ScoreGroupList;
`

func TestParse(t *testing.T) {
	t.Parallel()

	api, err := Parse(sampleAPI)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(api.Structs) != 5 {
		t.Fatalf("Parse() structs = %d, want 5", len(api.Structs))
	}
	if len(api.Enums) != 1 {
		t.Fatalf("Parse() enums = %d, want 1", len(api.Enums))
	}
	if len(api.Arrays) != 1 {
		t.Fatalf("Parse() arrays = %d, want 1", len(api.Arrays))
	}
	if len(api.Slices) != 7 {
		t.Fatalf("Parse() slices = %d, want 7", len(api.Slices))
	}
	if len(api.Funcs) != 16 {
		t.Fatalf("Parse() funcs = %d, want 16", len(api.Funcs))
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
	if got := api.Funcs[1].Codegen.BridgeCall; got != model.CallHintNoInline {
		t.Fatalf("login bridge call hint = %q, want noinline", got)
	}
	if !api.Funcs[1].Codegen.GoNoInline {
		t.Fatal("login should enable go noinline hint")
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
	if got := api.Funcs[8].Return.Kind; got != model.TypeSlice {
		t.Fatalf("duplicate_digest return kind = %v, want slice", got)
	}
	if got := api.Funcs[8].Return.Elem.Kind; got != model.TypeArray {
		t.Fatalf("duplicate_digest elem kind = %v, want array", got)
	}
	if got := api.Funcs[9].Params[0].Type.Elem.Kind; got != model.TypeStruct {
		t.Fatalf("mirror_metrics elem kind = %v, want struct", got)
	}
	if got := api.Funcs[10].Params[0].Type.Elem.Kind; got != model.TypeStruct {
		t.Fatalf("mirror_users elem kind = %v, want struct", got)
	}
	if got := api.Funcs[11].Params[0].Type.Elem.Kind; got != model.TypeStruct {
		t.Fatalf("mirror_buckets elem kind = %v, want struct", got)
	}
	if got := api.Funcs[12].Return.Kind; got != model.TypeOptional {
		t.Fatalf("maybe_kind return kind = %v, want optional", got)
	}
	if got := api.Funcs[13].Return.Elem.Kind; got != model.TypeArray {
		t.Fatalf("maybe_digest payload kind = %v, want array", got)
	}
	if got := api.Funcs[14].Params[1].Type.Kind; got != model.TypeOptional {
		t.Fatalf("choose_limit second param kind = %v, want optional", got)
	}
	if got := api.Funcs[15].Params[0].Type.Elem.Kind; got != model.TypeSlice {
		t.Fatalf("mirror_score_groups elem kind = %v, want slice", got)
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
