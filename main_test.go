package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateFileName(t *testing.T) {
	valid := []string{
		"4171_report.txt",
		"4171_Sample2A_S4-A4_1_54910.d.zip",
	}
	for _, name := range valid {
		if err := validateFileName(name); err != nil {
			t.Fatalf("expected %q to be valid: %v", name, err)
		}
	}

	invalid := []string{
		".hidden_file",
		"sample 01.raw",
		"experiment@02.mzML",
	}
	for _, name := range invalid {
		if err := validateFileName(name); err == nil {
			t.Fatalf("expected %q to be invalid", name)
		}
	}
}

func TestRunWritesPrideChecksumFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "file2.txt"), "two")
	writeFile(t, filepath.Join(dir, "file1.txt"), "one")

	if err := run([]string{dir}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	gotBytes, err := os.ReadFile(filepath.Join(dir, checksumFileName))
	if err != nil {
		t.Fatalf("could not read checksum file: %v", err)
	}

	got := string(gotBytes)
	wantLines := []string{
		"# SHA-1 Checksum ",
		"file1.txt\tfe05bcdcdc4928012781a5f1a2a77cbb5398e106",
		"file2.txt\tad782ecdac770fc6eb9a62e44f90873fb97fb26b",
	}
	for _, want := range wantLines {
		if !strings.Contains(got, want) {
			t.Fatalf("checksum file missing %q:\n%s", want, got)
		}
	}
}

func TestCollectFilesRejectsSubfolders(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subfolder"), 0o755); err != nil {
		t.Fatalf("could not create subfolder: %v", err)
	}

	if _, err := collectFiles(dir); err == nil {
		t.Fatal("expected subfolders to be rejected")
	}
}

func TestRunFailureDoesNotOverwriteExistingChecksum(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, checksumFileName), "previous checksum\n")
	writeFile(t, filepath.Join(dir, "bad name.txt"), "bad")

	if err := run([]string{dir}); err == nil {
		t.Fatal("expected invalid filename to fail")
	}

	gotBytes, err := os.ReadFile(filepath.Join(dir, checksumFileName))
	if err != nil {
		t.Fatalf("could not read checksum file: %v", err)
	}
	if got := string(gotBytes); got != "previous checksum\n" {
		t.Fatalf("existing checksum was overwritten: %q", got)
	}
}

func TestWriteChecksumRejectsExistingChecksumDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, checksumFileName), 0o755); err != nil {
		t.Fatalf("could not create checksum directory: %v", err)
	}

	err := writeChecksumFile(dir, []fileResult{{name: "file1.txt", sum: "abc123"}})
	if err == nil {
		t.Fatal("expected existing checksum directory to fail")
	}
	if _, statErr := os.Stat(filepath.Join(dir, tempFileName)); !os.IsNotExist(statErr) {
		t.Fatalf("temporary checksum file was not cleaned up, stat err: %v", statErr)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("could not write %s: %v", path, err)
	}
}
