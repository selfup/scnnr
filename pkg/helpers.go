package scnnr

import (
	"os"
	"runtime"
)

func PathDirection() string {
	if runtime.GOOS == "windows" {
		return "\\"
	} else {
		return "/"
	}
}

func CheckDirOrPanic(directory string) {
	_, err := os.ReadDir(directory)

	if err != nil {
		panic(err)
	}
}

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
