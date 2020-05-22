package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

var fixtureText = []byte("hello world")

func main() {
	MkDir("artifact")

	for i := 1; i <= 150000; i++ {
		fileName := fmt.Sprintf("artifact/file_%d.txt", i)

		Wr(fileName, fixtureText)
	}
}

// Wr writes a file
func Wr(destination string, contents []byte) error {
	return ioutil.WriteFile(destination, contents, 777)
}

// MkDir is like mkdir -p
func MkDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}
