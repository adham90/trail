package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSerialize(t *testing.T) {
	p := &Plan{
		Name:         "test",
		Goal:         "Build trail",
		Branch:       "main",
		Status:       "active",
		SessionCount: 1,
		Created:      "2026-03-14",
		Updated:      "2026-03-14",
		CurrentTask:  0,
		Tasks: []Task{
			{Text: "Init module", Status: "done"},
			{Text: "Add deps", Status: "active"},
		},
		Context: Context{},
		Decisions: []Decision{
			{Time: time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC), Text: "Go over Ruby"},
		},
	}
	notes := "Some notes here."

	data, err := Serialize(p, notes)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := string(data)

	// Should start with ---
	if !strings.HasPrefix(output, "---\n") {
		t.Error("output does not start with ---")
	}

	// Should contain the closing --- separator
	if !strings.Contains(output, "\n---\n") {
		t.Error("output does not contain closing ---")
	}

	// Should contain goal in YAML frontmatter
	if !strings.Contains(output, "goal: Build trail") {
		t.Error("output missing goal field in YAML")
	}

	// Should contain rendered goal section
	if !strings.Contains(output, "## goal\n\nBuild trail") {
		t.Error("output missing rendered goal section")
	}

	// Should contain tasks
	if !strings.Contains(output, "text: Init module") {
		t.Error("output missing task text")
	}

	// Should contain notes after frontmatter
	if !strings.Contains(output, "## notes") {
		t.Error("output missing notes section")
	}

	// Round-trip: parse the serialized output
	p2, notes2, err := Parse(data)
	if err != nil {
		t.Fatalf("Round-trip Parse failed: %v", err)
	}
	if p2.Goal != p.Goal {
		t.Errorf("Round-trip Goal = %q, want %q", p2.Goal, p.Goal)
	}
	if len(p2.Tasks) != len(p.Tasks) {
		t.Errorf("Round-trip len(Tasks) = %d, want %d", len(p2.Tasks), len(p.Tasks))
	}
	if p2.Tasks[0].Status != "done" {
		t.Errorf("Round-trip Tasks[0].Status = %q, want %q", p2.Tasks[0].Status, "done")
	}
	if notes2 != "Some notes here." {
		t.Errorf("Round-trip notes = %q, want %q", notes2, "Some notes here.")
	}
}

func TestSerializeEmptyContext(t *testing.T) {
	p := &Plan{
		Name:         "test",
		Goal:    "Test",
		Branch:  "main",
		Status:  "active",
		Context: Context{},
		Tasks:   []Task{},
	}

	data, err := Serialize(p, "")
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Empty context fields should serialize as null (~), not ""
	output := string(data)
	if strings.Contains(output, `current_file: ""`) {
		t.Error("empty context field serialized as empty string instead of null")
	}
	// Verify it contains null representation
	if !strings.Contains(output, "current_file: null") {
		t.Errorf("expected null for empty context field, got:\n%s", output)
	}
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-plan.md")

	p := &Plan{
		Name:         "test",
		Goal:         "Test atomic write",
		Branch:       "test",
		Status:       "active",
		SessionCount: 1,
		Created:      "2026-03-14",
		Updated:      "2026-03-14",
		CurrentTask:  0,
		Tasks:        []Task{{Text: "First task", Status: "todo"}},
		Context:      Context{},
		Decisions:    []Decision{},
	}

	err := WriteFile(path, p, "")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// File should exist
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Should round-trip
	p2, _, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse after WriteFile failed: %v", err)
	}
	if p2.Goal != "Test atomic write" {
		t.Errorf("Goal = %q, want %q", p2.Goal, "Test atomic write")
	}

	// No temp files should remain
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestWriteFileCreatesBackup(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, ".plans")
	os.MkdirAll(plansDir, 0o755)
	path := filepath.Join(plansDir, "test.md")

	p := &Plan{
		Name:         "test",
		Goal:    "v1",
		Branch:  "test",
		Status:  "active",
		Tasks:   []Task{{Text: "First", Status: "todo"}},
		Context: Context{},
	}

	// First write — no backup since file doesn't exist yet
	err := WriteFile(path, p, "")
	if err != nil {
		t.Fatalf("First WriteFile failed: %v", err)
	}

	backupPath := filepath.Join(plansDir, ".backup")
	if _, err := os.Stat(backupPath); err == nil {
		t.Error("backup should not exist after first write (no previous file)")
	}

	// Second write — should create backup
	p.Goal = "v2"
	err = WriteFile(path, p, "")
	if err != nil {
		t.Fatalf("Second WriteFile failed: %v", err)
	}

	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}

	backupPlan, _, err := Parse(backupData)
	if err != nil {
		t.Fatalf("Parse backup failed: %v", err)
	}
	if backupPlan.Goal != "v1" {
		t.Errorf("backup Goal = %q, want %q", backupPlan.Goal, "v1")
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	p := &Plan{
		Name:         "test",
		Goal:         "Read test",
		Branch:       "main",
		Status:       "active",
		SessionCount: 1,
		Created:      "2026-03-14",
		Updated:      "2026-03-14",
		Tasks:        []Task{{Text: "A task", Status: "todo"}},
		Context:      Context{},
		Decisions:    []Decision{},
	}

	err := WriteFile(path, p, "Hello.")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	p2, notes, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if p2.Goal != "Read test" {
		t.Errorf("Goal = %q, want %q", p2.Goal, "Read test")
	}
	if !strings.Contains(notes, "Hello.") {
		t.Errorf("notes = %q, want to contain 'Hello.'", notes)
	}
}

func TestSerializeRendersMarkdown(t *testing.T) {
	p := &Plan{
		Name:         "test",
		Goal:   "Test rendering",
		Branch: "main",
		Status: "active",
		Tasks: []Task{
			{Text: "First", Status: "done"},
			{Text: "Second", Status: "active"},
			{Text: "Third", Status: "todo"},
			{Text: "Fourth", Status: "blocked", Reason: "waiting"},
		},
		Context: Context{
			CurrentFile: "main.go",
		},
		Decisions: []Decision{
			{Time: time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC), Text: "Use Go"},
		},
	}

	data, err := Serialize(p, "My notes here.")
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := string(data)

	// Should have rendered task checkboxes
	if !strings.Contains(output, "- [x] 00 · First") {
		t.Error("missing done task checkbox")
	}
	if !strings.Contains(output, "- [▶] 01 · Second") {
		t.Error("missing active task checkbox")
	}
	if !strings.Contains(output, "- [ ] 02 · Third") {
		t.Error("missing todo task checkbox")
	}
	if !strings.Contains(output, "- [!] 03 · Fourth — blocked: waiting") {
		t.Error("missing blocked task with reason")
	}

	// Should have context table
	if !strings.Contains(output, "| current_file | main.go |") {
		t.Error("missing context table entry")
	}
	if !strings.Contains(output, "| last_error | ~ |") {
		t.Error("missing null context field")
	}

	// Should have decisions
	if !strings.Contains(output, "2026-03-14 · Use Go") {
		t.Error("missing decision")
	}

	// Should have notes
	if !strings.Contains(output, "My notes here.") {
		t.Error("missing notes content")
	}

	// Should have generated comment
	if !strings.Contains(output, "<!-- generated below") {
		t.Error("missing generated comment")
	}
}

func TestSerializeWithConstraintsAndFiles(t *testing.T) {
	p := &Plan{
		Name:         "test",
		Goal:   "Rich plan",
		Branch: "main",
		Status: "active",
		Constraints: []string{
			"No external APIs",
			"Atomic writes only",
		},
		PlanFiles: []FileRef{
			{Path: "model.go", Role: "core"},
			{Path: "writer.go", Role: "serialization"},
		},
		Tasks:     []Task{{Text: "First", Status: "active"}},
		Context:   Context{},
		Decisions: []Decision{},
	}

	data, err := Serialize(p, "")
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := string(data)

	// Constraints section
	if !strings.Contains(output, "## constraints") {
		t.Error("missing constraints section")
	}
	if !strings.Contains(output, "- No external APIs") {
		t.Error("missing constraint item")
	}

	// Files table
	if !strings.Contains(output, "## files") {
		t.Error("missing files section")
	}
	if !strings.Contains(output, "| model.go | core |") {
		t.Error("missing file table row")
	}

	// Round-trip
	p2, _, err := Parse(data)
	if err != nil {
		t.Fatalf("Round-trip Parse failed: %v", err)
	}
	if len(p2.Constraints) != 2 {
		t.Errorf("Round-trip Constraints count = %d, want 2", len(p2.Constraints))
	}
	if len(p2.PlanFiles) != 2 {
		t.Errorf("Round-trip PlanFiles count = %d, want 2", len(p2.PlanFiles))
	}
}

func TestSerializeActiveTaskExpansion(t *testing.T) {
	p := &Plan{
		Name:         "test",
		Goal:   "Test expansion",
		Branch: "main",
		Status: "active",
		Tasks: []Task{
			{Text: "Done task", Status: "done", Spec: "old spec"},
			{
				Text:   "Active task",
				Status: "active",
				Spec:   "Implement the feature",
				Verify: []string{"tests pass", "no regressions"},
				Files:  []string{"model.go", "writer.go"},
			},
			{Text: "Todo task", Status: "todo", Spec: "future spec"},
		},
		Context:   Context{},
		Decisions: []Decision{},
	}

	data, err := Serialize(p, "")
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := string(data)

	// Active task should be expanded
	if !strings.Contains(output, "**spec:** Implement the feature") {
		t.Error("missing active task spec")
	}
	if !strings.Contains(output, "- [ ] tests pass") {
		t.Error("missing active task verify checkbox")
	}
	if !strings.Contains(output, "model.go, writer.go") {
		t.Error("missing active task files")
	}

	// Done task spec should NOT appear in rendered markdown body
	// (it's in the YAML frontmatter but not rendered)
	lines := strings.Split(output, "---\n")
	if len(lines) < 3 {
		t.Fatal("unexpected output format")
	}
	renderedBody := lines[2] // everything after the closing ---
	if strings.Contains(renderedBody, "old spec") {
		t.Error("done task spec should not be rendered in markdown body")
	}
	if strings.Contains(renderedBody, "future spec") {
		t.Error("todo task spec should not be rendered in markdown body")
	}
}

func TestSerializeVerifyPassed(t *testing.T) {
	p := &Plan{
		Name:         "test",
		Goal:   "Test verify",
		Branch: "main",
		Status: "active",
		Tasks: []Task{
			{
				Text:   "Task with mixed verify",
				Status: "active",
				Verify: []string{"✓ unit tests pass", "integration tests"},
			},
		},
		Context:   Context{},
		Decisions: []Decision{},
	}

	data, err := Serialize(p, "")
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "- [x] ✓ unit tests pass") {
		t.Error("passed verify step should render with [x] checkbox")
	}
	if !strings.Contains(output, "- [ ] integration tests") {
		t.Error("pending verify step should render with [ ] checkbox")
	}
}

func TestRoundTripRichPlan(t *testing.T) {
	p := &Plan{
		Name:         "test",
		Goal:         "Round trip test",
		Branch:       "main",
		Status:       "active",
		SessionCount: 1,
		Created:      "2026-03-14",
		Updated:      "2026-03-14",
		Constraints:  []string{"constraint A", "constraint B"},
		PlanFiles:    []FileRef{{Path: "a.go", Role: "main entry"}},
		Tasks: []Task{
			{
				Text:   "Rich task",
				Status: "active",
				Spec:   "do the thing",
				Verify: []string{"step 1", "step 2"},
				Files:  []string{"a.go", "b.go"},
			},
		},
		Context:   Context{},
		Decisions: []Decision{},
	}

	data, err := Serialize(p, "some notes")
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	p2, notes2, err := Parse(data)
	if err != nil {
		t.Fatalf("Round-trip Parse failed: %v", err)
	}

	if p2.Goal != "Round trip test" {
		t.Errorf("Goal = %q", p2.Goal)
	}
	if len(p2.Constraints) != 2 {
		t.Errorf("Constraints count = %d", len(p2.Constraints))
	}
	if len(p2.PlanFiles) != 1 {
		t.Errorf("PlanFiles count = %d", len(p2.PlanFiles))
	}
	if p2.Tasks[0].Spec != "do the thing" {
		t.Errorf("Task Spec = %q", p2.Tasks[0].Spec)
	}
	if len(p2.Tasks[0].Verify) != 2 {
		t.Errorf("Task Verify count = %d", len(p2.Tasks[0].Verify))
	}
	if len(p2.Tasks[0].Files) != 2 {
		t.Errorf("Task Files count = %d", len(p2.Tasks[0].Files))
	}
	if notes2 != "some notes" {
		t.Errorf("notes = %q, want 'some notes'", notes2)
	}
}

func TestSerializeWithDiagram(t *testing.T) {
	p := &Plan{
		Name:    "test",
		Goal:    "Test diagram support",
		Diagram: "graph TD\n    A[Start] --> B[End]",
		Branch:  "main",
		Status:  "active",
		Tasks:   []Task{{Text: "First", Status: "todo"}},
		Context: Context{},
	}

	data, err := Serialize(p, "")
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := string(data)

	// Should contain diagram section with mermaid code block
	if !strings.Contains(output, "## diagram") {
		t.Error("missing diagram section")
	}
	if !strings.Contains(output, "```mermaid\ngraph TD\n    A[Start] --> B[End]\n```") {
		t.Error("missing mermaid code block")
	}

	// Diagram section should appear before tasks
	diagramIdx := strings.Index(output, "## diagram")
	tasksIdx := strings.Index(output, "## tasks")
	if diagramIdx > tasksIdx {
		t.Error("diagram section should appear before tasks")
	}

	// Goal section should appear before diagram
	goalIdx := strings.Index(output, "## goal")
	if goalIdx > diagramIdx {
		t.Error("goal section should appear before diagram")
	}

	// Round-trip
	p2, _, err := Parse(data)
	if err != nil {
		t.Fatalf("Round-trip Parse failed: %v", err)
	}
	if p2.Diagram != p.Diagram {
		t.Errorf("Round-trip Diagram = %q, want %q", p2.Diagram, p.Diagram)
	}
}

func TestSerializeWithoutDiagram(t *testing.T) {
	p := &Plan{
		Name:    "test",
		Goal:    "No diagram plan",
		Status:  "active",
		Tasks:   []Task{{Text: "First", Status: "todo"}},
		Context: Context{},
	}

	data, err := Serialize(p, "")
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	output := string(data)

	// Should NOT contain diagram section when diagram is empty
	if strings.Contains(output, "## diagram") {
		t.Error("diagram section should not appear when diagram is empty")
	}

	// Should still have goal section
	if !strings.Contains(output, "## goal") {
		t.Error("missing goal section")
	}
}

func TestReadFileNotFound(t *testing.T) {
	_, _, err := ReadFile("/nonexistent/path/plan.md")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
