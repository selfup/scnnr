package scnnr

import (
	"io/ioutil"
	"os"
	"runtime"
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

	if runtime.GOOS == "windows" {
		fsf.Direction = "\\"
	} else {
		fsf.Direction = "/"
	}

	switch size {
	case "1MB":
		fsf.Size = 1000000
		break
	case "10MB":
		fsf.Size = 10000000
		break
	case "100MB":
		fsf.Size = 100000000
		break
	case "1GB":
		fsf.Size = 1000000000
		break
	case "10GB":
		fsf.Size = 10000000000
		break
	case "100GB":
		fsf.Size = 1000000000000
		break
	case "1TB":
		fsf.Size = 1000000000000000
		break
	default:
		panic("please provide a size 1MB 10MB 100MB 1GB 10GB 100GB 1TB")
	}

	return fsf
}

// Scan is a concurrent/parallel directory walker
func (f *FileSizeFinder) Scan(directory string) {
	_, err := ioutil.ReadDir(directory)
	if err != nil {
		panic(err)
	}

	f.findFiles(directory, "")
}

func (f *FileSizeFinder) findFiles(directory string, prefix string) {
	paths, _ := ioutil.ReadDir(directory)

	var dirs []os.FileInfo
	var files []os.FileInfo

	for _, path := range paths {
		if path.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
	}

	for _, file := range files {
		if file.Size() >= f.Size {
			f.mutex.Lock()
			f.Files = append(f.Files, directory+f.Direction+file.Name())
			f.mutex.Unlock()
		}
	}

	dirLen := len(dirs)
	if dirLen > 0 {
		var dirGroup sync.WaitGroup
		dirGroup.Add(dirLen)

		for _, dir := range dirs {
			go func(diR os.FileInfo, direcTory string, direcTion string) {
				f.findFiles(direcTory+direcTion+diR.Name(), direcTory)
				dirGroup.Done()
			}(dir, directory, f.Direction)
		}

		dirGroup.Wait()
	}
}
