package dcdmaker

import (
	"fmt"
	"regexp"
	"strings"
)

type sectionInfo struct {
	Name    string
	Vars    []string
	Keys    []string
	Formats []string
}

var (
	reSectionHeader = regexp.MustCompile(`^\[section(?::next-page)?\s+\d+\]`)
	reVarField      = regexp.MustCompile(`\{\{(\w+)\.(\w+)\}\}`)
	rePlainVar      = regexp.MustCompile(`\{\{(\w+)\}\}`)
	reLoopFrom      = regexp.MustCompile(`<loop(?::\w+)?(?:\s+[^\s=]+=[^\s]+)*\s+\w+\s+from\s+(\w+)>`)
	reLoopIter      = regexp.MustCompile(`<loop(?::\w+)?(?:\s+[^\s=]+=[^\s]+)*\s+(\w+)\s+from\s+\w+>`)
)

func parseSections(dcd string) []sectionInfo {
	lines := strings.Split(dcd, "\n")
	var sections []sectionInfo
	var cur *sectionInfo
	inBody := false

	for _, line := range lines {
		if reSectionHeader.MatchString(line) {
			cur = &sectionInfo{}
			sections = append(sections, *cur)
			cur = &sections[len(sections)-1]
			inBody = false
			continue
		}
		if cur == nil {
			continue
		}

		trim := strings.TrimSpace(line)
		if trim == "--- BODY ---" {
			inBody = true
			continue
		}

		if !inBody {
			switch {
			case strings.HasPrefix(trim, "name="):
				cur.Name = strings.TrimPrefix(trim, "name=")
			case strings.HasPrefix(trim, "var="):
				cur.Vars = splitCSV(strings.TrimPrefix(trim, "var="))
			case strings.HasPrefix(trim, "keys="):
				cur.Keys = splitCSV(strings.TrimPrefix(trim, "keys="))
			case strings.HasPrefix(trim, "formats="):
				cur.Formats = splitCSV(strings.TrimPrefix(trim, "formats="))
			}
		} else {
			// body line — we only need to scan body globally via scanBody(), not per-section
		}
	}
	return sections
}

func splitCSV(s string) []string {
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

type bodyUsage struct {
	Var    string
	Field  string // empty for plain {{var}}
	IsLoop bool   // true if sourced from <loop x from var>
}

func scanBody(dcd string) []bodyUsage {
	var out []bodyUsage
	seen := map[string]bool{}

	for _, m := range reVarField.FindAllStringSubmatch(dcd, -1) {
		key := m[1] + "." + m[2]
		if !seen[key] {
			seen[key] = true
			out = append(out, bodyUsage{Var: m[1], Field: m[2]})
		}
	}

	for _, m := range rePlainVar.FindAllStringSubmatch(dcd, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, bodyUsage{Var: m[1]})
		}
	}

	for _, m := range reLoopFrom.FindAllStringSubmatch(dcd, -1) {
		key := "loop:" + m[1]
		if !seen[key] {
			seen[key] = true
			out = append(out, bodyUsage{Var: m[1], IsLoop: true})
		}
	}

	return out
}

func validateVarsAndKeys(dcd string) (warnings []string, err error) {
	sections := parseSections(dcd)
	usages := scanBody(dcd)
	loopIterVars := collectLoopIterVars(dcd)

	declaredVars := map[string][]string{} // varName -> [field1, field2, ...]
	declaredKeys := map[string]bool{}
	var declaredVarList []string
	var declaredKeyList []string

	var errs []string

	for _, s := range sections {
		if len(s.Vars) > 3 {
			errs = append(errs, fmt.Sprintf("section %q has %d vars (max 3)", s.Name, len(s.Vars)))
		}
		if len(s.Keys) > 15 {
			errs = append(errs, fmt.Sprintf("section %q has %d keys (max 15)", s.Name, len(s.Keys)))
		}
		for _, v := range s.Vars {
			name := strings.TrimPrefix(v, "[]")
			if _, ok := declaredVars[name]; !ok {
				declaredVars[name] = nil
				declaredVarList = append(declaredVarList, name)
			}
		}
		for _, k := range s.Keys {
			if !declaredKeys[k] {
				declaredKeys[k] = true
				declaredKeyList = append(declaredKeyList, k)
			}
		}
	}

	usedVars := map[string]bool{}
	usedVarFields := map[string]map[string]bool{} // varName -> {field: true}
	loopVars := map[string]bool{}

	for _, u := range usages {
		usedVars[u.Var] = true
		if u.Field != "" {
			if usedVarFields[u.Var] == nil {
				usedVarFields[u.Var] = map[string]bool{}
			}
			usedVarFields[u.Var][u.Field] = true
		}
		if u.IsLoop {
			loopVars[u.Var] = true
		}
	}

	for _, v := range declaredVarList {
		if !usedVars[v] {
			warnings = append(warnings, fmt.Sprintf("declared var %q is not used in body", v))
		}
	}

	for v := range usedVars {
		if loopIterVars[v] {
			continue
		}
		if _, ok := declaredVars[v]; !ok {
			errs = append(errs, fmt.Sprintf("var %q used in body is not declared in any [section] var=", v))
		} else {
			if loopVars[v] && !isLoopVarInSection(v, sections) {
				// loop source should be fine even without special handling
			}
			for f := range usedVarFields[v] {
				found := false
				for _, s := range sections {
					if contains(s.Vars, v) && contains(s.Keys, f) {
						found = true
						break
					}
				}
				if !found {
					errs = append(errs, fmt.Sprintf("field %q of var %q used in body but not in any keys=", f, v))
				}
			}
		}
	}

	if len(errs) > 0 {
		return warnings, fmt.Errorf("dcd validation:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return warnings, nil
}

func collectLoopIterVars(dcd string) map[string]bool {
	result := map[string]bool{}
	for _, m := range reLoopIter.FindAllStringSubmatch(dcd, -1) {
		result[m[1]] = true
	}
	return result
}

func ValidationErrorCount(err error) int {
	if err == nil {
		return 0
	}
	return strings.Count(err.Error(), "\n  - ")
}

type CategorizedErrors struct {
	SectionLimits []string
	MissingKeys   []string
	UndeclaredVars []string
	Other         []string
}

func categorizeErrors(err error) CategorizedErrors {
	var ce CategorizedErrors
	if err == nil {
		return ce
	}
	for _, line := range strings.Split(err.Error(), "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		if line == "" || strings.HasPrefix(line, "dcd validation") {
			continue
		}
		switch {
		case strings.Contains(line, "has ") && strings.Contains(line, "vars (max") ||
			strings.Contains(line, "has ") && strings.Contains(line, "keys (max"):
			ce.SectionLimits = append(ce.SectionLimits, line)
		case strings.Contains(line, "used in body but not in any keys="):
			ce.MissingKeys = append(ce.MissingKeys, line)
		case strings.Contains(line, "used in body is not declared"):
			ce.UndeclaredVars = append(ce.UndeclaredVars, line)
		default:
			ce.Other = append(ce.Other, line)
		}
	}
	return ce
}

func buildRetryFeedback(err error, result string, attempt int) string {
	ce := categorizeErrors(err)
	errCount := ValidationErrorCount(err)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n\nThe previous attempt had %d validation error(s):\n\n", errCount))

	if len(ce.SectionLimits) > 0 {
		b.WriteString("SECTION LIMITS VIOLATED (§7 — max 3 vars, max 15 keys per section):\n")
		for _, e := range ce.SectionLimits {
			b.WriteString(fmt.Sprintf("  - %s\n", e))
		}
		b.WriteString("FIX: Split the section into multiple [section N] blocks.\n\n")
	}

	if len(ce.MissingKeys) > 0 {
		b.WriteString("FIELDS USED BUT NOT DECLARED IN keys= (§7C — every {{var.field}} in body must be in keys=):\n")
		for _, e := range ce.MissingKeys {
			b.WriteString(fmt.Sprintf("  - %s\n", e))
		}
		b.WriteString("FIX: Add the missing field names to the keys= attribute of the section that declares the var.\n\n")
	}

	if len(ce.UndeclaredVars) > 0 {
		b.WriteString("VARS USED BUT NOT DECLARED (§7C — every {{var}} in body must be in var= or [object-unpredictable]):\n")
		for _, e := range ce.UndeclaredVars {
			b.WriteString(fmt.Sprintf("  - %s\n", e))
		}
		b.WriteString("FIX: Declare the variable in var= of the appropriate section, or use [object-unpredictable].\n\n")
	}

	if len(ce.Other) > 0 {
		b.WriteString("OTHER ERRORS:\n")
		for _, e := range ce.Other {
			b.WriteString(fmt.Sprintf("  - %s\n", e))
		}
		b.WriteString("\n")
	}

	if attempt >= 1 {
		b.WriteString("OUTPUT WITH ERRORS (showing only sections that need fixing):\n---\n")
		sections := strings.Split(result, "[section")
		for _, sec := range sections {
			if sec == "" {
				continue
			}
			secName := ""
			if idx := strings.Index(sec, "]"); idx > 0 {
				secName = sec[:idx+1]
			}
			for _, e := range ce.SectionLimits {
				if strings.Contains(e, secName) || strings.Contains(sec, e[:min(len(e), 20)]) {
					b.WriteString("[section" + sec[:min(len(sec), 500)])
					if len(sec) > 500 {
						b.WriteString("...[truncated]")
					}
					b.WriteString("\n\n")
					break
				}
			}
		}
		b.WriteString("---\n\n")
	}

	b.WriteString("Regenerate the FULL valid DCD template, fixing ALL the issues above.\n")
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isLoopVarInSection(v string, sections []sectionInfo) bool {
	for _, s := range sections {
		if contains(s.Vars, v) {
			return true
		}
	}
	return false
}

// fixUnpredictableOverlap removes predicted names from [object-unpredictable] and
// [keys-unpredictable] sections. Object names matching predicted VarObject/VarArray names
// are removed. Key entries matching predicted names are removed. Field-level checks are
// NOT performed (fields like "nama" are scoped to their object and may repeat across objects).
func fixUnpredictableOverlap(dcd string, predictableKeys []KeyDef) string {
	if len(predictableKeys) == 0 {
		return dcd
	}

	predictedNames := map[string]bool{}
	for _, k := range predictableKeys {
		switch k.Type {
		case VarObject, VarArray:
			predictedNames[k.Name] = true
		case VarKeys:
			if k.FieldDefs != nil {
				for _, f := range k.FieldDefs {
					predictedNames[f.Name] = true
				}
			} else {
				for _, f := range k.Fields {
					predictedNames[f] = true
				}
			}
		}
	}

	result := dcd

	// Fix [object-unpredictable] — remove object lines whose name is predicted
	if objStart := strings.Index(result, "[object-unpredictable]"); objStart >= 0 {
		rest := result[objStart:]
		endIdx := strings.Index(rest, "\n[")
		if endIdx < 0 {
			endIdx = len(rest)
		}
		section := rest[:endIdx]

		var newSection strings.Builder
		newSection.WriteString("[object-unpredictable]\n")
		origLen := len("[object-unpredictable]\n")

		hasContent := false
		for _, line := range strings.Split(section[origLen:], "\n") {
			trim := strings.TrimSpace(line)
			if trim == "" {
				continue
			}
			m := reObjectLine.FindStringSubmatch(trim)
			if m != nil && predictedNames[m[1]] {
				continue
			}
			newSection.WriteString(line)
			newSection.WriteString("\n")
			hasContent = true
		}

		if !hasContent {
			result = result[:objStart] + strings.TrimPrefix(rest[endIdx:], "\n")
		} else {
			result = result[:objStart] + newSection.String() + rest[endIdx:]
		}
	}

	// Fix [keys-unpredictable] — remove key entries matching predicted names
	if keysStart := strings.Index(result, "[keys-unpredictable]"); keysStart >= 0 {
		rest := result[keysStart:]
		endIdx := strings.Index(rest, "\n[")
		if endIdx < 0 {
			endIdx = len(rest)
		}
		section := rest[:endIdx]

		var newSection strings.Builder
		newSection.WriteString("[keys-unpredictable]\n")
		origLen := len("[keys-unpredictable]\n")

		hasContent := false
		for _, line := range strings.Split(section[origLen:], "\n") {
			trim := strings.TrimSpace(line)
			if trim == "" {
				continue
			}
			keyLine := trim
			hasDash := strings.HasPrefix(keyLine, "- ")
			if hasDash {
				keyLine = strings.TrimPrefix(keyLine, "- ")
			}
			kept := make([]string, 0, len(keyLine)/2)
			for _, k := range splitFields(keyLine) {
				if !predictedNames[k] {
					kept = append(kept, k)
				}
			}
			if len(kept) > 0 {
				if hasDash {
					newSection.WriteString("- ")
					newSection.WriteString(strings.Join(kept, ", "))
					newSection.WriteString("\n")
				} else {
					newSection.WriteString(strings.Join(kept, ", "))
					newSection.WriteString("\n")
				}
				hasContent = true
			}
		}

		if !hasContent {
			result = result[:keysStart] + strings.TrimPrefix(rest[endIdx:], "\n")
		} else {
			result = result[:keysStart] + newSection.String() + rest[endIdx:]
		}
	}

	return result
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
