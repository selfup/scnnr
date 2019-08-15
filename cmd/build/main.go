package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func main() {
	os.MkdirAll("scnnr_bins/mac", os.ModePerm)
	os.MkdirAll("scnnr_bins/linux", os.ModePerm)
	os.MkdirAll("scnnr_bins/windows", os.ModePerm)

	if fileExists("main") {
		deleteFile("main")
	}

	if fileExists("main.exe") {
		deleteFile("main.exe")
	}

	compile("darwin", "amd64")
	mv("main", "scnnr_bins/mac/scnnr")

	compile("linux", "amd64")
	mv("main", "scnnr_bins/linux/scnnr")

	compile("windows", "386")
	mv("main.exe", "scnnr_bins/windows/scnnr.exe")

	nixFiles := []string{"scnnr_bins/linux/scnnr", "scnnr_bins/mac/scnnr"}

	chmod(nixFiles, 0777)

	ci := os.Getenv("CI")
	version := os.Getenv("VERSION")

	if ci != "" && version != "" {
		versionNumber := []byte(version + "\n")

		err := ioutil.WriteFile("scnnr_bins/version", versionNumber, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	cp("README.md", "scnnr_bins/README.md")
	cp("LICENSE", "scnnr_bins/LICENSE")
}

func compile(goos string, arch string) {
	setEnv("CGO_ENALED", "0")
	setEnv("GOOS", goos)
	setEnv("GOARCH", arch)

	cmd := exec.Command("go", "build", "main.go")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func mv(old string, new string) {
	err := os.Rename(old, new)
	if err != nil {
		log.Fatal(err)
	}
}

func cp(source string, destination string) {
	input, err := ioutil.ReadFile(source)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(destination, input, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func setEnv(key string, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		log.Fatal(err)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func chmod(files []string, mode os.FileMode) {
	for _, file := range files {
		err := os.Chmod(file, mode)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func deleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Fatal(err)
	}
}
