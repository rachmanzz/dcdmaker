package dcdmaker

import (
	"strings"
	"unicode"
)

func isDCDValid(dcd string) bool {
	dcd = strings.TrimSpace(dcd)

	if len(dcd) == 0 {
		return false
	}

	hasSection := strings.Contains(dcd, "[section")
	hasBody := strings.Contains(dcd, "--- BODY ---")

	if !hasSection && !hasBody {
		return false
	}

	tags := struct {
		open  string
		close string
		count int
	}{
		open:  "<loop",
		close: "</loop",
	}

	openCount := strings.Count(dcd, tags.open)
	closeCount := strings.Count(dcd, tags.close)
	if openCount != closeCount {
		return false
	}

	for _, pair := range []struct{ open, close string }{
		{"<table", "</table"},
		{"<ul>", "</ul>"},
		{"<ol>", "</ol>"},
	} {
		oc := strings.Count(dcd, pair.open)
		cc := strings.Count(dcd, pair.close)
		if oc != cc {
			return false
		}
	}

	if strings.Count(dcd, "{{") != strings.Count(dcd, "}}") {
		return false
	}

	lines := strings.Split(dcd, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "---") {
			continue
		}
		if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
			continue
		}
		if !containsAlpha(trimmed) && !strings.Contains(trimmed, "{{") {
			continue
		}
	}

	return true
}

func isTruncated(dcd string) bool {
	dcd = strings.TrimSpace(dcd)
	if strings.HasSuffix(dcd, "<TRUNCATED/>") {
		return true
	}

	lines := strings.Split(dcd, "\n")
	if len(lines) == 0 {
		return false
	}

	last := strings.TrimSpace(lines[len(lines)-1])

	if strings.Count(dcd, "[section") != strings.Count(dcd, "--- BODY ---") {
		return true
	}

	tags := []struct{ open, close string }{
		{"<loop", "</loop"},
		{"<table", "</table"},
		{"<ul>", "</ul>"},
		{"<ol>", "</ol>"},
	}
	for _, t := range tags {
		if strings.Count(dcd, t.open) != strings.Count(dcd, t.close) {
			return true
		}
	}

	if strings.Count(dcd, "{{") != strings.Count(dcd, "}}") {
		return true
	}

	_ = last
	return false
}

func sanitizeDCD(dcd string) string {
	dcd = strings.TrimSpace(dcd)

	if strings.HasPrefix(dcd, "```") {
		lines := strings.Split(dcd, "\n")
		if len(lines) >= 3 && strings.HasPrefix(lines[0], "```") {
			if strings.HasPrefix(lines[0], "```dcd") || strings.HasPrefix(lines[0], "```") {
				dcd = strings.Join(lines[1:], "\n")
				if strings.HasSuffix(dcd, "```") {
					dcd = strings.TrimSuffix(dcd, "```")
					dcd = strings.TrimSuffix(dcd, "\n")
				}
			}
		}
	}

	return strings.TrimSpace(dcd)
}

func containsAlpha(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
