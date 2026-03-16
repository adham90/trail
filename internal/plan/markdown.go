package plan

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// TaskInfo holds parsed info about a top-level task under ## Tasks.
type TaskInfo struct {
	Index   int    // 1-based task number
	Text    string // Full text of the task line (after checkbox)
	Done    bool   // true if [x] or [X]
	Blocked bool   // true if text contains "[blocked]" marker
	LineNum int    // 0-based line number in the file
}

// PlanStatus holds parsed status information for a plan file.
type PlanStatus struct {
	Name      string
	FilePath  string
	Tasks     []TaskInfo
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

// ParseTasks finds the ## Tasks section and parses top-level checkboxes.
// Only top-level items (lines starting with "- [ ]" or "- [x]") are parsed.
// Sub-items (indented lines) are ignored.
func ParseTasks(data []byte) []TaskInfo {
	lines := strings.Split(string(data), "\n")
	inTasks := false
	var tasks []TaskInfo
	taskNum := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for ## Tasks heading (case-insensitive)
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
		if taskLineRe.MatchString(line) {
			taskNum++
			match := taskLineRe.FindStringSubmatch(line)
			done := match[1] == "x" || match[1] == "X"
			text := strings.TrimSpace(line[len(match[0]):])
			blocked := strings.Contains(strings.ToLower(text), "[blocked")

			tasks = append(tasks, TaskInfo{
				Index:   taskNum,
				Text:    text,
				Done:    done,
				Blocked: blocked,
				LineNum: i,
			})
		}
	}

	return tasks
}

// ParsePlanStatus reads a plan file and returns its status.
func ParsePlanStatus(path string) (*PlanStatus, error) {
	data, err := readFileBytes(path)
	if err != nil {
		return nil, err
	}

	title := ParseTitle(data)
	tasks := ParseTasks(data)

	doneCount := 0
	for _, t := range tasks {
		if t.Done {
			doneCount++
		}
	}

	return &PlanStatus{
		Name:      title,
		FilePath:  path,
		Tasks:     tasks,
		DoneCount: doneCount,
		Total:     len(tasks),
	}, nil
}

// SetTaskDone marks a task as done by flipping [ ] → [x]. taskNum is 1-based.
func SetTaskDone(data []byte, taskNum int) ([]byte, error) {
	lines := strings.Split(string(data), "\n")
	inTasks := false
	current := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inTasks {
			if isTasksHeading(trimmed) {
				inTasks = true
			}
			continue
		}

		if strings.HasPrefix(trimmed, "## ") {
			break
		}

		if taskLineRe.MatchString(line) {
			current++
			if current == taskNum {
				lines[i] = strings.Replace(line, "- [ ] ", "- [x] ", 1)
				return []byte(strings.Join(lines, "\n")), nil
			}
		}
	}

	return nil, fmt.Errorf("task %d not found (plan has %d tasks)", taskNum, current)
}

// SetTaskBlocked marks a task with a [blocked] annotation. taskNum is 1-based.
func SetTaskBlocked(data []byte, taskNum int, reason string) ([]byte, error) {
	lines := strings.Split(string(data), "\n")
	inTasks := false
	current := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inTasks {
			if isTasksHeading(trimmed) {
				inTasks = true
			}
			continue
		}

		if strings.HasPrefix(trimmed, "## ") {
			break
		}

		if taskLineRe.MatchString(line) {
			current++
			if current == taskNum {
				// Remove existing [blocked] annotation if present
				text := taskLineRe.ReplaceAllString(line, "")
				text = removeBlockedAnnotation(text)
				annotation := fmt.Sprintf(" [blocked: %s]", reason)
				lines[i] = "- [ ] " + text + annotation
				return []byte(strings.Join(lines, "\n")), nil
			}
		}
	}

	return nil, fmt.Errorf("task %d not found (plan has %d tasks)", taskNum, current)
}

// isTasksHeading checks if a line is a ## Tasks heading (case-insensitive).
func isTasksHeading(line string) bool {
	lower := strings.ToLower(line)
	return lower == "## tasks" || strings.HasPrefix(lower, "## tasks ")
}

var blockedAnnotationRe = regexp.MustCompile(`\s*\[blocked:.*?\]`)

func removeBlockedAnnotation(text string) string {
	return blockedAnnotationRe.ReplaceAllString(text, "")
}
