package model

import "testing"

func TestSortedStructs(t *testing.T) {
	t.Parallel()

	api, err := New(
		[]*Struct{
			{Name: "LoginRequest", Fields: []Field{{Name: "user", Type: TypeRef{Kind: TypeStruct, Name: "User", Raw: "User"}}}},
			{Name: "User", Fields: []Field{{Name: "id", Type: TypeRef{Kind: TypePrimitive, Name: "u64", Raw: "u64", Primitive: PrimitiveInfo{Go: "uint64"}}}}},
		},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ordered, err := api.SortedStructs()
	if err != nil {
		t.Fatalf("SortedStructs() error = %v", err)
	}
	if len(ordered) != 2 {
		t.Fatalf("SortedStructs() len = %d, want 2", len(ordered))
	}
	if ordered[0].Name != "User" || ordered[1].Name != "LoginRequest" {
		t.Fatalf("SortedStructs() order = [%s %s], want [User LoginRequest]", ordered[0].Name, ordered[1].Name)
	}
}

func TestSortedStructsRejectsCycle(t *testing.T) {
	t.Parallel()

	api, err := New(
		[]*Struct{
			{Name: "A", Fields: []Field{{Name: "b", Type: TypeRef{Kind: TypeStruct, Name: "B", Raw: "B"}}}},
			{Name: "B", Fields: []Field{{Name: "a", Type: TypeRef{Kind: TypeStruct, Name: "A", Raw: "A"}}}},
		},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := api.SortedStructs(); err == nil {
		t.Fatal("SortedStructs() error = nil, want cycle error")
	}
}

func TestTypeNeedsAllocationAndArena(t *testing.T) {
	t.Parallel()

	api, err := New(
		[]*Struct{
			{Name: "User", Fields: []Field{{Name: "name", Type: TypeRef{Kind: TypeString, Name: "String", Raw: "String"}}}},
			{Name: "Wrapper", Fields: []Field{{Name: "user", Type: TypeRef{Kind: TypeStruct, Name: "User", Raw: "User"}}}},
		},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if !api.TypeNeedsAllocation(TypeRef{Kind: TypeStruct, Name: "User", Raw: "User"}) {
		t.Fatal("TypeNeedsAllocation(User) = false, want true")
	}
	if !api.TypeNeedsFree(TypeRef{Kind: TypeStruct, Name: "Wrapper", Raw: "Wrapper"}) {
		t.Fatal("TypeNeedsFree(Wrapper) = false, want true")
	}
	if api.TypeNeedsAllocation(TypeRef{Kind: TypePrimitive, Name: "u64", Raw: "u64", Primitive: PrimitiveInfo{Go: "uint64"}}) {
		t.Fatal("TypeNeedsAllocation(u64) = true, want false")
	}

	fn := &Function{
		Name:   "rename_user",
		Params: []Field{{Name: "user", Type: TypeRef{Kind: TypeStruct, Name: "Wrapper", Raw: "Wrapper"}}},
	}
	if !api.FunctionNeedsArena(fn) {
		t.Fatal("FunctionNeedsArena(rename_user) = false, want true")
	}
}

func TestNewResolvesEnumsAndArrays(t *testing.T) {
	t.Parallel()

	api, err := New(
		[]*Struct{{
			Name: "User",
			Fields: []Field{
				{Name: "kind", Type: TypeRef{Kind: TypeStruct, Name: "UserKind", Raw: "UserKind"}},
				{Name: "scores", Type: TypeRef{Kind: TypeArray, Raw: "[3]u16", ArrayLen: 3, Elem: &TypeRef{Kind: TypePrimitive, Name: "u16", Raw: "u16", Primitive: PrimitiveInfo{Go: "uint16", Zig: "u16"}}}},
			},
		}},
		[]*Enum{{Name: "UserKind", BaseName: "u8", Values: []EnumValue{{Name: "guest"}, {Name: "member"}}}},
		[]*Function{{Name: "digest", Return: TypeRef{Kind: TypeArray, Raw: "[3]UserKind", ArrayLen: 3, Elem: &TypeRef{Kind: TypeStruct, Name: "UserKind", Raw: "UserKind"}}}},
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := api.Struct("User").Fields[0].Type.Kind; got != TypeEnum {
		t.Fatalf("User.kind kind = %v, want enum", got)
	}
	if got := api.Funcs[0].Return.Elem.Kind; got != TypeEnum {
		t.Fatalf("digest elem kind = %v, want enum", got)
	}
}
