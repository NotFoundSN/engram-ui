package installer

import (
	"io/fs"
	"os"
	"path/filepath"
)

// CopySkill copies the embedded skill tree from srcRoot within the given embed.FS
// to the destination directory dst. It creates directories as needed (0755) and
// writes files with mode 0644. The operation is idempotent — running it again
// overwrites existing files without error.
func CopySkill(dst string, fsys fs.FS, srcRoot string) error {
	return fs.WalkDir(fsys, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path from srcRoot (using slash-separated path from embed.FS).
		// filepath.Rel needs OS-native separators, but embed.FS always uses '/'.
		relSlash := path[len(srcRoot):]
		if len(relSlash) > 0 && relSlash[0] == '/' {
			relSlash = relSlash[1:]
		}
		// Convert to OS-native path separator for destination.
		rel := filepath.FromSlash(relSlash)
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		// Read file content from embed.FS.
		data, err := fsys.(interface {
			ReadFile(string) ([]byte, error)
		}).ReadFile(path)
		if err != nil {
			return err
		}

		// Ensure parent directory exists.
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		return os.WriteFile(target, data, 0o644)
	})
}

// InstallClaudeCodeSkill copies the embedded skill to the Claude Code skills directory.
func InstallClaudeCodeSkill() (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	return installSkill(ClaudeSkillDir(home))
}

// InstallOpenCodeSkill copies the embedded skill to the OpenCode skills directory.
func InstallOpenCodeSkill() (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")
	return installSkill(OpenCodeSkillDir(home, xdg))
}

// installSkill is the shared implementation — copies the embedded skill to destRoot.
func installSkill(destRoot string) (Result, error) {
	// Determine action: installed or overwritten.
	_, statErr := os.Stat(filepath.Join(destRoot, "SKILL.md"))
	action := ActionInstalled
	if statErr == nil {
		action = ActionOverwritten
	}

	if err := CopySkill(destRoot, skillFS, skillEmbedRoot); err != nil {
		return Result{}, err
	}

	return Result{
		Destination: destRoot,
		Action:      action,
	}, nil
}
