package scnnr

import (
	"fmt"
	"os"
	"sync"
)

// FileSizeFinder finds files at least Size bytes long.
type FileSizeFinder struct {
	mutex     sync.Mutex
	Files     []string
	Direction string
	Size      int64
}

// ParseSize converts one of the supported decimal size thresholds into bytes.
func ParseSize(size string) (int64, error) {
	switch size {
	case "1MB":
		return 1000000, nil
	case "10MB":
		return 10000000, nil
	case "100MB":
		return 100000000, nil
	case "1GB":
		return 1000000000, nil
	case "10GB":
		return 10000000000, nil
	case "100GB":
		return 100000000000, nil
	case "1TB":
		return 1000000000000, nil
	default:
		return 0, fmt.Errorf("invalid size %q: please provide a size 1MB 10MB 100MB 1GB 10GB 100GB 1TB", size)
	}
}

// NewFileSizeFinder preserves the original panic for unsupported thresholds.
func NewFileSizeFinder(size string) *FileSizeFinder {
	threshold, err := ParseSize(size)

	if err != nil {
		panic("please provide a size 1MB 10MB 100MB 1GB 10GB 100GB 1TB")
	}

	return &FileSizeFinder{Direction: PathDirection(), Size: threshold}
}

// Scan preserves the original void API, root panic, and append behavior.
func (f *FileSizeFinder) Scan(directory string) {
	CheckDirOrPanic(directory)

	f.findFiles(directory)
}

func (f *FileSizeFinder) findFiles(directory string) {
	walkFinderFiles(directory, f.Direction, func(directory string, file os.FileInfo) {
		if file.Size() >= f.Size {
			// Keep the original size finder's path construction for compatibility.
			path := FullFilePath(directory, directory, file)

			f.mutex.Lock()

			f.Files = append(f.Files, path)

			f.mutex.Unlock()
		}
	})
}
