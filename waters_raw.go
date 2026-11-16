package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Waters instruments save each run as a ".raw" folder. PRIDE takes files only
// and asks for each such folder as its own archive with the folder structure kept.

type rawEntry struct {
	name  string // slash-separated, starts with the .raw folder name
	path  string
	size  int64
	mod   time.Time
	isDir bool
}

func isWatersRawFolder(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".raw")
}

// zipWatersRawFolders makes "<folder>.raw.zip" beside each .raw folder and
// returns how many .raw folders it found.
func zipWatersRawFolders(dataFolder string) (int, error) {
	entries, err := os.ReadDir(dataFolder)
	if err != nil {
		return 0, withHints(
			fmt.Errorf("could not list folder %q: %w", dataFolder, err),
			"Check that you have permission to list the folder.",
			"If this is a mapped network drive, check that the connection is still active.",
		)
	}

	var folders []string
	for _, entry := range entries {
		if entry.IsDir() && isWatersRawFolder(entry.Name()) {
			folders = append(folders, entry.Name())
		}
	}

	for i, folder := range folders {
		if err := ensureRawZip(i+1, len(folders), dataFolder, folder); err != nil {
			return 0, err
		}
	}
	return len(folders), nil
}

func ensureRawZip(index, total int, dataFolder, folder string) error {
	zipName := folder + ".zip"
	zipPath := filepath.Join(dataFolder, zipName)

	entries, err := rawFolderEntries(dataFolder, folder)
	if err != nil {
		return err
	}

	if _, err := os.Stat(zipPath); err == nil {
		if err := checkExistingRawZip(zipPath, folder, entries); err != nil {
			return err
		}
		fmt.Printf("[ %d / %d ] Already zipped, reusing: %s\n", index, total, zipName)
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return withHints(
			fmt.Errorf("could not inspect %q: %w", zipPath, err),
			"Check that you have permission to read the data folder.",
		)
	}

	fmt.Printf("[ %d / %d ] Zipping Waters .raw folder: %s\n", index, total, folder)
	if err := writeRawZip(zipPath, entries, func(percent float64) {
		fmt.Printf("[ %d / %d ] Zip progress: %s %.1f%%\n", index, total, folder, percent)
	}); err != nil {
		return withHints(
			fmt.Errorf("could not zip Waters .raw folder %q: %w", folder, err),
			"Check that there is enough free space and that you can write to the data folder.",
			"Check that no file in the .raw folder is open in another program.",
		)
	}
	return nil
}

func rawFolderEntries(dataFolder, folder string) ([]rawEntry, error) {
	var entries []rawEntry
	files := 0
	err := filepath.WalkDir(filepath.Join(dataFolder, folder), func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dataFolder, path)
		if err != nil {
			return err
		}
		entry := rawEntry{name: filepath.ToSlash(rel), path: path, size: info.Size(), mod: info.ModTime(), isDir: d.IsDir()}
		if !entry.isDir {
			if !info.Mode().IsRegular() {
				return fmt.Errorf("%q is not a regular file", path)
			}
			files++
		}
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		return nil, withHints(
			fmt.Errorf("could not read Waters .raw folder %q: %w", folder, err),
			"Check that the folder is complete and readable.",
		)
	}
	if files == 0 {
		return nil, fmt.Errorf("Waters .raw folder %q does not contain any files", folder)
	}
	return entries, nil
}

// checkExistingRawZip refuses to reuse a zip that does not hold the folder's files.
// shortcut: compares names and sizes only, so a same-size edit is missed; hash the contents if that ever bites.
func checkExistingRawZip(zipPath, folder string, entries []rawEntry) error {
	stale := func(reason string) error {
		return withHints(
			fmt.Errorf("%q does not match the folder %q: %s", filepath.Base(zipPath), folder, reason),
			"The folder may have changed since the zip was made, or the zip was made another way.",
			"Delete or move the zip, then run again to make a fresh one.",
		)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return stale(fmt.Sprintf("the zip could not be read (%v)", err))
	}
	defer reader.Close()

	zipped := make(map[string]int64)
	for _, file := range reader.File {
		name := strings.ReplaceAll(file.Name, "\\", "/")
		if !strings.HasSuffix(name, "/") {
			zipped[name] = int64(file.UncompressedSize64)
		}
	}

	files := 0
	for _, entry := range entries {
		if entry.isDir {
			continue
		}
		files++
		size, ok := zipped[entry.name]
		if !ok {
			return stale(fmt.Sprintf("%s is missing from the zip", entry.name))
		}
		if size != entry.size {
			return stale(fmt.Sprintf("%s has a different size in the zip", entry.name))
		}
	}
	if len(zipped) != files {
		return stale("the zip holds files that are not in the folder")
	}
	return nil
}

func writeRawZip(zipPath string, entries []rawEntry, progress func(percent float64)) error {
	var total int64
	for _, entry := range entries {
		total += entry.size
	}

	tmpPath := zipPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(file)
	err = addRawEntries(zipWriter, entries, total, progress)
	if closeErr := zipWriter.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmpPath, zipPath)
	}
	if err != nil {
		_ = os.Remove(tmpPath)
	}
	return err
}

func addRawEntries(zipWriter *zip.Writer, entries []rawEntry, total int64, progress func(percent float64)) error {
	buffer := make([]byte, bufferSize)
	var done int64
	lastProgress := time.Now()

	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate, Modified: entry.mod}
		if entry.isDir {
			header.Name += "/"
			header.Method = zip.Store
		}
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if entry.isDir {
			continue
		}

		source, err := os.Open(entry.path)
		if err != nil {
			return err
		}
		for {
			n, readErr := source.Read(buffer)
			if n > 0 {
				if _, err := writer.Write(buffer[:n]); err != nil {
					source.Close()
					return err
				}
				done += int64(n)
				if time.Since(lastProgress) >= time.Second {
					progress(float64(done) * 100 / float64(total))
					lastProgress = time.Now()
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				source.Close()
				return readErr
			}
		}
		source.Close()
	}
	return nil
}
