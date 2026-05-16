package installer

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCopySkill(t *testing.T) {
	dst := t.TempDir()
	srcRoot := "skills/brainstorm/claude"

	// First run: install.
	if err := CopySkill(dst, skillFS, srcRoot); err != nil {
		t.Fatalf("CopySkill first run: %v", err)
	}

	// Verify every embedded file appears at dst with matching content.
	err := fs.WalkDir(skillFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, relErr := filepath.Rel(filepath.FromSlash(srcRoot), filepath.FromSlash(path))
		if relErr != nil {
			return relErr
		}
		dstPath := filepath.Join(dst, rel)

		wantData, readErr := skillFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		gotData, readErr := os.ReadFile(dstPath)
		if readErr != nil {
			t.Errorf("CopySkill: file %q not found at destination: %v", dstPath, readErr)
			return nil
		}

		if string(gotData) != string(wantData) {
			t.Errorf("CopySkill: file %q content mismatch", dstPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir over skillFS: %v", err)
	}

	// Second run: idempotent — must not fail.
	if err := CopySkill(dst, skillFS, srcRoot); err != nil {
		t.Fatalf("CopySkill second run (idempotent): %v", err)
	}

	skillMD := filepath.Join(dst, "SKILL.md")
	if _, err := os.Stat(skillMD); err != nil {
		t.Errorf("CopySkill idempotent: SKILL.md missing: %v", err)
	}
}
