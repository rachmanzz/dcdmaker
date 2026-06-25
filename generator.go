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

	tags := []struct{ open, close string }{
		{"<loop", "</loop"},
		{"<table", "</table"},
		{"<ul>", "</ul>"},
		{"<ol>", "</ol>"},
	}
	for _, t := range tags {
		if strings.Count(dcd, t.open) != strings.Count(dcd, t.close) {
			return false
		}
	}

	if strings.Count(dcd, "{{") != strings.Count(dcd, "}}") {
		return false
	}

	return true
}

func isTruncated(dcd string) bool {
	dcd = strings.TrimSpace(dcd)
	if strings.HasSuffix(dcd, "<TRUNCATED/>") {
		return true
	}

	mainSections := strings.Count(dcd, "[section")
	mainBodies := strings.Count(dcd, "--- BODY ---")
	if mainSections != mainBodies {
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

	return false
}

func sanitizeDCD(dcd string) string {
	dcd = strings.TrimSpace(dcd)

	if strings.HasPrefix(dcd, "```") {
		lines := strings.Split(dcd, "\n")
		if len(lines) >= 3 && strings.HasPrefix(lines[0], "```") {
			dcd = strings.Join(lines[1:], "\n")
			if strings.HasSuffix(dcd, "```") {
				dcd = strings.TrimSuffix(dcd, "```")
				dcd = strings.TrimSuffix(dcd, "\n")
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
