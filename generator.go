package dcdmaker

import (
	"fmt"
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

	sectionCount := strings.Count(dcd, "[section")
	bodyCount := strings.Count(dcd, "--- BODY ---")
	if sectionCount != bodyCount {
		return false, fmt.Sprintf("section/body mismatch: %d sections, %d bodies", sectionCount, bodyCount)
	}

	return true, ""
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


