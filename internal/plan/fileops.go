package plan

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWriteFile writes data to path using a temp file + rename.
func AtomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// CreateBackup copies the file at path to plans/.backup in the same directory.
func CreateBackup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	backupPath := filepath.Join(dir, ".backup")
	return os.WriteFile(backupPath, data, 0o644)
}

// readFileBytes reads a file and returns its contents.
func readFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}
