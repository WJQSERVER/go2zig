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
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := api.SortedStructs(); err == nil {
		t.Fatal("SortedStructs() error = nil, want cycle error")
	}
}
