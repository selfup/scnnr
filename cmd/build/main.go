package main

import (
	"log"
	"os"
	"os/exec"
)

const BIN = "/scnnr"
const EXE = "/scnnr.exe"

const LINUX_ROOT = "scnnr_bins/linux"
const LINUX_INTEL = LINUX_ROOT + "/intel"
const LINUX_INTEL_BIN = LINUX_INTEL + BIN
const LINUX_ARM = LINUX_ROOT + "/arm"
const LINUX_ARM_BIN = LINUX_ARM + BIN

const MAC_ROOT = "scnnr_bins/mac"
const MAC_INTEL = MAC_ROOT + "/intel"
const MAC_INTEL_BIN = MAC_INTEL + BIN
const MAC_ARM = MAC_ROOT + "/arm"
const MAC_ARM_BIN = MAC_ARM + BIN

const WINDOWS = "scnnr_bins/windows"
const WINDOWS_BIN = WINDOWS + EXE

func main() {
	os.MkdirAll(MAC_ROOT, os.ModePerm)
	os.MkdirAll(MAC_INTEL, os.ModePerm)
	os.MkdirAll(MAC_ARM, os.ModePerm)

	os.MkdirAll(LINUX_ROOT, os.ModePerm)
	os.MkdirAll(LINUX_INTEL, os.ModePerm)
	os.MkdirAll(LINUX_ARM, os.ModePerm)

	os.MkdirAll(WINDOWS, os.ModePerm)

	if fileExists("main") {
		deleteFile("main")
	}

	if fileExists("main.exe") {
		deleteFile("main.exe")
	}

	compile("darwin", "amd64")
	mv("main", MAC_INTEL_BIN)

	compile("darwin", "arm64")
	mv("main", MAC_ARM_BIN)

	compile("linux", "amd64")
	mv("main", LINUX_INTEL_BIN)

	compile("linux", "arm64")
	mv("main", LINUX_ARM_BIN)

	compile("windows", "386")
	mv("main.exe", WINDOWS_BIN)

	nixFiles := []string{LINUX_ARM_BIN, LINUX_INTEL_BIN, MAC_ARM_BIN, MAC_INTEL_BIN}

	chmod(nixFiles, 0777)

	ci := os.Getenv("CI")
	version := os.Getenv("VERSION")

	if ci != "" && version != "" {
		versionNumber := []byte(version + "\n")

		err := os.WriteFile("scnnr_bins/version", versionNumber, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	cp("README.md", "scnnr_bins/README.md")
	cp("LICENSE", "scnnr_bins/LICENSE")
}

func compile(goos string, arch string) {
	setEnv("CGO_ENABLED", "0")
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
	input, err := os.ReadFile(source)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(destination, input, 0777)
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
