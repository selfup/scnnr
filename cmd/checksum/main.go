package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/selfup/scnnr/internal/digest"
)

const (
	source      = "scnnr_bins.zip"
	destination = "scnnr_bins.zip.sha256"
)

func main() {
	if err := run(source, destination, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func checksum(r io.Reader, name string, stdout io.Writer) (string, error) {
	sum, err := digest.SHA256(r)
	if err != nil {
		return "", err
	}
	if _, err := fmt.Fprintln(stdout, sum); err != nil {
		return "", err
	}
	return sum + "  " + name + "\n", nil
}

func run(source, destination string, stdout io.Writer) (err error) {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	line, err := checksum(file, source, stdout)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, []byte(line), 0644)
}
