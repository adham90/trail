package plan

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

// PlanStatus holds parsed status information for a plan file.
type PlanStatus struct {
	Name      string
	FilePath  string
	DoneCount int
	Total     int
}

var taskLineRe = regexp.MustCompile(`^- \[([ xX])\] `)

// ParseTitle extracts the first # heading from the data.
func ParseTitle(data []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

// ParseTaskCounts finds the ## Tasks section and counts top-level checkboxes.
// Only lines starting with "- [ ]" or "- [x]" (no leading whitespace) are counted.
// Indented checkboxes (sub-tasks) are ignored.
func ParseTaskCounts(data []byte) (done, total int) {
	lines := strings.Split(string(data), "\n")
	inTasks := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inTasks {
			if isTasksHeading(trimmed) {
				inTasks = true
			}
			continue
		}

		// Stop at next ## heading
		if strings.HasPrefix(trimmed, "## ") {
			break
		}

		// Only match top-level checkboxes (no leading whitespace)
		if match := taskLineRe.FindStringSubmatch(line); match != nil {
			total++
			if match[1] == "x" || match[1] == "X" {
				done++
			}
		}
	}

	return done, total
}

// ParsePlanStatus reads a plan file and returns its status.
func ParsePlanStatus(path string) (*PlanStatus, error) {
	data, err := readFileBytes(path)
	if err != nil {
		return nil, err
	}

	title := ParseTitle(data)
	done, total := ParseTaskCounts(data)

	return &PlanStatus{
		Name:      title,
		FilePath:  path,
		DoneCount: done,
		Total:     total,
	}, nil
}

// isTasksHeading checks if a line is a ## Tasks heading (case-insensitive).
func isTasksHeading(line string) bool {
	lower := strings.ToLower(line)
	return lower == "## tasks" || strings.HasPrefix(lower, "## tasks ")
}
