package main

import (
	"flag"
	"log"
	"strings"

	"github.com/selfup/scnnr/scnnr"
)

func main() {
	var directory string
	var extensions []string
	var keywords []string

	var dir string
	flag.StringVar(&dir, "dir", "", "directory where scnnr will scan")

	var ext string
	flag.StringVar(&ext, "ext", "", "a comma delimted list of file extensions to search")

	var kwd string
	flag.StringVar(&kwd, "kwd", "", "a comma delimted list of keywords to search for in a file")

	var rgx bool
	flag.BoolVar(&rgx, "rgx", false, "wether to use the regex engine or not - defaults to false")

	flag.Parse()

	if dir == "" && kwd == "" && ext == "" {
		log.Fatal("please use the -h flag to see how to use this tool")
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
