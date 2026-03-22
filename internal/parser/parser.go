package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"go2zig/internal/model"
)

var (
	structPattern = regexp.MustCompile(`(?s)pub\s+const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*extern\s+struct\s*\{(.*?)\}\s*;`)
	funcPattern   = regexp.MustCompile(`(?:pub\s+)?(?:extern|export)\s+fn\s+([A-Za-z_][A-Za-z0-9_]*)\s*\((.*?)\)\s*([A-Za-z_][A-Za-z0-9_\.]*)\s*(?:;|\{)`)
)

func ParseFile(path string) (*model.API, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read zig api %q: %w", path, err)
	}
	return Parse(string(content))
}

func Parse(content string) (*model.API, error) {
	clean := stripComments(content)
	structs, err := parseStructs(clean)
	if err != nil {
		return nil, err
	}
	funcs, err := parseFunctions(clean)
	if err != nil {
		return nil, err
	}
	return model.New(structs, funcs)
}

func parseStructs(content string) ([]*model.Struct, error) {
	matches := structPattern.FindAllStringSubmatch(content, -1)
	structs := make([]*model.Struct, 0, len(matches))
	for _, match := range matches {
		if match[1] == "String" || match[1] == "Bytes" {
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

func parseFunctions(content string) ([]*model.Function, error) {
	matches := funcPattern.FindAllStringSubmatch(content, -1)
	funcs := make([]*model.Function, 0, len(matches))
	for _, match := range matches {
		params, err := parseFields(match[2])
		if err != nil {
			return nil, fmt.Errorf("parse function %q params: %w", match[1], err)
		}
		ret, err := parseType(strings.TrimSpace(match[3]))
		if err != nil {
			return nil, fmt.Errorf("parse function %q return: %w", match[1], err)
		}
		funcs = append(funcs, &model.Function{Name: match[1], Params: params, Return: ret})
	}
	return funcs, nil
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
