package scnnr

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

// walkFiles shares traversal and error handling across scanner modes. Symlinked
// files are read through their targets; symlinked directories are not traversed.
func walkFiles(root string, excludeDirs []string, include func(string) bool, visit func(FileData) error) error {
	if root == "" {
		return errors.New("directory is required")
	}

	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if slices.Contains(excludeDirs, entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply path filters before stat so excluded entries require no access.
		if include != nil && !include(path) {
			return nil
		}

		info, err := entry.Info()
		if entry.Type()&os.ModeSymlink != 0 {
			info, err = os.Stat(path)
		}

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return visit(FileData{Path: path, Info: info})
	})
}

// walkFinderFiles retains the finders' original parallel traversal: failed
// directory reads and nil file metadata are skipped, and sibling walks continue.
func walkFinderFiles(directory, direction string, visit func(string, os.FileInfo)) {
	files, dirs := CollectFilesAndDirs(directory, direction)
	for _, file := range files {
		if file != nil {
			visit(directory, file)
		}
	}

	var wg sync.WaitGroup

	wg.Add(len(dirs))

	for _, dir := range dirs {
		go func() {
			defer wg.Done()
			walkFinderFiles(directory+direction+dir.Name(), direction, visit)
		}()
	}

	wg.Wait()
}
