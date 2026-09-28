package scnnr

import (
	"os"
	"strings"
	"sync"
)

// FileNameFinder finds filenames containing the supplied keywords.
type FileNameFinder struct {
	mutex     sync.Mutex
	Direction string
	Files     []string
	Keywords  []string
}

func NewFileNameFinder(keywords []string) *FileNameFinder {
	return &FileNameFinder{Direction: PathDirection(), Keywords: keywords}
}

// Scan preserves the original void API and root-directory panic. Results append
// as files are visited, including a result for each matching keyword.
func (f *FileNameFinder) Scan(directory string) {
	CheckDirOrPanic(directory)

	f.findFiles(directory)
}

func (f *FileNameFinder) findFiles(directory string) {
	walkFinderFiles(directory, f.Direction, func(directory string, file os.FileInfo) {
		// Keep the original name finder's path construction for compatibility.
		path := FullFilePath(directory, directory, file)

		for _, keyword := range f.Keywords {
			if strings.Contains(file.Name(), keyword) {
				f.mutex.Lock()
				f.Files = append(f.Files, path)
				f.mutex.Unlock()
			}
		}
	})
}

func containsAny(value string, keywords []string) bool {
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(value, keyword) {
			return true
		}
	}

	return false
}
