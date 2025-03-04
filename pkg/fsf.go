package scnnr

import (
	"os"
	"sync"
)

// FileSizeFinder struct contains needed data to perform concurrent operations
type FileSizeFinder struct {
	mutex     sync.Mutex
	Files     []string
	Direction string
	Size      int64
}

// NewFileSizeFinder creates a pointer to FileSizeFinder with default values
func NewFileSizeFinder(size string) *FileSizeFinder {
	fsf := new(FileSizeFinder)

	fsf.Direction = PathDirection()

	switch size {
	case "1MB":
		fsf.Size = 1000000
	case "10MB":
		fsf.Size = 10000000
	case "100MB":
		fsf.Size = 100000000
	case "1GB":
		fsf.Size = 1000000000
	case "10GB":
		fsf.Size = 10000000000
	case "100GB":
		fsf.Size = 100000000000
	case "1TB":
		fsf.Size = 1000000000000
	default:
		panic("please provide a size 1MB 10MB 100MB 1GB 10GB 100GB 1TB")
	}

	return fsf
}

// Scan is a concurrent/parallel directory walker
func (f *FileSizeFinder) Scan(directory string) {
	CheckDirOrPanic(directory)

	f.findFiles(directory)
}

func (f *FileSizeFinder) findFiles(directory string) {
	files, dirs := CollectFilesAndDirs(directory, f.Direction)

	for _, file := range files {
		if file != nil {
			fullFilePath := FullFilePath(directory, directory, file)

			if file.Size() >= f.Size {
				f.mutex.Lock()

				f.Files = append(f.Files, fullFilePath)

				f.mutex.Unlock()
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
