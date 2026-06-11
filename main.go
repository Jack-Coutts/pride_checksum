package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	checksumFileName = "checksum.txt"
	tempFileName     = "checksum.txt.tmp"
	bufferSize       = 4 * 1024 * 1024
)

type fileResult struct {
	name string
	sum  string
}

type hintedError struct {
	err   error
	hints []string
}

func (e hintedError) Error() string {
	return e.err.Error()
}

func (e hintedError) Unwrap() error {
	return e.err
}

func withHints(err error, hints ...string) error {
	return hintedError{err: err, hints: hints}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Error:", err)

		var hinted hintedError
		if errors.As(err, &hinted) && len(hinted.hints) > 0 {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "What to try:")
			for _, hint := range hinted.hints {
				fmt.Fprintln(os.Stderr, "-", hint)
			}
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: pride-checksum-helper <folder>")
	}

	dataFolder, err := filepath.Abs(args[0])
	if err != nil {
		return withHints(
			fmt.Errorf("could not resolve folder path: %w", err),
			"Check that the dragged path still exists and does not contain malformed characters.",
		)
	}

	info, err := os.Stat(dataFolder)
	if err != nil {
		return withHints(
			fmt.Errorf("could not access folder %q: %w", dataFolder, err),
			"Check that the folder exists and that you have permission to read it.",
			"If this is a mapped network drive, check that the connection is still active.",
		)
	}
	if !info.IsDir() {
		return fmt.Errorf("this does not look like a folder: %s", dataFolder)
	}

	files, err := collectFiles(dataFolder)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no submission files found in %s", dataFolder)
	}

	fmt.Println("Creating PRIDE checksum file for:")
	fmt.Println(dataFolder)
	fmt.Println()

	results := make([]fileResult, 0, len(files))
	for i, file := range files {
		sum, err := hashFile(i+1, len(files), file.path, file.name)
		if err != nil {
			return err
		}

		fmt.Printf("[ %d / %d ] Generated checksum for: %s -> %s\n", i+1, len(files), file.name, sum)
		results = append(results, fileResult{name: file.name, sum: sum})
	}

	if err := writeChecksumFile(dataFolder, results); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Done.")
	fmt.Println("checksum.txt was saved in:")
	fmt.Println(dataFolder)
	return nil
}

type submissionFile struct {
	name string
	path string
	size int64
}

func collectFiles(dataFolder string) ([]submissionFile, error) {
	entries, err := os.ReadDir(dataFolder)
	if err != nil {
		return nil, withHints(
			fmt.Errorf("could not list folder %q: %w", dataFolder, err),
			"Check that you have permission to list the folder.",
			"If this is a mapped network drive, check that the connection is still active.",
		)
	}

	files := make([]submissionFile, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.EqualFold(name, checksumFileName) || strings.EqualFold(name, tempFileName) {
			continue
		}

		if entry.IsDir() {
			return nil, fmt.Errorf("PRIDE checksum input must contain files only; found folder %q", name)
		}
		if err := validateFileName(name); err != nil {
			return nil, err
		}

		fullPath := filepath.Join(dataFolder, name)
		info, err := entry.Info()
		if err != nil {
			return nil, withHints(
				fmt.Errorf("could not inspect %q: %w", fullPath, err),
				"Check that the file still exists and is not locked by another program.",
			)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("PRIDE checksum input must contain regular files only; found %q", name)
		}

		files = append(files, submissionFile{name: name, path: fullPath, size: info.Size()})
	}

	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
	})

	return files, nil
}

func validateFileName(name string) error {
	if name == "" {
		return errors.New("found a file with an empty name")
	}
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("invalid filename %q: hidden files are not allowed", name)
	}

	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '_' || r == '-' || r == '.' {
			continue
		}
		return fmt.Errorf("invalid filename %q: character %q is not allowed; use only letters, numbers, underscores, hyphens, and dots", name, r)
	}

	return nil
}

func hashFile(index, total int, path, name string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", withHints(
			fmt.Errorf("could not open %q: %w", path, err),
			"Check that the file is not open in another program.",
			"If this is a mapped network drive, check that the connection is still active.",
		)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", withHints(
			fmt.Errorf("could not inspect %q: %w", path, err),
			"Check that the file still exists and is not locked by another program.",
		)
	}

	fmt.Printf("[ %d / %d ] Processing: %s\n", index, total, path)

	hash := sha1.New()
	buffer := make([]byte, bufferSize)
	var readTotal int64
	lastProgress := time.Now()

	for {
		n, readErr := file.Read(buffer)
		if n > 0 {
			readTotal += int64(n)
			if _, err := hash.Write(buffer[:n]); err != nil {
				return "", fmt.Errorf("could not hash %q: %w", path, err)
			}

			if info.Size() > 0 && time.Since(lastProgress) >= time.Second {
				fmt.Printf("[ %d / %d ] Progress: %s %.1f%%\n", index, total, name, float64(readTotal)*100/float64(info.Size()))
				lastProgress = time.Now()
			}
		}

		if readErr == nil {
			continue
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		return "", withHints(
			fmt.Errorf("could not read %q after %d bytes: %w", path, readTotal, readErr),
			"Check that the file is not open in another program and was fully copied before running the checksum.",
			"If this is a mapped network drive, try running again or copying the file to a local drive first.",
		)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeChecksumFile(dataFolder string, results []fileResult) error {
	tempPath := filepath.Join(dataFolder, tempFileName)
	finalPath := filepath.Join(dataFolder, checksumFileName)

	file, err := os.Create(tempPath)
	if err != nil {
		return withHints(
			fmt.Errorf("could not create %q: %w", tempPath, err),
			"Check that you have permission to write to the data folder.",
			"Close checksum.txt if it is open in another program.",
		)
	}

	writeErr := writeChecksumContents(file, results)
	closeErr := file.Close()
	if writeErr != nil {
		_ = os.Remove(tempPath)
		return writeErr
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return withHints(
			fmt.Errorf("could not finish writing %q: %w", tempPath, closeErr),
			"Check that there is enough free space and that the network drive is still connected.",
		)
	}

	existingChecksum, existingMode, err := readExistingChecksum(finalPath)
	if err != nil {
		_ = os.Remove(tempPath)
		return err
	}

	if err := os.Remove(finalPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(tempPath)
		return withHints(
			fmt.Errorf("could not replace existing %q: %w", finalPath, err),
			"Close checksum.txt if it is open in another program.",
			"Check that you have permission to replace files in the data folder.",
		)
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		if restoreErr := restoreExistingChecksum(finalPath, existingChecksum, existingMode); restoreErr != nil {
			return withHints(
				fmt.Errorf("could not save %q: %w; also could not restore the previous checksum.txt: %v", finalPath, err, restoreErr),
				"Check that the network drive is still connected and that you have permission to write to the folder.",
			)
		}
		return withHints(
			fmt.Errorf("could not save %q: %w", finalPath, err),
			"The previous checksum.txt was restored.",
			"Check that the network drive is still connected and that you have permission to write to the folder.",
		)
	}

	return nil
}

func readExistingChecksum(path string) ([]byte, os.FileMode, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, withHints(
			fmt.Errorf("could not inspect existing %q: %w", path, err),
			"Close checksum.txt if it is open in another program.",
		)
	}
	if !info.Mode().IsRegular() {
		return nil, 0, fmt.Errorf("could not replace %q because it is not a regular file", path)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, withHints(
			fmt.Errorf("could not read existing %q before replacing it: %w", path, err),
			"Close checksum.txt if it is open in another program.",
		)
	}

	return contents, info.Mode().Perm(), nil
}

func restoreExistingChecksum(path string, contents []byte, mode os.FileMode) error {
	if contents == nil {
		return nil
	}
	if mode == 0 {
		mode = 0o666
	}
	return os.WriteFile(path, contents, mode)
}

func writeChecksumContents(w io.Writer, results []fileResult) error {
	if _, err := fmt.Fprintln(w, "# SHA-1 Checksum "); err != nil {
		return fmt.Errorf("could not write checksum header: %w", err)
	}

	for _, result := range results {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", result.name, result.sum); err != nil {
			return fmt.Errorf("could not write checksum for %q: %w", result.name, err)
		}
	}

	return nil
}
