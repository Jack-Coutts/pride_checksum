package main

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeRawFolder(t *testing.T, dir string) string {
	t.Helper()
	rawDir := filepath.Join(dir, "sample_1.raw")
	if err := os.MkdirAll(filepath.Join(rawDir, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(rawDir, "_HEADER.TXT"), "header")
	writeFile(t, filepath.Join(rawDir, "sub", "_FUNC001.DAT"), strings.Repeat("data", 1000))
	return rawDir
}

func TestRunZipsWatersRawFolderThenChecksumsTheZip(t *testing.T) {
	dir := t.TempDir()
	makeRawFolder(t, dir)
	writeFile(t, filepath.Join(dir, "results.txt"), "results")

	if err := run([]string{dir}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	zipPath := filepath.Join(dir, "sample_1.raw.zip")
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	contents := map[string]string{}
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "/") {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", file.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", file.Name, err)
		}
		contents[file.Name] = string(data)
	}
	reader.Close()
	if len(contents) != 2 ||
		contents["sample_1.raw/_HEADER.TXT"] != "header" ||
		contents["sample_1.raw/sub/_FUNC001.DAT"] != strings.Repeat("data", 1000) {
		t.Fatalf("zip does not hold the folder with its structure: %v", contents)
	}

	zipBytes, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	sum := sha1.Sum(zipBytes)
	want := "# SHA-1 Checksum \n" +
		"results.txt\tcdf7e925f5746741c316f5fbcf39ad0dfca90775\n" +
		"sample_1.raw.zip\t" + hex.EncodeToString(sum[:]) + "\n"
	got, err := os.ReadFile(filepath.Join(dir, checksumFileName))
	if err != nil {
		t.Fatalf("read checksum: %v", err)
	}
	if string(got) != want {
		t.Fatalf("checksum.txt:\n%s\nwant:\n%s", got, want)
	}

	// A second run reuses the zip, so the uploaded file and its checksum stay the same.
	if err := run([]string{dir}); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	again, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read zip again: %v", err)
	}
	if string(again) != string(zipBytes) {
		t.Fatal("second run rewrote the zip")
	}
}

func TestRunRefusesZipThatNoLongerMatchesTheFolder(t *testing.T) {
	dir := t.TempDir()
	rawDir := makeRawFolder(t, dir)
	if err := run([]string{dir}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	zipPath := filepath.Join(dir, "sample_1.raw.zip")
	before, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	checksumBefore, err := os.ReadFile(filepath.Join(dir, checksumFileName))
	if err != nil {
		t.Fatalf("read checksum: %v", err)
	}

	for name, change := range map[string]func(){
		"added file":   func() { writeFile(t, filepath.Join(rawDir, "_extern.inf"), "new") },
		"resized file": func() { writeFile(t, filepath.Join(rawDir, "_HEADER.TXT"), "longer header") },
	} {
		change()
		err := run([]string{dir})
		if err == nil || !strings.Contains(err.Error(), "does not match the folder") {
			t.Fatalf("%s: expected a stale zip error, got %v", name, err)
		}
	}

	after, _ := os.ReadFile(zipPath)
	checksumAfter, _ := os.ReadFile(filepath.Join(dir, checksumFileName))
	if string(after) != string(before) || string(checksumAfter) != string(checksumBefore) {
		t.Fatal("a refused run changed the zip or checksum.txt")
	}
}

func TestRunRejectsEmptyWatersRawFolder(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "empty.raw"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(dir, "results.txt"), "results")

	if err := run([]string{dir}); err == nil {
		t.Fatal("expected an empty .raw folder to fail")
	}
	if _, err := os.Stat(filepath.Join(dir, checksumFileName)); !os.IsNotExist(err) {
		t.Fatalf("checksum.txt should not be written, stat err: %v", err)
	}
}

func TestRunRefusesDamagedZip(t *testing.T) {
	dir := t.TempDir()
	makeRawFolder(t, dir)
	if err := run([]string{dir}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	zipPath := filepath.Join(dir, "sample_1.raw.zip")

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	var offset int64
	for _, file := range reader.File {
		if file.Name == "sample_1.raw/sub/_FUNC001.DAT" {
			if offset, err = file.DataOffset(); err != nil {
				t.Fatalf("data offset: %v", err)
			}
		}
	}
	reader.Close()

	damaged, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	damaged[offset+5] ^= 0xff
	writeFile(t, zipPath, string(damaged))

	if err := run([]string{dir}); err == nil || !strings.Contains(err.Error(), "damaged") {
		t.Fatalf("expected a damaged zip error, got %v", err)
	}
}

func TestRunRemovesLeftoverTempZip(t *testing.T) {
	dir := t.TempDir()
	makeRawFolder(t, dir)
	if err := run([]string{dir}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	writeFile(t, filepath.Join(dir, "sample_1.raw.zip.tmp"), "partial")

	if err := run([]string{dir}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, checksumFileName))
	if err != nil {
		t.Fatalf("read checksum: %v", err)
	}
	if strings.Contains(string(got), ".tmp") {
		t.Fatalf("leftover temporary zip was checksummed:\n%s", got)
	}
}
