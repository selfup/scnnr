package main

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func main() {
	err := run("scnnr_bins", "scnnr_bins.zip", func() error {
		cmd := exec.Command("go", "run", "cmd/build/main.go")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
	if err != nil {
		log.Fatal(err)
	}
}

func run(source, target string, build func() error) error {
	if err := build(); err != nil {
		return err
	}
	return zipBins(source, target)
}

func zipBins(source, target string) (err error) {
	if _, err := os.Stat(source); err != nil {
		return err
	}
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	return writeZip(file, source)
}

// writeZip closes the archive before returning so buffered write errors are
// observable. It leaves ownership of the supplied writer with the caller.
func writeZip(w io.Writer, source string) (err error) {
	if _, err := os.Stat(source); err != nil {
		return err
	}
	archive := zip.NewWriter(w)
	defer func() { err = errors.Join(err, archive.Close()) }()
	base := filepath.Base(filepath.Clean(source))
	return filepath.Walk(source, func(filename string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = base
		if info.IsDir() || filename != source {
			relative, err := filepath.Rel(source, filename)
			if err != nil {
				return err
			}
			header.Name = path.Join(base, filepath.ToSlash(relative))
		}
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		writer, err := archive.CreateHeader(header)
		if err != nil || info.IsDir() {
			return err
		}
		return copyFile(writer, filename)
	})
}

func copyFile(w io.Writer, filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	_, err = io.Copy(w, file)
	return err
}
