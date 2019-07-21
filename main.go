package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/selfup/scnnr/scnnr"
)

func main() {
	var directory string
	var extensions []string
	var keywords []string

	var dir string
	flag.StringVar(&dir, "d", "", `REQUIRED
    directory where scnnr will scan`)

	var ext string
	flag.StringVar(&ext, "e", "", `REQUIRED
    a comma delimted list of file extensions to scan`)

	var kwd string
	flag.StringVar(&kwd, "k", "", `REQUIRED
    a comma delimted list of keywords to search for in a file`)

	var rgx bool
	flag.BoolVar(&rgx, "r", false, `OPTIONAL
    wether to use the regex engine or not
    defaults to false and will not use the regex engine for scans unless set to a truthy value
    truthy values are: 1, t, T, true, True, TRUE
    flasey values are: 0, f, F, false, False, FALSE`)

	flag.Parse()

	if dir == "" && kwd == "" && ext == "" {
		flag.PrintDefaults()
		fmt.Print("\nERROR - scannr has required arguments - please read above output - exiting..\n\n")
		os.Exit(1)
	}

	directory = dir
	extensions = strings.Split(ext, ",")
	keywords = strings.Split(kwd, ",")

	scanner := scnnr.Scanner{
		Regex:          rgx,
		Keywords:       keywords,
		Directory:      directory,
		FileExtensions: extensions,
	}

	err := scanner.Scan()

	if err != nil {
		log.Fatal(err)
	}
}
