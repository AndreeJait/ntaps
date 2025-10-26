package util

import (
	"regexp"
	"strings"
)

var wordRe = regexp.MustCompile(`[A-Za-z0-9]+`)

func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := wordRe.FindAllString(s, -1)
	if len(parts) == 0 {
		return ""
	}
	for i, p := range parts {
		if len(p) == 1 {
			parts[i] = strings.ToUpper(p)
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}

func ToCamelCase(s string) string {
	p := ToPascalCase(s)
	if p == "" {
		return ""
	}
	return strings.ToLower(p[:1]) + p[1:]
}

func HumanizePascal(s string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	return strings.TrimSpace(re.ReplaceAllString(s, "$1 $2"))
}

func Conditional(ok bool, val string) string {
	if ok {
		return val
	}
	return ""
}

func RouterPath(pkg, endpointType, endpoint string) string {
	switch strings.ToLower(endpointType) {
	case "internal":
		return "/internal/" + pkg + endpoint
	default:
		return "/" + pkg + endpoint
	}
}
