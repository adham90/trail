package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	data := []byte("# Test\n\nContent here.\n")
	if err := AtomicWriteFile(path, data); err != nil {
		t.Fatalf("AtomicWriteFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("file content = %q, want %q", got, data)
	}

	// No temp file left behind
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after write")
	}
}

func TestAtomicWriteFileCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "test.md")

	data := []byte("content")
	if err := AtomicWriteFile(path, data); err != nil {
		t.Fatalf("AtomicWriteFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != "content" {
		t.Errorf("file content = %q, want %q", got, "content")
	}
}

func TestCreateBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plan.md")

	original := []byte("original content")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatalf("writing original: %v", err)
	}

	if err := CreateBackup(path); err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	backupPath := filepath.Join(dir, ".backup")
	got, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(got) != "original content" {
		t.Errorf("backup content = %q, want %q", got, "original content")
	}
}

func TestCreateBackupMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.md")

	err := CreateBackup(path)
	if err == nil {
		t.Error("expected error for missing file")
	}
}
