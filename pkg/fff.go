package scnnr

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"sync"
)

// FileFingerprintFinder struct contains needed data to perform concurrent operations
type FileFingerprintFinder struct {
	mutex        sync.Mutex
	Direction    string
	Files        []string
	FingerPrints []string
}

// NewFileFingerprintFinder creates a pointer to FileFingerprintFinder with default values
func NewFileFingerprintFinder(fingerPrints []string) *FileFingerprintFinder {
	fnf := new(FileFingerprintFinder)

	fnf.Direction = PathDirection()

	fnf.FingerPrints = fingerPrints

	return fnf
}

// Scan is a concurrent/parallel directory walker
func (f *FileFingerprintFinder) Scan(directory string) {
	CheckDirOrPanic(directory)
	f.findFiles(directory)
}

func (f *FileFingerprintFinder) findFiles(directory string) {
	files, dirs := CollectFilesAndDirs(directory, f.Direction)

	for _, file := range files {
		if file != nil {
			fullFilePath := FullFilePath(directory, f.Direction, file)

			zip, _ := os.ReadFile(fullFilePath)

			sum := sha256.Sum256(zip)
			sumStr := fmt.Sprintf("%x", sum)

			for _, fingerPrint := range f.FingerPrints {
				shaString := string(sumStr)

				if strings.Contains(shaString, fingerPrint) {
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
