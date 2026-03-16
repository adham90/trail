package plan

import (
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
