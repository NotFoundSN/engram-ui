package installer

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

// CopySkill copies the embedded skill tree from srcRoot within the given embed.FS
// to the destination directory dst. It creates directories as needed (0755) and
// writes files with mode 0644. The operation is idempotent — running it again
// overwrites existing files without error.
func CopySkill(dst string, fsys fs.FS, srcRoot string) error {
	return fs.WalkDir(fsys, srcRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path from srcRoot (using slash-separated path from embed.FS).
		relSlash := p[len(srcRoot):]
		if len(relSlash) > 0 && relSlash[0] == '/' {
			relSlash = relSlash[1:]
		}
		rel := filepath.FromSlash(relSlash)
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := fsys.(interface {
			ReadFile(string) ([]byte, error)
		}).ReadFile(p)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		return os.WriteFile(target, data, 0o644)
	})
}

// InstallClaudeCodeSkill copies the embedded Claude variant of the named skill
// (skills/{name}/claude/) to the Claude Code skills directory.
func InstallClaudeCodeSkill(name string) (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	srcRoot := path.Join(skillsEmbedRoot, name, "claude")
	return installSkill(ClaudeSkillDir(home, name), srcRoot)
}

// InstallOpenCodeSkill copies the embedded OpenCode variant of the named skill
// (skills/{name}/opencode/) to the OpenCode skills directory.
func InstallOpenCodeSkill(name string) (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")
	srcRoot := path.Join(skillsEmbedRoot, name, "opencode")
	return installSkill(OpenCodeSkillDir(home, xdg, name), srcRoot)
}

// UninstallClaudeCodeSkill removes the named skill from the Claude Code skills
// directory. Returns ActionRemoved when the directory existed and was removed,
// ActionNotRegistered when it did not exist.
func UninstallClaudeCodeSkill(name string) (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	return uninstallSkill(ClaudeSkillDir(home, name))
}

// UninstallOpenCodeSkill removes the named skill from the OpenCode skills directory.
func UninstallOpenCodeSkill(name string) (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")
	return uninstallSkill(OpenCodeSkillDir(home, xdg, name))
}

func uninstallSkill(destRoot string) (Result, error) {
	if _, err := os.Stat(destRoot); os.IsNotExist(err) {
		return Result{Destination: destRoot, Action: ActionNotRegistered}, nil
	}
	if err := os.RemoveAll(destRoot); err != nil {
		return Result{Destination: destRoot}, err
	}
	return Result{Destination: destRoot, Action: ActionRemoved}, nil
}

// installSkill copies the embedded subtree at srcRoot to destRoot. It returns
// ActionInstalled when destRoot is empty, ActionOverwritten when SKILL.md
// already exists there.
func installSkill(destRoot, srcRoot string) (Result, error) {
	// Verify srcRoot exists in the embed before doing any FS writes — gives a
	// clear error when caller passes a name that has no claude/ or opencode/
	// variant yet (e.g. skill still pending migration).
	if _, err := fs.Stat(skillFS, srcRoot+"/SKILL.md"); err != nil {
		return Result{}, fmt.Errorf("skill source %q not found in embed: %w", srcRoot, err)
	}

	_, statErr := os.Stat(filepath.Join(destRoot, "SKILL.md"))
	action := ActionInstalled
	if statErr == nil {
		action = ActionOverwritten
	}

	if err := CopySkill(destRoot, skillFS, srcRoot); err != nil {
		return Result{}, err
	}

	return Result{
		Destination: destRoot,
		Action:      action,
	}, nil
}
