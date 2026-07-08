package dcdmaker

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var reWrongLoop = regexp.MustCompile(`<(ol|ul)\s+\w+\s+from\s+`)

func isDCDValid(dcd string) (bool, string) {
	dcd = strings.TrimSpace(dcd)

	if len(dcd) == 0 {
		return false, "empty output"
	}

	if m := reWrongLoop.FindString(dcd); m != "" {
		return false, fmt.Sprintf("use <loop:ol> or <loop:ul> instead of plain tag: found %q", m)
	}

	hasSection := strings.Contains(dcd, "[section")
	hasBody := strings.Contains(dcd, "--- BODY ---")

	if !hasSection && !hasBody {
		return false, "missing [section] or --- BODY ---"
	}

	tags := []struct{ open, close string }{
		{"<loop", "</loop"},
		{"<table", "</table"},
		{"<ul", "</ul"},
		{"<ol", "</ol"},
	}
	for _, t := range tags {
		o := strings.Count(dcd, t.open)
		c := strings.Count(dcd, t.close)
		if o != c {
			return false, fmt.Sprintf("unbalanced %s: %d open, %d close", t.open, o, c)
		}
	}

	openCurl := strings.Count(dcd, "{{")
	closeCurl := strings.Count(dcd, "}}")
	if openCurl != closeCurl {
		return false, fmt.Sprintf("unbalanced {{ }}: %d open, %d close", openCurl, closeCurl)
	}

	sections := parseSections(dcd)
	usages := scanBody(dcd)

	usedVars := map[string]bool{}
	for _, u := range usages {
		usedVars[u.Var] = true
	}

	for _, s := range sections {
		if s.Name == "" {
			continue
		}
		for _, v := range s.Vars {
			if !usedVars[v] {
				return false, fmt.Sprintf("section %q declares var %q but never uses it", s.Name, v)
			}
		}
	}

	return true, ""
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
		{"<ul", "</ul"},
		{"<ol", "</ol"},
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

func isIncomplete(dcd string) bool {
	dcd = strings.TrimSpace(dcd)

	lines := strings.Split(dcd, "\n")
	if len(lines) < 30 {
		return true
	}

	sectionCount := strings.Count(dcd, "[section")
	bodyCount := strings.Count(dcd, "--- BODY ---")

	hasUnpredictable := strings.Contains(dcd, "[object-unpredictable]") ||
		strings.Contains(dcd, "[keys-unpredictable]")

	if hasUnpredictable {
		objs := parseUnpredictableObjects(dcd)
		keys := parseUnpredictableKeys(dcd)
		if len(objs) == 0 && len(keys) == 0 {
			return true
		}
		return false
	}

	if sectionCount >= 1 && bodyCount < sectionCount {
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
