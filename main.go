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
	} else {
		log.Fatal(`
    
      Not enough args!

      scnnr <directory> <extension(s)> <keyword(s)>

      Extension(s) and Keywords(s) can be in csv format
      
      snnr <directory> .js,.jsx,.py cache=,cache:

      Example:

      $ go run main.go ./ .md,.markdown cache=,main.go
      File patterns:  [.md .markdown]
      Keywords:  [cache= main.go]
      Files matching pattern(s):  [README.md]
    `)
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
