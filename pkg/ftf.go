package scnnr

import (
	"os"
	"sync"
)

// FileTotalFinder struct contains needed data to perform concurrent operations
type FileTotalFinder struct {
	mutex            sync.Mutex
	Files            []string
	Direction        string
	Directory        string
	Keywords         []string
	KeywordMatches   []string
	MatchedFilePaths []FileData
}

// NewFileTotalFinder creates a pointer to FileTotalFinder with default values
func NewFileTotalFinder(keywords []string) *FileTotalFinder {
	fileTotalFinder := new(FileTotalFinder)

	fileTotalFinder.Direction = PathDirection()

	fileTotalFinder.Keywords = keywords

	return fileTotalFinder
}

// Scan is a concurrent/parallel directory walker
func (f *FileTotalFinder) Scan(directory string) {
	CheckDirOrPanic(directory)

	f.findFiles(directory)
}

func (f *FileTotalFinder) findFiles(directory string) {
	files, dirs := CollectFilesAndDirs(directory, f.Direction)

	for _, file := range files {
		if file != nil {
			fullFilePath := FullFilePath(directory, f.Direction, file)

			f.mutex.Lock()

			f.Files = append(f.Files, fullFilePath)

			f.mutex.Unlock()
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
