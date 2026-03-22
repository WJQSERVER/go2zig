package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type TypeKind int

const (
	TypeVoid TypeKind = iota
	TypePrimitive
	TypeString
	TypeBytes
	TypeStruct
	TypeEnum
	TypeOptional
	TypeSlice
	TypeArray
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
	Alias     string
	Primitive PrimitiveInfo
	Elem      *TypeRef
	ArrayLen  int
}

type Field struct {
	Name string
	Type TypeRef
}

type Struct struct {
	Name   string
	Fields []Field
}

type EnumValue struct {
	Name  string
	Value string
}

type Enum struct {
	Name      string
	BaseName  string
	Primitive PrimitiveInfo
	Values    []EnumValue
}

type Slice struct {
	Name string
	Elem TypeRef
}

type ArrayAlias struct {
	Name string
	Type TypeRef
}

type Function struct {
	Name   string
	Params []Field
	Return TypeRef
	CanErr bool
}

type API struct {
	Structs []*Struct
	Enums   []*Enum
	Slices  []*Slice
	Arrays  []*ArrayAlias
	Funcs   []*Function

	structByName map[string]*Struct
	enumByName   map[string]*Enum
	sliceByName  map[string]*Slice
	arrayByName  map[string]*ArrayAlias
}

func New(structs []*Struct, enums []*Enum, slices []*Slice, arrays []*ArrayAlias, funcs []*Function) (*API, error) {
	structByName := make(map[string]*Struct, len(structs))
	for _, item := range structs {
		if _, ok := structByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate struct %q", item.Name)
		}
		structByName[item.Name] = item
	}

	enumByName := make(map[string]*Enum, len(enums))
	for _, item := range enums {
		if _, ok := structByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate type %q", item.Name)
		}
		if _, ok := enumByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate enum %q", item.Name)
		}
		base, ok := Primitive(item.BaseName)
		if !ok || item.BaseName == "bool" {
			return nil, fmt.Errorf("enum %q uses unsupported base type %q", item.Name, item.BaseName)
		}
		item.Primitive = base
		enumByName[item.Name] = item
	}

	sliceByName := make(map[string]*Slice, len(slices))
	for _, item := range slices {
		if _, ok := structByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate type %q", item.Name)
		}
		if _, ok := enumByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate type %q", item.Name)
		}
		if _, ok := sliceByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate slice %q", item.Name)
		}
		sliceByName[item.Name] = item
	}

	arrayByName := make(map[string]*ArrayAlias, len(arrays))
	for _, item := range arrays {
		if _, ok := structByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate type %q", item.Name)
		}
		if _, ok := enumByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate type %q", item.Name)
		}
		if _, ok := sliceByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate type %q", item.Name)
		}
		if _, ok := arrayByName[item.Name]; ok {
			return nil, fmt.Errorf("duplicate array alias %q", item.Name)
		}
		arrayByName[item.Name] = item
	}

	api := &API{
		Structs:      structs,
		Enums:        enums,
		Slices:       slices,
		Arrays:       arrays,
		Funcs:        funcs,
		structByName: structByName,
		enumByName:   enumByName,
		sliceByName:  sliceByName,
		arrayByName:  arrayByName,
	}

	for _, item := range structs {
		for i := range item.Fields {
			if err := api.resolveType(&item.Fields[i].Type, "struct "+item.Name); err != nil {
				return nil, err
			}
		}
	}

	for _, item := range arrays {
		if err := api.resolveType(&item.Type, "array alias "+item.Name); err != nil {
			return nil, err
		}
		if item.Type.Kind != TypeArray {
			return nil, fmt.Errorf("array alias %q must reference an array type", item.Name)
		}
		item.Type.Alias = item.Name
	}

	for _, item := range slices {
		if err := api.resolveType(&item.Elem, "slice "+item.Name); err != nil {
			return nil, err
		}
		if item.Elem.Kind == TypeString || item.Elem.Kind == TypeBytes {
			return nil, fmt.Errorf("slice %q uses unsupported element type %q", item.Name, item.Elem.TypeName())
		}
		if !api.SupportsSliceElem(item.Elem) {
			return nil, fmt.Errorf("slice %q uses unsupported element type %q", item.Name, item.Elem.TypeName())
		}
	}

	for _, item := range structs {
		for i := range item.Fields {
			if err := api.resolveType(&item.Fields[i].Type, "struct "+item.Name); err != nil {
				return nil, err
			}
		}
	}

	for _, item := range funcs {
		for i := range item.Params {
			if err := api.resolveType(&item.Params[i].Type, "function "+item.Name); err != nil {
				return nil, err
			}
		}
		if err := api.resolveType(&item.Return, "function "+item.Name); err != nil {
			return nil, err
		}
	}

	return api, nil
}

func (a *API) resolveType(t *TypeRef, owner string) error {
	switch t.Kind {
	case TypeStruct:
		if _, ok := a.structByName[t.Name]; ok {
			return nil
		}
		if enum, ok := a.enumByName[t.Name]; ok {
			t.Kind = TypeEnum
			t.Primitive = enum.Primitive
			return nil
		}
		if slice, ok := a.sliceByName[t.Name]; ok {
			t.Kind = TypeSlice
			elem := slice.Elem.Clone()
			t.Elem = &elem
			return nil
		}
		if array, ok := a.arrayByName[t.Name]; ok {
			copy := array.Type.Clone()
			*t = copy
			return nil
		}
		return fmt.Errorf("%s uses unknown type %q", owner, t.Name)
	case TypeArray:
		if t.Elem == nil {
			return fmt.Errorf("%s uses malformed array type %q", owner, t.Raw)
		}
		return a.resolveType(t.Elem, owner)
	case TypeSlice:
		if t.Elem == nil {
			return fmt.Errorf("%s uses malformed slice type %q", owner, t.Raw)
		}
		return a.resolveType(t.Elem, owner)
	case TypeOptional:
		if t.Elem == nil {
			return fmt.Errorf("%s uses malformed optional type %q", owner, t.Raw)
		}
		return a.resolveType(t.Elem, owner)
	default:
		return nil
	}
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

func (t TypeRef) Clone() TypeRef {
	copy := t
	if t.Elem != nil {
		elem := t.Elem.Clone()
		copy.Elem = &elem
	}
	return copy
}

func (t TypeRef) TypeName() string {
	switch t.Kind {
	case TypeStruct, TypeEnum:
		return t.Name
	case TypeOptional:
		if t.Elem == nil {
			return t.Raw
		}
		return "?" + t.Elem.TypeName()
	case TypeSlice:
		return t.Name
	case TypeArray:
		if t.Alias != "" {
			return t.Alias
		}
		if t.Elem == nil {
			return t.Raw
		}
		return "[" + strconv.Itoa(t.ArrayLen) + "]" + t.Elem.TypeName()
	default:
		return t.Raw
	}
}

func (t TypeRef) Key() string {
	switch t.Kind {
	case TypeSlice:
		if t.Elem == nil {
			return strings.TrimSpace(t.Raw)
		}
		return fmt.Sprintf("slice:%s", t.Elem.Key())
	case TypeOptional:
		if t.Elem == nil {
			return strings.TrimSpace(t.Raw)
		}
		return fmt.Sprintf("optional:%s", t.Elem.Key())
	case TypeArray:
		if t.Alias != "" {
			return "arrayalias:" + t.Alias
		}
		if t.Elem == nil {
			return strings.TrimSpace(t.Raw)
		}
		return fmt.Sprintf("array:%d:%s", t.ArrayLen, t.Elem.Key())
	default:
		return fmt.Sprintf("%d:%s", t.Kind, t.TypeName())
	}
}

func (a *API) Struct(name string) *Struct {
	return a.structByName[name]
}

func (a *API) Enum(name string) *Enum {
	return a.enumByName[name]
}

func (a *API) Slice(name string) *Slice {
	return a.sliceByName[name]
}

func (a *API) ArrayAlias(name string) *ArrayAlias {
	return a.arrayByName[name]
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
	case TypeSlice:
		return false
	case TypeOptional:
		if t.Elem == nil {
			return false
		}
		return a.typeNeedsAllocation(*t.Elem, seen)
	case TypeArray:
		if t.Elem == nil {
			return false
		}
		return a.typeNeedsAllocation(*t.Elem, seen)
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
	case TypeSlice:
		return true
	case TypeOptional:
		if t.Elem == nil {
			return false
		}
		return a.typeNeedsFree(*t.Elem, seen)
	case TypeArray:
		if t.Elem == nil {
			return false
		}
		return a.typeNeedsFree(*t.Elem, seen)
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

	var visitType func(owner string, t TypeRef) error
	visitType = func(owner string, t TypeRef) error {
		switch t.Kind {
		case TypeStruct:
			dep := a.Struct(t.Name)
			if dep == nil {
				return fmt.Errorf("%s uses unknown type %q", owner, t.Name)
			}
			return visit(dep)
		case TypeArray:
			if t.Elem == nil {
				return fmt.Errorf("%s uses malformed array type %q", owner, t.Raw)
			}
			return visitType(owner, *t.Elem)
		case TypeSlice:
			if t.Elem == nil {
				return fmt.Errorf("%s uses malformed slice type %q", owner, t.Raw)
			}
			return visitType(owner, *t.Elem)
		case TypeOptional:
			if t.Elem == nil {
				return fmt.Errorf("%s uses malformed optional type %q", owner, t.Raw)
			}
			return visitType(owner, *t.Elem)
		default:
			return nil
		}
	}

	visit = func(item *Struct) error {
		switch state[item.Name] {
		case 1:
			return fmt.Errorf("cyclic struct dependency involving %q", item.Name)
		case 2:
			return nil
		}
		state[item.Name] = 1
		for _, field := range item.Fields {
			if err := visitType("struct "+item.Name, field.Type); err != nil {
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

func (a *API) IsPOD(t TypeRef) bool {
	switch t.Kind {
	case TypePrimitive, TypeEnum:
		return true
	case TypeOptional:
		if t.Elem == nil {
			return false
		}
		return a.IsPOD(*t.Elem)
	case TypeSlice:
		if t.Elem == nil {
			return false
		}
		return a.IsPOD(*t.Elem)
	case TypeArray:
		if t.Elem == nil {
			return false
		}
		return a.IsPOD(*t.Elem)
	case TypeStruct:
		item := a.Struct(t.Name)
		if item == nil {
			return false
		}
		for _, field := range item.Fields {
			if !a.IsPOD(field.Type) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (a *API) SupportsSliceElem(t TypeRef) bool {
	switch t.Kind {
	case TypePrimitive, TypeEnum, TypeString, TypeBytes:
		return true
	case TypeOptional:
		if t.Elem == nil {
			return false
		}
		return a.IsPOD(*t.Elem)
	case TypeSlice:
		if t.Elem == nil {
			return false
		}
		return a.IsPOD(*t.Elem)
	case TypeArray:
		if t.Elem == nil {
			return false
		}
		return a.SupportsSliceElem(*t.Elem)
	case TypeStruct:
		item := a.Struct(t.Name)
		if item == nil {
			return false
		}
		for _, field := range item.Fields {
			if field.Type.Kind == TypeSlice {
				if field.Type.Elem == nil || !a.IsPOD(*field.Type.Elem) {
					return false
				}
				continue
			}
			if !a.SupportsSliceElem(field.Type) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (a *API) TypeNeedsKeepAlive(t TypeRef) bool {
	switch t.Kind {
	case TypeSlice:
		return true
	case TypeOptional:
		if t.Elem == nil {
			return false
		}
		return a.TypeNeedsKeepAlive(*t.Elem)
	case TypeArray:
		if t.Elem == nil {
			return false
		}
		return a.TypeNeedsKeepAlive(*t.Elem)
	case TypeStruct:
		item := a.Struct(t.Name)
		if item == nil {
			return false
		}
		for _, field := range item.Fields {
			if a.TypeNeedsKeepAlive(field.Type) {
				return true
			}
		}
	}
	return false
}
