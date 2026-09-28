package scnnr

import (
	"os"
)

// FullFilePath uses the supplied direction verbatim, as in the original API.
func FullFilePath(directory, direction string, file os.FileInfo) string {
	return directory + direction + file.Name()
}

func PathDirection() string {
	return string(os.PathSeparator)
}

// CheckDirOrPanic panics when the root directory cannot be read.
func CheckDirOrPanic(directory string) {
	_, err := os.ReadDir(directory)

	if err != nil {
		panic(err)
	}
}

// CollectFilesAndDirs preserves the original best-effort directory collection.
func CollectFilesAndDirs(directory string, direction string) ([]os.FileInfo, []os.FileInfo) {
	paths, _ := os.ReadDir(directory)

	var dirs []os.FileInfo
	var files []os.FileInfo

	for _, path := range paths {
		pathName := path.Name()

		fullPath := directory + direction + pathName

		if path.IsDir() {
			p, _ := os.Stat(fullPath)

			dirs = append(dirs, p)
		} else {
			f, _ := os.Stat(fullPath)

			files = append(files, f)
		}
	}

	return files, dirs
}
