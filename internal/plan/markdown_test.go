package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTitle(t *testing.T) {
	data := []byte("# My Great Plan\n\nSome goal.\n")
	got := ParseTitle(data)
	if got != "My Great Plan" {
		t.Errorf("ParseTitle = %q, want %q", got, "My Great Plan")
	}
}

func TestParseTitleMissing(t *testing.T) {
	data := []byte("No heading here\n")
	got := ParseTitle(data)
	if got != "" {
		t.Errorf("ParseTitle = %q, want empty", got)
	}
}

func TestParseTasks(t *testing.T) {
	data := []byte(`# Test Plan

Goal.

## Tasks

- [x] **1.** First task done
  - [x] 1.1. sub-step
- [ ] **2.** Second task pending
  Description text.
  - [ ] 2.1. verify step
- [ ] **3.** Third task

## Notes

Some notes.
`)
	tasks := ParseTasks(data)
	if len(tasks) != 3 {
		t.Fatalf("len(tasks) = %d, want 3", len(tasks))
	}

	if !tasks[0].Done {
		t.Error("task 1 should be done")
	}
	if tasks[0].Index != 1 {
		t.Errorf("task 1 index = %d, want 1", tasks[0].Index)
	}

	if tasks[1].Done {
		t.Error("task 2 should not be done")
	}
	if tasks[1].Index != 2 {
		t.Errorf("task 2 index = %d, want 2", tasks[1].Index)
	}

	if tasks[2].Done {
		t.Error("task 3 should not be done")
	}
}

func TestParseTasksIgnoresSubItems(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Top task
  - [ ] 1.1. sub
  - [x] 1.2. sub done
- [x] **2.** Done task
`)
	tasks := ParseTasks(data)
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
}

func TestParseTasksCaseInsensitiveHeading(t *testing.T) {
	data := []byte(`## tasks

- [ ] **1.** A task
`)
	tasks := ParseTasks(data)
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
}

func TestParseTasksStopsAtNextSection(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Real task

## Decisions

- [ ] Not a task (in decisions section)
`)
	tasks := ParseTasks(data)
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
}

func TestSetTaskDone(t *testing.T) {
	data := []byte(`# Plan

## Tasks

- [ ] **1.** First task
- [ ] **2.** Second task
- [ ] **3.** Third task
`)
	result, err := SetTaskDone(data, 2)
	if err != nil {
		t.Fatalf("SetTaskDone: %v", err)
	}

	lines := strings.Split(string(result), "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Second task") {
			if !strings.Contains(line, "- [x]") {
				t.Errorf("task 2 not marked done: %q", line)
			}
			found = true
		}
	}
	if !found {
		t.Error("task 2 text not found in output")
	}

	// Verify other tasks unchanged
	tasks := ParseTasks(result)
	if tasks[0].Done {
		t.Error("task 1 should still be unchecked")
	}
	if !tasks[1].Done {
		t.Error("task 2 should be done")
	}
	if tasks[2].Done {
		t.Error("task 3 should still be unchecked")
	}
}

func TestSetTaskDoneOutOfRange(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Only task
`)
	_, err := SetTaskDone(data, 5)
	if err == nil {
		t.Fatal("expected error for out of range")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetTaskBlocked(t *testing.T) {
	data := []byte(`# Plan

## Tasks

- [ ] **1.** First task
- [ ] **2.** Second task
`)
	result, err := SetTaskBlocked(data, 1, "waiting on API")
	if err != nil {
		t.Fatalf("SetTaskBlocked: %v", err)
	}

	lines := strings.Split(string(result), "\n")
	for _, line := range lines {
		if strings.Contains(line, "First task") {
			if !strings.Contains(line, "[blocked: waiting on API]") {
				t.Errorf("task 1 not blocked: %q", line)
			}
			break
		}
	}
}

func TestSetTaskBlockedOutOfRange(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Only task
`)
	_, err := SetTaskBlocked(data, 3, "reason")
	if err == nil {
		t.Fatal("expected error for out of range")
	}
}

func TestSetTaskBlockedReplacesExisting(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Task [blocked: old reason]
`)
	result, err := SetTaskBlocked(data, 1, "new reason")
	if err != nil {
		t.Fatalf("SetTaskBlocked: %v", err)
	}

	s := string(result)
	if strings.Contains(s, "old reason") {
		t.Error("old blocked annotation should be removed")
	}
	if !strings.Contains(s, "[blocked: new reason]") {
		t.Error("new blocked annotation should be present")
	}
}

func TestPreservesFormatting(t *testing.T) {
	data := []byte(`# My Plan

This is the goal with **bold** and _italic_.

## Acceptance Criteria

- [ ] criterion 1
- [ ] criterion 2

## Tasks

- [x] **1.** Done task
  Long description here.
  - [x] 1.1. verified
- [ ] **2.** Pending task
  Another description.

## Notes

Custom notes here with formatting.
`)
	result, err := SetTaskDone(data, 2)
	if err != nil {
		t.Fatalf("SetTaskDone: %v", err)
	}

	// Check that non-task content is preserved
	s := string(result)
	if !strings.Contains(s, "**bold** and _italic_") {
		t.Error("formatting should be preserved")
	}
	if !strings.Contains(s, "Custom notes here") {
		t.Error("notes should be preserved")
	}
	if !strings.Contains(s, "Long description here.") {
		t.Error("task descriptions should be preserved")
	}
}

func TestSetSubTaskDone(t *testing.T) {
	data := []byte(`# Plan

## Tasks

- [ ] **1.** First task
  - [ ] 1.1. verify API works
  - [ ] 1.2. verify DB works
- [ ] **2.** Second task
  - [ ] 2.1. check logs
`)
	result, err := SetSubTaskDone(data, "1.1")
	if err != nil {
		t.Fatalf("SetSubTaskDone: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, "- [x] 1.1. verify API works") {
		t.Error("sub-task 1.1 should be marked done")
	}
	// Other sub-tasks unchanged
	if !strings.Contains(s, "- [ ] 1.2. verify DB works") {
		t.Error("sub-task 1.2 should still be unchecked")
	}
	if !strings.Contains(s, "- [ ] 2.1. check logs") {
		t.Error("sub-task 2.1 should still be unchecked")
	}
	// Parent task unchanged
	if !strings.Contains(s, "- [ ] **1.** First task") {
		t.Error("parent task should still be unchecked")
	}
}

func TestSetSubTaskDoneSecondTask(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** First task
  - [ ] 1.1. step one
- [ ] **2.** Second task
  - [ ] 2.1. step one
  - [ ] 2.2. step two
`)
	result, err := SetSubTaskDone(data, "2.2")
	if err != nil {
		t.Fatalf("SetSubTaskDone: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, "- [x] 2.2. step two") {
		t.Error("sub-task 2.2 should be marked done")
	}
	if !strings.Contains(s, "- [ ] 2.1. step one") {
		t.Error("sub-task 2.1 should still be unchecked")
	}
}

func TestSetSubTaskDoneNotFound(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Task
  - [ ] 1.1. step
`)
	_, err := SetSubTaskDone(data, "1.5")
	if err == nil {
		t.Fatal("expected error for nonexistent sub-task")
	}
}

func TestSetSubTaskDoneAlreadyDone(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Task
  - [x] 1.1. already done
  - [ ] 1.2. not done
`)
	// Already done sub-task — should still work (no-op on the checkbox)
	result, err := SetSubTaskDone(data, "1.2")
	if err != nil {
		t.Fatalf("SetSubTaskDone: %v", err)
	}
	if !strings.Contains(string(result), "- [x] 1.2. not done") {
		t.Error("sub-task 1.2 should be marked done")
	}
}

func TestParsePlanStatus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	data := []byte(`# My Plan

Goal.

## Tasks

- [x] **1.** Done task
- [x] **2.** Also done
- [ ] **3.** Not done
`)
	os.WriteFile(path, data, 0o644)

	status, err := ParsePlanStatus(path)
	if err != nil {
		t.Fatalf("ParsePlanStatus: %v", err)
	}
	if status.Name != "My Plan" {
		t.Errorf("Name = %q, want 'My Plan'", status.Name)
	}
	if status.Total != 3 {
		t.Errorf("Total = %d, want 3", status.Total)
	}
	if status.DoneCount != 2 {
		t.Errorf("DoneCount = %d, want 2", status.DoneCount)
	}
	if status.FilePath != path {
		t.Errorf("FilePath = %q, want %q", status.FilePath, path)
	}
}

func TestParsePlanStatusMissingFile(t *testing.T) {
	_, err := ParsePlanStatus("/nonexistent/path.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParsePlanStatusNoTasks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")
	os.WriteFile(path, []byte("# Empty Plan\n\n## Tasks\n"), 0o644)

	status, err := ParsePlanStatus(path)
	if err != nil {
		t.Fatalf("ParsePlanStatus: %v", err)
	}
	if status.Total != 0 {
		t.Errorf("Total = %d, want 0", status.Total)
	}
	if status.DoneCount != 0 {
		t.Errorf("DoneCount = %d, want 0", status.DoneCount)
	}
}

func TestSetSubTaskDonePreservesParentAndSiblings(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Parent task
  Some description.
  - [ ] 1.1. first step
  - [ ] 1.2. second step
  - [ ] 1.3. third step
`)
	// Mark middle sub-task
	result, err := SetSubTaskDone(data, "1.2")
	if err != nil {
		t.Fatalf("SetSubTaskDone: %v", err)
	}
	s := string(result)
	if !strings.Contains(s, "- [x] 1.2. second step") {
		t.Error("1.2 should be done")
	}
	if !strings.Contains(s, "- [ ] 1.1. first step") {
		t.Error("1.1 should still be unchecked")
	}
	if !strings.Contains(s, "- [ ] 1.3. third step") {
		t.Error("1.3 should still be unchecked")
	}
	if !strings.Contains(s, "- [ ] **1.** Parent task") {
		t.Error("parent should still be unchecked")
	}
	if !strings.Contains(s, "Some description.") {
		t.Error("description should be preserved")
	}
}

func TestParseTasksWithBlocked(t *testing.T) {
	data := []byte(`## Tasks

- [ ] **1.** Task one [blocked: waiting on API]
- [x] **2.** Task two
`)
	tasks := ParseTasks(data)
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	if !tasks[0].Blocked {
		t.Error("task 1 should be blocked")
	}
	if tasks[1].Blocked {
		t.Error("task 2 should not be blocked")
	}
}
