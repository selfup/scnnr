package scnnr

import (
	"os"
	"strings"
	"sync"
)

// FileNameFinder struct contains needed data to perform concurrent operations
type FileNameFinder struct {
	mutex     sync.Mutex
	Direction string
	Files     []string
	Keywords  []string
}

// NewFileNameFinder creates a pointer to FileNameFinder with default values
func NewFileNameFinder(keywords []string) *FileNameFinder {
	fnf := new(FileNameFinder)

	fnf.Direction = PathDirection()

	fnf.Keywords = keywords

	return fnf
}

// Scan is a concurrent/parallel directory walker
func (f *FileNameFinder) Scan(directory string) {
	CheckDirOrPanic(directory)

	f.findFiles(directory)
}

func (f *FileNameFinder) findFiles(directory string) {
	files, dirs := CollectFilesAndDirs(directory, f.Direction)

	for _, file := range files {
		if file != nil {
			fileName := file.Name()

			fullFilePath := FullFilePath(directory, directory, file)

			for _, keyword := range f.Keywords {
				if strings.Contains(fileName, keyword) {
					f.mutex.Lock()

					f.Files = append(f.Files, fullFilePath)

					f.mutex.Unlock()
				}
			}
		}
	}

	dirLen := len(dirs)
	if dirLen > 0 {
		var dirGroup sync.WaitGroup

		dirGroup.Add(dirLen)

		for _, dir := range dirs {
			go func(dirInfo os.FileInfo, dirName string, pathDirection string) {
				f.findFiles(dirName + pathDirection + dirInfo.Name())

				dirGroup.Done()
			}(dir, directory, f.Direction)
		}

		dirGroup.Wait()
	}
}
