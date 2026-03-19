package plan

import (
	"os"
	"path/filepath"
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

func TestParseTaskCounts(t *testing.T) {
	data := []byte(`# Test Plan

Goal.

## Tasks

- [x] First task done
  - [x] sub-step (ignored)
- [ ] Second task pending
  Description text.
  - [ ] verify step (ignored)
- [ ] Third task

## Notes

Some notes.
`)
	done, total := ParseTaskCounts(data)
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if done != 1 {
		t.Errorf("done = %d, want 1", done)
	}
}

func TestParseTaskCountsIgnoresSubItems(t *testing.T) {
	data := []byte(`## Tasks

- [ ] Top task
  - [ ] sub
  - [x] sub done
- [x] Done task
`)
	done, total := ParseTaskCounts(data)
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if done != 1 {
		t.Errorf("done = %d, want 1", done)
	}
}

func TestParseTaskCountsCaseInsensitiveHeading(t *testing.T) {
	data := []byte(`## tasks

- [ ] A task
`)
	done, total := ParseTaskCounts(data)
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if done != 0 {
		t.Errorf("done = %d, want 0", done)
	}
}

func TestParseTaskCountsStopsAtNextSection(t *testing.T) {
	data := []byte(`## Tasks

- [ ] Real task

## Notes

- [ ] Not a task (in notes section)
`)
	_, total := ParseTaskCounts(data)
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
}

func TestParseTaskCountsAllDone(t *testing.T) {
	data := []byte(`## Tasks

- [x] Task one
- [x] Task two
- [x] Task three
`)
	done, total := ParseTaskCounts(data)
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if done != 3 {
		t.Errorf("done = %d, want 3", done)
	}
}

func TestParseTaskCountsNoneDone(t *testing.T) {
	data := []byte(`## Tasks

- [ ] Task one
- [ ] Task two
`)
	done, total := ParseTaskCounts(data)
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if done != 0 {
		t.Errorf("done = %d, want 0", done)
	}
}

func TestParseTaskCountsEmpty(t *testing.T) {
	data := []byte(`## Tasks
`)
	done, total := ParseTaskCounts(data)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if done != 0 {
		t.Errorf("done = %d, want 0", done)
	}
}

func TestParseTaskCountsNoTasksSection(t *testing.T) {
	data := []byte(`# Plan

Some text without a tasks section.
`)
	done, total := ParseTaskCounts(data)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if done != 0 {
		t.Errorf("done = %d, want 0", done)
	}
}

func TestParseTaskCountsUppercaseX(t *testing.T) {
	data := []byte(`## Tasks

- [X] Task with uppercase X
- [ ] Undone task
`)
	done, total := ParseTaskCounts(data)
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if done != 1 {
		t.Errorf("done = %d, want 1", done)
	}
}

func TestParsePlanStatus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	data := []byte(`# My Plan

Goal.

## Tasks

- [x] Done task
- [x] Also done
- [ ] Not done
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
