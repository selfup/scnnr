package scnnr

import (
	"fmt"
	"io/ioutil"
	"os"
)

// ParWalk is a concurrent/parallel directory walker
func ParWalk(directory string) ([]os.FileInfo, []string, error) {
	var dirs []os.FileInfo
	var files []os.FileInfo
	var badFiles []string

	paths, err := ioutil.ReadDir(directory)
	if err != nil {
		return files, badFiles, err
	}

	for _, path := range paths {
		if path.IsDir() {
			fmt.Println(path.Name())
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
	}

	return files, badFiles, nil
}
