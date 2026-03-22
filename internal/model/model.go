package model

import (
	"fmt"
	"sort"
)

type TypeKind int

const (
	TypeVoid TypeKind = iota
	TypePrimitive
	TypeString
	TypeBytes
	TypeStruct
)

type PrimitiveInfo struct {
	Zig string
	C   string
	Go  string
	CGo string
}

var primitiveTypes = map[string]PrimitiveInfo{
	"bool":  {Zig: "bool", C: "bool", Go: "bool", CGo: "C.bool"},
	"u8":    {Zig: "u8", C: "uint8_t", Go: "uint8", CGo: "C.uint8_t"},
	"u16":   {Zig: "u16", C: "uint16_t", Go: "uint16", CGo: "C.uint16_t"},
	"u32":   {Zig: "u32", C: "uint32_t", Go: "uint32", CGo: "C.uint32_t"},
	"u64":   {Zig: "u64", C: "uint64_t", Go: "uint64", CGo: "C.uint64_t"},
	"usize": {Zig: "usize", C: "size_t", Go: "uint", CGo: "C.size_t"},
	"i8":    {Zig: "i8", C: "int8_t", Go: "int8", CGo: "C.int8_t"},
	"i16":   {Zig: "i16", C: "int16_t", Go: "int16", CGo: "C.int16_t"},
	"i32":   {Zig: "i32", C: "int32_t", Go: "int32", CGo: "C.int32_t"},
	"i64":   {Zig: "i64", C: "int64_t", Go: "int64", CGo: "C.int64_t"},
	"isize": {Zig: "isize", C: "ptrdiff_t", Go: "int", CGo: "C.ptrdiff_t"},
	"f32":   {Zig: "f32", C: "float", Go: "float32", CGo: "C.float"},
	"f64":   {Zig: "f64", C: "double", Go: "float64", CGo: "C.double"},
}

type TypeRef struct {
	Kind      TypeKind
	Raw       string
	Name      string
	Primitive PrimitiveInfo
}

type Field struct {
	Name string
	Type TypeRef
}

type Struct struct {
	Name   string
	Fields []Field
}

type Function struct {
	Name   string
	Params []Field
	Return TypeRef
	CanErr bool
}

type API struct {
	Structs []*Struct
	Funcs   []*Function

	structByName map[string]*Struct
}

func New(structs []*Struct, funcs []*Function) (*API, error) {
	structByName := make(map[string]*Struct, len(structs))
	for _, item := range structs {
		if _, ok := structByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate struct %q", item.Name)
		}
		structByName[item.Name] = item
	}

	for _, item := range structs {
		for _, field := range item.Fields {
			if field.Type.Kind == TypeStruct {
				if _, ok := structByName[field.Type.Name]; !ok {
					return nil, fmt.Errorf("struct %q uses unknown type %q", item.Name, field.Type.Name)
				}
			}
		}
	}

	for _, item := range funcs {
		for _, param := range item.Params {
			if param.Type.Kind == TypeStruct {
				if _, ok := structByName[param.Type.Name]; !ok {
					return nil, fmt.Errorf("function %q uses unknown type %q", item.Name, param.Type.Name)
				}
			}
		}
		if item.Return.Kind == TypeStruct {
			if _, ok := structByName[item.Return.TypeName()]; !ok {
				return nil, fmt.Errorf("function %q returns unknown type %q", item.Name, item.Return.TypeName())
			}
		}
	}

	return &API{
		Structs:      structs,
		Funcs:        funcs,
		structByName: structByName,
	}, nil
}

func Primitive(raw string) (PrimitiveInfo, bool) {
	info, ok := primitiveTypes[raw]
	return info, ok
}

func PrimitiveNames() []string {
	names := make([]string, 0, len(primitiveTypes))
	for name := range primitiveTypes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (t TypeRef) TypeName() string {
	if t.Kind == TypeStruct {
		return t.Name
	}
	return t.Raw
}

func (a *API) Struct(name string) *Struct {
	return a.structByName[name]
}

func (a *API) StructNeedsAllocation(name string) bool {
	return a.typeNeedsAllocation(TypeRef{Kind: TypeStruct, Name: name, Raw: name}, map[string]bool{})
}

func (a *API) TypeNeedsAllocation(t TypeRef) bool {
	return a.typeNeedsAllocation(t, map[string]bool{})
}

func (a *API) typeNeedsAllocation(t TypeRef, seen map[string]bool) bool {
	switch t.Kind {
	case TypeString, TypeBytes:
		return true
	case TypeStruct:
		if seen[t.Name] {
			return false
		}
		seen[t.Name] = true
		item := a.Struct(t.Name)
		if item == nil {
			return false
		}
		for _, field := range item.Fields {
			if a.typeNeedsAllocation(field.Type, seen) {
				return true
			}
		}
	}
	return false
}

func (a *API) TypeNeedsFree(t TypeRef) bool {
	return a.typeNeedsFree(t, map[string]bool{})
}

func (a *API) typeNeedsFree(t TypeRef, seen map[string]bool) bool {
	switch t.Kind {
	case TypeString, TypeBytes:
		return true
	case TypeStruct:
		if seen[t.Name] {
			return false
		}
		seen[t.Name] = true
		item := a.Struct(t.Name)
		if item == nil {
			return false
		}
		for _, field := range item.Fields {
			if a.typeNeedsFree(field.Type, seen) {
				return true
			}
		}
	}
	return false
}

func (a *API) FunctionNeedsArena(fn *Function) bool {
	for _, param := range fn.Params {
		if a.TypeNeedsAllocation(param.Type) {
			return true
		}
	}
	return false
}

func (a *API) SortedStructs() ([]*Struct, error) {
	order := make([]*Struct, 0, len(a.Structs))
	state := make(map[string]int, len(a.Structs))

	var visit func(item *Struct) error
	visit = func(item *Struct) error {
		switch state[item.Name] {
		case 1:
			return fmt.Errorf("cyclic struct dependency involving %q", item.Name)
		case 2:
			return nil
		}
		state[item.Name] = 1
		for _, field := range item.Fields {
			if field.Type.Kind != TypeStruct {
				continue
			}
			dep := a.Struct(field.Type.Name)
			if dep == nil {
				return fmt.Errorf("struct %q uses unknown type %q", item.Name, field.Type.Name)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}
		state[item.Name] = 2
		order = append(order, item)
		return nil
	}

	for _, item := range a.Structs {
		if err := visit(item); err != nil {
			return nil, err
		}
	}

	return order, nil
}
