package dcdmaker

import (
	"strings"
)

func isDCDValid(dcd string) (bool, string) {
	dcd = strings.TrimSpace(dcd)

	if len(dcd) == 0 {
		return false, "empty output"
	}

	hasSection := strings.Contains(dcd, "[section")
	hasBody := strings.Contains(dcd, "--- BODY ---")

	if !hasSection && !hasBody {
		return false, "missing [section] or --- BODY ---"
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


