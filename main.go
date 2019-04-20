package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/selfup/scnnr/cmd"
)

func main() {
	var directory string
	var patterns []string
	var keywords []string

	if len(os.Args) > 3 {
		directory = os.Args[1]
		patterns = strings.Split(os.Args[2], ",")
		keywords = strings.Split(os.Args[3], ",")
	}

	fmt.Println("File patterns: ", patterns, "\nKeywords: ", keywords)

	scanner := cmd.Scanner{
		Keywords:     keywords,
		Directory:    directory,
		FilePatterns: patterns,
	}

	err := scanner.Scan()

	if err != nil {
		log.Fatal(err)
	}
}
