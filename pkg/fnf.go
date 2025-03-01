package scnnr

import (
	"os"
	"runtime"
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

	if runtime.GOOS == "windows" {
		fnf.Direction = "\\"
	} else {
		fnf.Direction = "/"
	}

	fnf.Keywords = keywords

	return fnf
}

// Scan is a concurrent/parallel directory walker
func (f *FileNameFinder) Scan(directory string) {
	_, err := os.ReadDir(directory)

	if err != nil {
		panic(err)
	}

	f.findFiles(directory)
}

func (f *FileNameFinder) findFiles(directory string) {
	paths, _ := os.ReadDir(directory)

	var dirs []os.FileInfo
	var files []os.FileInfo

	for _, path := range paths {
		fullPath := directory + f.Direction + path.Name()

		if path.IsDir() {
			p, _ := os.Stat(fullPath)

			dirs = append(dirs, p)
		} else {

			f, _ := os.Stat(fullPath)

			files = append(files, f)
		}
	}

	for _, file := range files {
		if file != nil {
			for _, keyword := range f.Keywords {
				if strings.Contains(file.Name(), keyword) {
					f.mutex.Lock()

					fullFilePath := directory + f.Direction + file.Name()

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
