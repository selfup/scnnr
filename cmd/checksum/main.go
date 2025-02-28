package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
)

const (
	source      = "scnnr_bins.zip"
	destination = "scnnr_bins.zip.sha256"
)

func main() {
	zip, err := os.ReadFile(source)
	if err != nil {
		log.Fatalln(err)
	}

	sum := sha256.Sum256(zip)
	sumStr := fmt.Sprintf("%x", sum)
	sumStrBytes := []byte(sumStr + "  " + source + "\n")

	writeErr := os.WriteFile(destination, sumStrBytes, os.ModePerm)
	if writeErr != nil {
		log.Fatalln(writeErr)
	}
}
