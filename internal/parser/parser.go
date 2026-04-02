package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"go2zig/internal/model"
)

var (
	structPattern     = regexp.MustCompile(`(?s)pub\s+const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*extern\s+struct\s*\{(.*?)\}\s*;`)
	enumPattern       = regexp.MustCompile(`(?s)pub\s+const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*enum\s*\(([^)]+)\)\s*\{(.*?)\}\s*;`)
	slicePattern      = regexp.MustCompile(`(?s)pub\s+const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*extern\s+struct\s*\{\s*ptr\s*:\s*\?\s*\[\*\]const\s+([^,]+),\s*len\s*:\s*usize\s*,?\s*\}\s*;`)
	arrayAliasPattern = regexp.MustCompile(`(?m)^\s*pub\s+const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(\[[^;=]+)\s*;\s*$`)
	funcPattern       = regexp.MustCompile(`(?s)(?:pub\s+)?(?:extern|export)\s+fn\s+([A-Za-z_][A-Za-z0-9_]*)\s*\((.*?)\)\s*((?:error\s*\{[^}]*\}\s*!|[A-Za-z_][A-Za-z0-9_\.]*(?:\s*!\s*))?\?*(?:\[\d+\])*[A-Za-z_][A-Za-z0-9_\.]*)\s*(?:;|\{)`)
	funcDeclPattern   = regexp.MustCompile(`^(?:pub\s+)?(?:extern|export)\s+fn\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
	arrayPattern      = regexp.MustCompile(`^\[(\d+)\](.+)$`)
)

const codegenDirectivePrefix = "//go2zig:"

func ParseFile(path string) (*model.API, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read zig api %q: %w", path, err)
	}
	return Parse(string(content))
}

func Parse(content string) (*model.API, error) {
	directives, err := parseCodegenDirectives(content)
	if err != nil {
		return nil, err
	}
	clean := stripComments(content)
	structs, err := parseStructs(clean)
	if err != nil {
		return nil, err
	}
	enums, err := parseEnums(clean)
	if err != nil {
		return nil, err
	}
	slices, err := parseSlices(clean)
	if err != nil {
		return nil, err
	}
	arrays, err := parseArrayAliases(clean)
	if err != nil {
		return nil, err
	}
	funcs, err := parseFunctions(clean, directives)
	if err != nil {
		return nil, err
	}
	return model.New(structs, enums, slices, arrays, funcs)
}

func parseCodegenDirectives(content string) (map[string]model.FunctionCodegen, error) {
	directives := map[string]model.FunctionCodegen{}
	pending := []string{}
	var decl strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))
	maxTokenSize := len(content)
	if maxTokenSize < 64*1024 {
		maxTokenSize = 64 * 1024
	}
	scanner.Buffer(make([]byte, 0, 64*1024), maxTokenSize)
	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(trimmed, "///") || strings.HasPrefix(trimmed, "//!") {
			continue
		}
		if strings.HasPrefix(trimmed, codegenDirectivePrefix) {
			pending = append(pending, strings.TrimSpace(strings.TrimPrefix(trimmed, codegenDirectivePrefix)))
			continue
		}
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if len(pending) == 0 {
			continue
		}
		if decl.Len() > 0 {
			decl.WriteByte(' ')
		}
		decl.WriteString(trimmed)
		name, ok := parseFunctionNameFromDecl(decl.String())
		if !ok {
			continue
		}
		cfg := directives[name]
		for _, raw := range pending {
			if err := applyCodegenDirective(&cfg, raw); err != nil {
				return nil, fmt.Errorf("function %q: %w", name, err)
			}
		}
		directives[name] = cfg
		pending = nil
		decl.Reset()
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan zig api directives: %w", err)
	}
	if len(pending) > 0 {
		return nil, fmt.Errorf("codegen directive must be attached to a function declaration")
	}
	return directives, nil
}

func parseFunctionNameFromDecl(decl string) (string, bool) {
	match := funcDeclPattern.FindStringSubmatch(decl)
	if match == nil {
		return "", false
	}
	return match[1], true
}

func applyCodegenDirective(cfg *model.FunctionCodegen, raw string) error {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return fmt.Errorf("empty codegen directive")
	}
	switch fields[0] {
	case "bridge-call":
		if len(fields) != 2 {
			return fmt.Errorf("bridge-call directive expects exactly one value")
		}
		switch fields[1] {
		case "inline":
			return cfg.SetBridgeCallHint(model.CallHintInline)
		case "noinline":
			return cfg.SetBridgeCallHint(model.CallHintNoInline)
		default:
			return fmt.Errorf("unsupported bridge-call value %q", fields[1])
		}
	case "go-noinline":
		if len(fields) != 1 {
			return fmt.Errorf("go-noinline directive does not take values")
		}
		cfg.GoNoInline = true
	default:
		return fmt.Errorf("unsupported codegen directive %q", fields[0])
	}
	return nil
}

func parseStructs(content string) ([]*model.Struct, error) {
	matches := structPattern.FindAllStringSubmatch(content, -1)
	structs := make([]*model.Struct, 0, len(matches))
	for _, match := range matches {
		if match[1] == "String" || match[1] == "Bytes" || isSliceStruct(match[2]) {
			continue
		}
		fields, err := parseFields(match[2])
		if err != nil {
			return nil, fmt.Errorf("parse struct %q: %w", match[1], err)
		}
		structs = append(structs, &model.Struct{Name: match[1], Fields: fields})
	}
	return structs, nil
}

func parseEnums(content string) ([]*model.Enum, error) {
	matches := enumPattern.FindAllStringSubmatch(content, -1)
	enums := make([]*model.Enum, 0, len(matches))
	for _, match := range matches {
		values, err := parseEnumValues(match[3])
		if err != nil {
			return nil, fmt.Errorf("parse enum %q: %w", match[1], err)
		}
		enums = append(enums, &model.Enum{Name: match[1], BaseName: strings.TrimSpace(match[2]), Values: values})
	}
	return enums, nil
}

func parseSlices(content string) ([]*model.Slice, error) {
	matches := slicePattern.FindAllStringSubmatch(content, -1)
	slices := make([]*model.Slice, 0, len(matches))
	for _, match := range matches {
		name := match[1]
		if name == "String" || name == "Bytes" {
			continue
		}
		elem, err := parseType(strings.TrimSpace(match[2]))
		if err != nil {
			return nil, fmt.Errorf("parse slice %q: %w", name, err)
		}
		slices = append(slices, &model.Slice{Name: name, Elem: elem})
	}
	return slices, nil
}

func parseArrayAliases(content string) ([]*model.ArrayAlias, error) {
	matches := arrayAliasPattern.FindAllStringSubmatch(content, -1)
	aliases := make([]*model.ArrayAlias, 0, len(matches))
	for _, match := range matches {
		name := match[1]
		typ, err := parseType(strings.TrimSpace(match[2]))
		if err != nil {
			return nil, fmt.Errorf("parse array alias %q: %w", name, err)
		}
		if typ.Kind != model.TypeArray {
			continue
		}
		aliases = append(aliases, &model.ArrayAlias{Name: name, Type: typ})
	}
	return aliases, nil
}

func isSliceStruct(body string) bool {
	body = strings.TrimSpace(body)
	compact := strings.ReplaceAll(body, " ", "")
	compact = strings.ReplaceAll(compact, "\n", "")
	compact = strings.ReplaceAll(compact, "\r", "")
	compact = strings.ReplaceAll(compact, "\t", "")
	return strings.HasPrefix(compact, "ptr:?[*]const") && strings.Contains(compact, ",len:usize")
}

func parseEnumValues(body string) ([]model.EnumValue, error) {
	parts := strings.Split(body, ",")
	values := make([]model.EnumValue, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name := part
		value := ""
		if idx := strings.Index(part, "="); idx >= 0 {
			name = strings.TrimSpace(part[:idx])
			value = strings.TrimSpace(part[idx+1:])
		}
		if !isIdent(name) {
			return nil, fmt.Errorf("invalid enum value %q", part)
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("duplicate enum value %q", name)
		}
		seen[name] = struct{}{}
		values = append(values, model.EnumValue{Name: name, Value: value})
	}
	return values, nil
}

func parseFunctions(content string, directives map[string]model.FunctionCodegen) ([]*model.Function, error) {
	matches := funcPattern.FindAllStringSubmatch(content, -1)
	funcs := make([]*model.Function, 0, len(matches))
	for _, match := range matches {
		params, err := parseFields(match[2])
		if err != nil {
			return nil, fmt.Errorf("parse function %q params: %w", match[1], err)
		}
		ret, canErr, err := parseReturnType(strings.TrimSpace(match[3]))
		if err != nil {
			return nil, fmt.Errorf("parse function %q return: %w", match[1], err)
		}
		funcs = append(funcs, &model.Function{Name: match[1], Params: params, Return: ret, CanErr: canErr, Codegen: directives[match[1]]})
	}
	return funcs, nil
}

func parseReturnType(raw string) (model.TypeRef, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return model.TypeRef{}, false, fmt.Errorf("return type is empty")
	}
	if strings.Contains(raw, "!") {
		idx := strings.Index(raw, "!")
		payload := strings.TrimSpace(raw[idx+1:])
		if payload == "" {
			return model.TypeRef{}, false, fmt.Errorf("error union payload is empty")
		}
		t, err := parseType(payload)
		return t, true, err
	}
	t, err := parseType(raw)
	return t, false, err
}

func parseFields(body string) ([]model.Field, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, nil
	}
	parts := strings.Split(body, ",")
	fields := make([]model.Field, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		pieces := strings.SplitN(part, ":", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("invalid field declaration %q", part)
		}
		name := strings.TrimSpace(pieces[0])
		if name == "" {
			return nil, fmt.Errorf("field name is empty in %q", part)
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("duplicate field %q", name)
		}
		seen[name] = struct{}{}
		typ, err := parseType(strings.TrimSpace(pieces[1]))
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", name, err)
		}
		fields = append(fields, model.Field{Name: name, Type: typ})
	}
	return fields, nil
}

func parseType(raw string) (model.TypeRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return model.TypeRef{}, fmt.Errorf("type is empty")
	}
	if strings.HasPrefix(raw, "?") {
		elem, err := parseType(strings.TrimSpace(strings.TrimPrefix(raw, "?")))
		if err != nil {
			return model.TypeRef{}, err
		}
		return model.TypeRef{Kind: model.TypeOptional, Raw: raw, Elem: &elem}, nil
	}
	if match := arrayPattern.FindStringSubmatch(raw); match != nil {
		length, err := strconv.Atoi(match[1])
		if err != nil {
			return model.TypeRef{}, fmt.Errorf("invalid array length %q", match[1])
		}
		elem, err := parseType(strings.TrimSpace(match[2]))
		if err != nil {
			return model.TypeRef{}, err
		}
		return model.TypeRef{Kind: model.TypeArray, Raw: raw, Elem: &elem, ArrayLen: length}, nil
	}
	raw = strings.TrimPrefix(raw, "[*]const ")
	raw = strings.TrimPrefix(raw, "[*]")
	base := raw
	if idx := strings.LastIndex(base, "."); idx >= 0 {
		base = base[idx+1:]
	}
	switch base {
	case "void":
		return model.TypeRef{Kind: model.TypeVoid, Raw: raw}, nil
	case "String":
		return model.TypeRef{Kind: model.TypeString, Raw: raw, Name: base}, nil
	case "Bytes":
		return model.TypeRef{Kind: model.TypeBytes, Raw: raw, Name: base}, nil
	case "GoReader":
		return model.TypeRef{Kind: model.TypeGoReader, Raw: raw, Name: base}, nil
	case "GoWriter":
		return model.TypeRef{Kind: model.TypeGoWriter, Raw: raw, Name: base}, nil
	}
	if prim, ok := model.Primitive(base); ok {
		return model.TypeRef{Kind: model.TypePrimitive, Raw: raw, Name: base, Primitive: prim}, nil
	}
	if !isIdent(base) {
		return model.TypeRef{}, fmt.Errorf("unsupported type %q", raw)
	}
	return model.TypeRef{Kind: model.TypeStruct, Raw: raw, Name: base}, nil
}

func stripComments(input string) string {
	var out strings.Builder
	inBlock := false
	for i := 0; i < len(input); i++ {
		if inBlock {
			if i+1 < len(input) && input[i] == '*' && input[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if i+1 < len(input) && input[i] == '/' && input[i+1] == '*' {
			inBlock = true
			i++
			continue
		}
		if i+1 < len(input) && input[i] == '/' && input[i+1] == '/' {
			for i < len(input) && input[i] != '\n' {
				i++
			}
			if i < len(input) {
				out.WriteByte('\n')
			}
			continue
		}
		out.WriteByte(input[i])
	}
	return out.String()
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !(r == '_' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
				return false
			}
			continue
		}
		if !(r == '_' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
