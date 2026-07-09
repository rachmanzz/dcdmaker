package dcdmaker

import (
	"fmt"
	"regexp"
	"strings"
)

type sectionInfo struct {
	Name string
	Vars []string
	Keys []string
	Body string
	raw  string
}

var (
	reSectionHeader = regexp.MustCompile(`^\[section(?::next-page)?\s+\d+\]`)
	reVarField      = regexp.MustCompile(`\{\{(\w+)\.(\w+)\}\}`)
	rePlainVar      = regexp.MustCompile(`\{\{(\w+)\}\}`)
	reLoopFrom      = regexp.MustCompile(`<loop(?::\w+)?(?:\s+\w+=\w+)*\s+\w+\s+from\s+(\w+)>`)
)

func parseSections(dcd string) []sectionInfo {
	lines := strings.Split(dcd, "\n")
	var sections []sectionInfo
	var cur *sectionInfo
	inBody := false

	for _, line := range lines {
		if reSectionHeader.MatchString(line) {
			cur = &sectionInfo{raw: line}
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
			}
		} else {
			if cur.Body != "" {
				cur.Body += "\n"
			}
			cur.Body += line
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

func validateVarsAndKeys(dcd string) error {
	sections := parseSections(dcd)
	usages := scanBody(dcd)

	declaredVars := map[string][]string{}   // varName -> [field1, field2, ...]
	declaredKeys := map[string]bool{}
	var declaredVarList []string
	var declaredKeyList []string

	for _, s := range sections {
		for _, v := range s.Vars {
			if _, ok := declaredVars[v]; !ok {
				declaredVars[v] = nil
				declaredVarList = append(declaredVarList, v)
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

	var errs []string

	for _, v := range declaredVarList {
		if !usedVars[v] {
			errs = append(errs, fmt.Sprintf("declared var %q is not used in body", v))
		}
	}

	for v := range usedVars {
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
		return fmt.Errorf("dcd validation:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
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
			newSection.WriteString(line + "\n")
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
		endIdx := strings.Index(rest, "\n\n")
		if endIdx < 0 {
			endIdx = len(rest)
		} else {
			endIdx += 2 // include the blank line
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
					newSection.WriteString("- " + strings.Join(kept, ", ") + "\n")
				} else {
					newSection.WriteString(strings.Join(kept, ", ") + "\n")
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

func fixVarsAndKeys(dcd string) string {
	sections := parseSections(dcd)
	usages := scanBody(dcd)

	declaredVars := map[string]bool{}
	for _, s := range sections {
		for _, v := range s.Vars {
			declaredVars[v] = true
		}
	}

	usedVars := map[string]bool{}
	usedVarFields := map[string]map[string]bool{}

	for _, u := range usages {
		usedVars[u.Var] = true
		if u.Field != "" {
			if usedVarFields[u.Var] == nil {
				usedVarFields[u.Var] = map[string]bool{}
			}
			usedVarFields[u.Var][u.Field] = true
		}
	}

	addVars := map[string]bool{}
	for v := range usedVars {
		if !declaredVars[v] {
			addVars[v] = true
		}
	}

	removeVars := map[string]bool{}
	for _, s := range sections {
		for _, v := range s.Vars {
			if !usedVars[v] {
				removeVars[v] = true
			}
		}
	}

	removeKeys := map[string]bool{}
	for _, s := range sections {
		for _, k := range s.Keys {
			found := false
			for v := range usedVarFields {
				if usedVarFields[v][k] {
					found = true
					break
				}
			}
			if !found {
				removeKeys[k] = true
			}
		}
	}

	if len(addVars) == 0 && len(removeVars) == 0 && len(removeKeys) == 0 {
		return dcd
	}

	lines := strings.Split(dcd, "\n")
	var secName string

	for i, line := range lines {
		trim := strings.TrimSpace(line)

		if reSectionHeader.MatchString(trim) {
			secName = ""
			continue
		}
		if strings.HasPrefix(trim, "name=") {
			secName = strings.TrimPrefix(trim, "name=")
			continue
		}

		if strings.HasPrefix(trim, "var=") {
			existing := splitCSV(strings.TrimPrefix(trim, "var="))
			var kept []string
			for _, v := range existing {
				if !removeVars[v] {
					kept = append(kept, v)
				}
			}
			for v := range addVars {
				if !contains(kept, v) {
					kept = append(kept, v)
				}
			}
			lines[i] = "var=" + strings.Join(kept, ", ")
			continue
		}

		if strings.HasPrefix(trim, "keys=") {
			existing := splitCSV(strings.TrimPrefix(trim, "keys="))
			var keot []string
			for _, k := range existing {
				if !removeKeys[k] {
					keot = append(keot, k)
				}
			}
			// Add missing fields from this section's vars
			for _, s := range sections {
				if s.Name == secName {
					for _, v := range s.Vars {
						if removeVars[v] {
							continue
						}
						for f := range usedVarFields[v] {
							if !contains(keot, f) {
								keot = append(keot, f)
							}
						}
					}
					break
				}
			}
			lines[i] = "keys=" + strings.Join(keot, ", ")
			continue
		}
	}

	return strings.Join(lines, "\n")
}
