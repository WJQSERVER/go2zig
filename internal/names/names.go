package names

import (
	"strings"
	"unicode"
)

var initialisms = map[string]string{
	"api":    "API",
	"cpu":    "CPU",
	"ffi":    "FFI",
	"go":     "Go",
	"guid":   "GUID",
	"http":   "HTTP",
	"https":  "HTTPS",
	"id":     "ID",
	"ip":     "IP",
	"json":   "JSON",
	"ok":     "OK",
	"qps":    "QPS",
	"rpc":    "RPC",
	"tls":    "TLS",
	"ui":     "UI",
	"uid":    "UID",
	"url":    "URL",
	"utf8":   "UTF8",
	"uuid":   "UUID",
	"xml":    "XML",
	"zig":    "Zig",
	"go2zig": "Go2Zig",
}

func Exported(name string) string {
	parts := split(name)
	if len(parts) == 0 {
		return "X"
	}
	var b strings.Builder
	for _, part := range parts {
		lower := strings.ToLower(part)
		if init, ok := initialisms[lower]; ok {
			b.WriteString(init)
			continue
		}
		runes := []rune(lower)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	if b.Len() == 0 {
		return "X"
	}
	return b.String()
}

func LowerCamel(name string) string {
	exported := Exported(name)
	runes := []rune(exported)
	if len(runes) == 0 {
		return "x"
	}
	if len(runes) > 1 && unicode.IsUpper(runes[0]) && unicode.IsUpper(runes[1]) {
		for i := 0; i < len(runes); i++ {
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if i > 0 && nextLower {
				runes[i] = unicode.ToUpper(runes[i])
				break
			}
			runes[i] = unicode.ToLower(runes[i])
			if nextLower {
				break
			}
		}
		return string(runes)
	}
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func split(name string) []string {
	var parts []string
	var current []rune
	flush := func() {
		if len(current) == 0 {
			return
		}
		parts = append(parts, string(current))
		current = current[:0]
	}

	var prev rune
	for i, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			flush()
			prev = 0
			continue
		}
		if i > 0 && len(current) > 0 {
			nextBreak := unicode.IsLower(prev) && unicode.IsUpper(r)
			if !nextBreak && unicode.IsDigit(r) != unicode.IsDigit(prev) {
				nextBreak = true
			}
			if nextBreak {
				flush()
			}
		}
		current = append(current, r)
		prev = r
	}
	flush()
	return parts
}
