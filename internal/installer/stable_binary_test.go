package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureStableBinary_SourceAlreadyStable(t *testing.T) {
	// Create a temporary directory to act as home
	tmpDir := t.TempDir()
	localAppData := tmpDir
	if runtime.GOOS == "windows" {
		localAppData = filepath.Join(tmpDir, "AppData", "Local")
	}

	// Create a "stable" binary
	stablePath := StableBinaryPath(tmpDir, localAppData, runtime.GOOS)
	if err := os.MkdirAll(filepath.Dir(stablePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stablePath, []byte("binary content"), 0755); err != nil {
		t.Fatal(err)
	}

	// Test that EnsureStableBinary returns skipped when source is already at stable path
	result, err := EnsureStableBinary(stablePath, tmpDir, localAppData)
	if err != nil {
		t.Errorf("EnsureStableBinary() error = %v, want nil", err)
	}
	if result.Action != ActionSkipped {
		t.Errorf("EnsureStableBinary() action = %v, want %v", result.Action, ActionSkipped)
	}
}

func TestEnsureStableBinary_ByteIdentical(t *testing.T) {
	// Skip on Windows due to file locking issues in tests
	if runtime.GOOS == "windows" {
		t.Skip("Skipping byte-identical test on Windows")
	}

	tmpDir := t.TempDir()
	localAppData := tmpDir
	if runtime.GOOS == "windows" {
		localAppData = filepath.Join(tmpDir, "AppData", "Local")
	}

	// Create source binary
	sourcePath := filepath.Join(tmpDir, "source-engram-ui")
	content := []byte("binary content v1.0.0")
	if err := os.WriteFile(sourcePath, content, 0755); err != nil {
		t.Fatal(err)
	}

	// Create stable directory and stable binary with identical content
	stablePath := StableBinaryPath(tmpDir, localAppData, runtime.GOOS)
	if err := os.MkdirAll(filepath.Dir(stablePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stablePath, content, 0755); err != nil {
		t.Fatal(err)
	}

	// Test that EnsureStableBinary skips when content is identical
	result, err := EnsureStableBinary(sourcePath, tmpDir, localAppData)
	if err != nil {
		t.Errorf("EnsureStableBinary() error = %v, want nil", err)
	}
	if result.Action != ActionSkipped {
		t.Errorf("EnsureStableBinary() action = %v, want %v", result.Action, ActionSkipped)
	}
}

func TestEnsureStableBinary_NewInstall(t *testing.T) {
	// Skip on Windows due to file locking issues
	if runtime.GOOS == "windows" {
		t.Skip("Skipping new install test on Windows")
	}

	tmpDir := t.TempDir()
	localAppData := tmpDir
	if runtime.GOOS == "windows" {
		localAppData = filepath.Join(tmpDir, "AppData", "Local")
	}

	// Create source binary
	sourcePath := filepath.Join(tmpDir, "source-engram-ui")
	content := []byte("binary content")
	if err := os.WriteFile(sourcePath, content, 0755); err != nil {
		t.Fatal(err)
	}

	// Test new installation
	result, err := EnsureStableBinary(sourcePath, tmpDir, localAppData)
	if err != nil {
		t.Errorf("EnsureStableBinary() error = %v, want nil", err)
	}
	if result.Action != ActionInstalled {
		t.Errorf("EnsureStableBinary() action = %v, want %v", result.Action, ActionInstalled)
	}

	// Verify stable binary was created
	stablePath := StableBinaryPath(tmpDir, localAppData, runtime.GOOS)
	if _, err := os.Stat(stablePath); os.IsNotExist(err) {
		t.Errorf("Stable binary was not created at %s", stablePath)
	}

	// Verify content matches
	stableContent, err := os.ReadFile(stablePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(stableContent) != string(content) {
		t.Errorf("Stable binary content mismatch")
	}
}

func TestEnsureStableBinary_UpgradeAllowed(t *testing.T) {
	// Skip on Windows due to file locking issues
	if runtime.GOOS == "windows" {
		t.Skip("Skipping upgrade test on Windows")
	}

	tmpDir := t.TempDir()
	localAppData := tmpDir

	// Create source binary (simulating v2.0)
	sourcePath := filepath.Join(tmpDir, "source-engram-ui")
	sourceContent := []byte("binary content v2.0")
	if err := os.WriteFile(sourcePath, sourceContent, 0755); err != nil {
		t.Fatal(err)
	}

	// Create stable binary (simulating v1.0 already installed)
	stablePath := StableBinaryPath(tmpDir, localAppData, runtime.GOOS)
	if err := os.MkdirAll(filepath.Dir(stablePath), 0755); err != nil {
		t.Fatal(err)
	}
	stableContent := []byte("binary content v1.0")
	if err := os.WriteFile(stablePath, stableContent, 0755); err != nil {
		t.Fatal(err)
	}

	// Stub the version resolver to return older installed version
	originalResolver := resolveBinaryVersion
	resolveBinaryVersion = func(binaryPath string) (string, error) {
		return "1.0.0", nil
	}
	defer func() {
		resolveBinaryVersion = originalResolver
	}()

	// Execute upgrade
	result, err := EnsureStableBinary(sourcePath, tmpDir, localAppData)
	if err != nil {
		t.Errorf("EnsureStableBinary() error = %v, want nil", err)
	}
	if result.Action != ActionOverwritten {
		t.Errorf("EnsureStableBinary() action = %v, want %v", result.Action, ActionOverwritten)
	}

	// Verify stable binary was overwritten with source content
	finalContent, err := os.ReadFile(stablePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(finalContent) != string(sourceContent) {
		t.Errorf("Stable binary content mismatch after upgrade")
	}
}

func TestEnsureStableBinary_DowngradeBlocked(t *testing.T) {
	// Skip on Windows due to file locking issues
	if runtime.GOOS == "windows" {
		t.Skip("Skipping downgrade test on Windows")
	}

	tmpDir := t.TempDir()
	localAppData := tmpDir

	// Create source binary (simulating v1.0)
	sourcePath := filepath.Join(tmpDir, "source-engram-ui")
	sourceContent := []byte("binary content v1.0")
	if err := os.WriteFile(sourcePath, sourceContent, 0755); err != nil {
		t.Fatal(err)
	}

	// Create stable binary (simulating v2.0 already installed)
	stablePath := StableBinaryPath(tmpDir, localAppData, runtime.GOOS)
	if err := os.MkdirAll(filepath.Dir(stablePath), 0755); err != nil {
		t.Fatal(err)
	}
	stableContent := []byte("binary content v2.0")
	if err := os.WriteFile(stablePath, stableContent, 0755); err != nil {
		t.Fatal(err)
	}

	// Stub the version resolver to return newer installed version
	originalResolver := resolveBinaryVersion
	resolveBinaryVersion = func(binaryPath string) (string, error) {
		return "2.0.0", nil
	}
	defer func() {
		resolveBinaryVersion = originalResolver
	}()

	// Execute downgrade attempt
	result, err := EnsureStableBinary(sourcePath, tmpDir, localAppData)
	if err != ErrDowngradeBlocked {
		t.Errorf("EnsureStableBinary() error = %v, want ErrDowngradeBlocked", err)
	}
	if result.Action != ActionBlockedDowngrade {
		t.Errorf("EnsureStableBinary() action = %v, want %v", result.Action, ActionBlockedDowngrade)
	}

	// Verify stable binary was NOT overwritten
	finalContent, err := os.ReadFile(stablePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(finalContent) != string(stableContent) {
		t.Errorf("Stable binary was overwritten when downgrade should have been blocked")
	}
}

func TestFilesAreIdentical(t *testing.T) {
	tmpDir := t.TempDir()

	// Test identical files
	file1 := filepath.Join(tmpDir, "file1")
	file2 := filepath.Join(tmpDir, "file2")
	content := []byte("identical content")

	if err := os.WriteFile(file1, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, content, 0644); err != nil {
		t.Fatal(err)
	}

	identical, err := filesAreIdentical(file1, file2)
	if err != nil {
		t.Errorf("filesAreIdentical() error = %v", err)
	}
	if !identical {
		t.Error("filesAreIdentical() = false, want true for identical files")
	}

	// Test different files
	file3 := filepath.Join(tmpDir, "file3")
	if err := os.WriteFile(file3, []byte("different content"), 0644); err != nil {
		t.Fatal(err)
	}

	identical, err = filesAreIdentical(file1, file3)
	if err != nil {
		t.Errorf("filesAreIdentical() error = %v", err)
	}
	if identical {
		t.Error("filesAreIdentical() = true, want false for different files")
	}

	// Test non-existent file
	_, err = filesAreIdentical(file1, filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("filesAreIdentical() should return error for non-existent file")
	}
}

func TestCopyFile(t *testing.T) {
	// Skip on Windows
	if runtime.GOOS == "windows" {
		t.Skip("Skipping copy test on Windows")
	}

	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "source")
	dest := filepath.Join(tmpDir, "dest")
	content := []byte("file content")

	if err := os.WriteFile(source, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(source, dest); err != nil {
		t.Errorf("copyFile() error = %v", err)
	}

	// Verify content
	copied, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(copied) != string(content) {
		t.Errorf("copyFile() content mismatch")
	}

	// Verify permissions
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("copyFile() should set executable permissions")
	}
}

func TestResolveBinaryVersionReal(t *testing.T) {
	// Skip on Windows - would need different executable creation approach
	if runtime.GOOS == "windows" {
		t.Skip("Skipping resolveBinaryVersionReal test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a mock binary that prints version output
	mockBin := filepath.Join(tmpDir, "mock-engram-ui")
	// Simple shell script that outputs version when called with "version" arg
	script := `#!/bin/sh
if [ "$1" = "version" ]; then
	echo "engram-ui v1.2.3"
fi
`
	if err := os.WriteFile(mockBin, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	// Test successful version resolution
	version, err := resolveBinaryVersionReal(mockBin)
	if err != nil {
		t.Errorf("resolveBinaryVersionReal() error = %v, want nil", err)
	}
	if version != "v1.2.3" {
		t.Errorf("resolveBinaryVersionReal() = %v, want v1.2.3", version)
	}

	// Test with version without 'v' prefix
	scriptNoV := `#!/bin/sh
if [ "$1" = "version" ]; then
	echo "engram-ui 2.0.0"
fi
`
	mockBinNoV := filepath.Join(tmpDir, "mock-engram-ui-no-v")
	if err := os.WriteFile(mockBinNoV, []byte(scriptNoV), 0755); err != nil {
		t.Fatal(err)
	}

	version, err = resolveBinaryVersionReal(mockBinNoV)
	if err != nil {
		t.Errorf("resolveBinaryVersionReal() error = %v, want nil", err)
	}
	if version != "2.0.0" {
		t.Errorf("resolveBinaryVersionReal() = %v, want 2.0.0", version)
	}

	// Test with non-existent binary
	_, err = resolveBinaryVersionReal(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("resolveBinaryVersionReal() should return error for non-existent binary")
	}
}

func TestEnsureStableBinary_ParseFailureFallback(t *testing.T) {
	// Skip on Windows due to file locking issues
	if runtime.GOOS == "windows" {
		t.Skip("Skipping parse failure test on Windows")
	}

	tmpDir := t.TempDir()
	localAppData := tmpDir

	// Create source binary (simulating v1.0)
	sourcePath := filepath.Join(tmpDir, "source-engram-ui")
	sourceContent := []byte("binary content v1.0")
	if err := os.WriteFile(sourcePath, sourceContent, 0755); err != nil {
		t.Fatal(err)
	}

	// Create stable binary (different content - not byte-identical)
	stablePath := StableBinaryPath(tmpDir, localAppData, runtime.GOOS)
	if err := os.MkdirAll(filepath.Dir(stablePath), 0755); err != nil {
		t.Fatal(err)
	}
	stableContent := []byte("binary content existing")
	if err := os.WriteFile(stablePath, stableContent, 0755); err != nil {
		t.Fatal(err)
	}

	// Stub the version resolver to simulate parse failure (returns error)
	originalResolver := resolveBinaryVersion
	resolveBinaryVersion = func(binaryPath string) (string, error) {
		return "", fmt.Errorf("version parse failed: unparseable output")
	}
	defer func() {
		resolveBinaryVersion = originalResolver
	}()

	// Execute - should proceed with overwrite despite parse failure
	result, err := EnsureStableBinary(sourcePath, tmpDir, localAppData)
	if err != nil {
		t.Errorf("EnsureStableBinary() error = %v, want nil (should proceed on parse failure)", err)
	}
	// Should overwrite when version cannot be determined (downgrade guard bypassed)
	if result.Action != ActionOverwritten {
		t.Errorf("EnsureStableBinary() action = %v, want %v (overwrite should proceed when version parse fails)", result.Action, ActionOverwritten)
	}

	// Verify stable binary WAS overwritten (parse failure bypasses downgrade guard)
	finalContent, err := os.ReadFile(stablePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(finalContent) != string(sourceContent) {
		t.Errorf("Stable binary should be overwritten when version parse fails")
	}
}
