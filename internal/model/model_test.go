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
		nil,
		nil,
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

func TestNewResolvesPODSliceAliases(t *testing.T) {
	t.Parallel()

	api, err := New(
		nil,
		nil,
		[]*Slice{{Name: "ScoreList", Elem: TypeRef{Kind: TypePrimitive, Name: "u16", Raw: "u16", Primitive: PrimitiveInfo{Go: "uint16", Zig: "u16"}}}},
		nil,
		[]*Function{{Name: "scale_scores", Params: []Field{{Name: "scores", Type: TypeRef{Kind: TypeStruct, Name: "ScoreList", Raw: "ScoreList"}}}, Return: TypeRef{Kind: TypeStruct, Name: "ScoreList", Raw: "ScoreList"}}},
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := api.Funcs[0].Params[0].Type.Kind; got != TypeSlice {
		t.Fatalf("scale_scores param kind = %v, want slice", got)
	}
	if !api.TypeNeedsFree(api.Funcs[0].Return) {
		t.Fatal("slice return should require free")
	}
	if api.FunctionNeedsArena(api.Funcs[0]) {
		t.Fatal("POD slice params should not require arena allocation")
	}
}

func TestOptionalPODTraits(t *testing.T) {
	t.Parallel()

	api, err := New(nil, nil, nil, nil, []*Function{{
		Name: "choose_limit",
		Params: []Field{{
			Name: "value",
			Type: TypeRef{Kind: TypeOptional, Raw: "?u32", Elem: &TypeRef{Kind: TypePrimitive, Name: "u32", Raw: "u32", Primitive: PrimitiveInfo{Go: "uint32", Zig: "u32"}}},
		}},
		Return: TypeRef{Kind: TypeOptional, Raw: "?u32", Elem: &TypeRef{Kind: TypePrimitive, Name: "u32", Raw: "u32", Primitive: PrimitiveInfo{Go: "uint32", Zig: "u32"}}},
	}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !api.IsPOD(api.Funcs[0].Params[0].Type) {
		t.Fatal("optional u32 should be POD")
	}
	if api.TypeNeedsFree(api.Funcs[0].Return) {
		t.Fatal("optional u32 should not require free")
	}
	if api.TypeNeedsKeepAlive(api.Funcs[0].Return) {
		t.Fatal("optional u32 should not require keepalive")
	}
}

func TestNewRejectsDuplicateArrayAlias(t *testing.T) {
	t.Parallel()

	_, err := New(
		nil,
		nil,
		nil,
		[]*ArrayAlias{{Name: "Digest", Type: TypeRef{Kind: TypeArray, Raw: "[4]u8", ArrayLen: 4, Elem: &TypeRef{Kind: TypePrimitive, Name: "u8", Raw: "u8", Primitive: PrimitiveInfo{Go: "uint8", Zig: "u8"}}}}, {Name: "Digest", Type: TypeRef{Kind: TypeArray, Raw: "[8]u8", ArrayLen: 8, Elem: &TypeRef{Kind: TypePrimitive, Name: "u8", Raw: "u8", Primitive: PrimitiveInfo{Go: "uint8", Zig: "u8"}}}}},
		nil,
	)
	if err == nil {
		t.Fatal("New() error = nil, want duplicate array alias error")
	}
}

func TestNewResolvesArrayAliasInFunctions(t *testing.T) {
	t.Parallel()

	api, err := New(
		nil,
		nil,
		nil,
		[]*ArrayAlias{{Name: "Digest", Type: TypeRef{Kind: TypeArray, Raw: "[4]u8", ArrayLen: 4, Elem: &TypeRef{Kind: TypePrimitive, Name: "u8", Raw: "u8", Primitive: PrimitiveInfo{Go: "uint8", Zig: "u8"}}}}},
		[]*Function{{Name: "digest_name", Return: TypeRef{Kind: TypeStruct, Name: "Digest", Raw: "Digest"}}},
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := api.Funcs[0].Return.Kind; got != TypeArray {
		t.Fatalf("digest_name return kind = %v, want array", got)
	}
	if got := api.Funcs[0].Return.Alias; got != "Digest" {
		t.Fatalf("digest_name return alias = %q, want Digest", got)
	}
}

func TestTypeNeedsKeepAliveForNestedSliceFields(t *testing.T) {
	t.Parallel()

	api, err := New(
		[]*Struct{{Name: "Bucket", Fields: []Field{{Name: "scores", Type: TypeRef{Kind: TypeStruct, Name: "ScoreList", Raw: "ScoreList"}}}}},
		nil,
		[]*Slice{{Name: "ScoreList", Elem: TypeRef{Kind: TypePrimitive, Name: "u16", Raw: "u16", Primitive: PrimitiveInfo{Go: "uint16", Zig: "u16"}}}},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !api.TypeNeedsKeepAlive(TypeRef{Kind: TypeStruct, Name: "Bucket", Raw: "Bucket"}) {
		t.Fatal("Bucket should require keepalive because it contains a slice field")
	}
}

func TestStructSliceElemSupportMatrix(t *testing.T) {
	t.Parallel()

	api, err := New(
		[]*Struct{
			{Name: "Bucket", Fields: []Field{{Name: "scores", Type: TypeRef{Kind: TypeStruct, Name: "ScoreList", Raw: "ScoreList"}}}},
			{Name: "Nested", Fields: []Field{{Name: "groups", Type: TypeRef{Kind: TypeStruct, Name: "NestedList", Raw: "NestedList"}}}},
		},
		nil,
		[]*Slice{{Name: "ScoreList", Elem: TypeRef{Kind: TypePrimitive, Name: "u16", Raw: "u16", Primitive: PrimitiveInfo{Go: "uint16", Zig: "u16"}}}, {Name: "NestedList", Elem: TypeRef{Kind: TypeStruct, Name: "ScoreList", Raw: "ScoreList"}}},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !api.SupportsSliceElem(TypeRef{Kind: TypeStruct, Name: "Nested", Raw: "Nested"}) {
		t.Fatal("Nested should be allowed as slice element when its slice field element is POD")
	}

	api, err = New(
		[]*Struct{{Name: "Bucket", Fields: []Field{{Name: "scores", Type: TypeRef{Kind: TypeStruct, Name: "ScoreList", Raw: "ScoreList"}}}}},
		nil,
		[]*Slice{{Name: "ScoreList", Elem: TypeRef{Kind: TypePrimitive, Name: "u16", Raw: "u16", Primitive: PrimitiveInfo{Go: "uint16", Zig: "u16"}}}, {Name: "BucketList", Elem: TypeRef{Kind: TypeStruct, Name: "Bucket", Raw: "Bucket"}}},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !api.SupportsSliceElem(TypeRef{Kind: TypeStruct, Name: "Bucket", Raw: "Bucket"}) {
		t.Fatal("Bucket should be allowed as slice element when its slice field element is POD")
	}
}
