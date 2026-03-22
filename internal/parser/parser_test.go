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

/* user models */
pub const User = extern struct {
    id: u64,
    name: String,
    email: String,
};

pub const LoginRequest = extern struct {
    user: User,
    password: String,
};

pub const LoginResponse = extern struct {
    ok: bool,
    message: String,
    token: Bytes,
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
`

func TestParse(t *testing.T) {
	t.Parallel()

	api, err := Parse(sampleAPI)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(api.Structs) != 3 {
		t.Fatalf("Parse() structs = %d, want 3", len(api.Structs))
	}
	if len(api.Funcs) != 4 {
		t.Fatalf("Parse() funcs = %d, want 4", len(api.Funcs))
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
