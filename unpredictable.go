package dcdmaker

import (
	"regexp"
	"strings"
)

type UnpredictableObject struct {
	Name    string
	Fields  []string
	IsArray bool
}

var reObjectLine = regexp.MustCompile(`^(?:\-\s)?(\w+?)(\[\])?=(.+)$`)

func parseUnpredictableObjects(dcd string) []UnpredictableObject {
	section := extractSection(dcd, "[object-unpredictable]")
	if section == "" {
		return nil
	}

	var out []UnpredictableObject
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "#") {
			continue
		}
		m := reObjectLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		out = append(out, UnpredictableObject{
			Name:    m[1],
			Fields:  splitFields(m[3]),
			IsArray: m[2] == "[]",
		})
	}
	return out
}

func parseUnpredictableKeys(dcd string) []string {
	section := extractSection(dcd, "[keys-unpredictable]")
	if section == "" {
		return nil
	}

	var out []string
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			line = strings.TrimPrefix(line, "- ")
		}
		out = append(out, splitFields(line)...)
	}
	return out
}

func extractSection(dcd, header string) string {
	idx := strings.Index(dcd, header)
	if idx < 0 {
		return ""
	}
	rest := dcd[idx+len(header):]
	end := strings.Index(rest, "\n[")
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

func splitFields(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
