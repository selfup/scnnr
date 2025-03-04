/**
The MIT License

Copyright (c) 2019 Regis Boudinot (selfup) https://selfup.me

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	scnnr "github.com/selfup/scnnr/pkg"
)

func main() {
	var directory string
	var extensions []string
	var keywords []string

	var mode string
	flag.StringVar(&mode, "m", "scn", `OPTIONAL
    mode that scnnr will run in

    options are:
	(scn) for scnnr (default)
	(fnf) for File/NameFinder
	(fsf) for File/SizeFinder
	(fff) for File/FingerprintFinder (uses SHA2-256)

    ex: scnnr -d $HOME/Documents -k password,token,authorization
    ex: scnnr -m fsf -d E:/ -s 100MB
    ex: scnnr -m fnf -p /tmp,$HOME/Documents -f DEFCON
    ex: scnnr -m fff -d $HOME/Documents -k de4f51f97fa690026e225798ff294cd182b93847aaa46fe1e32b848eb9e985bd

`)

	var dir string
	flag.StringVar(&dir, "d", ".", `OPTIONAL Scnnr MODE and OPTIONAL FingerrintFinder MODE
    directory where scnnr will scan
    default is current directory and all child directories`)

	var ext string
	flag.StringVar(&ext, "e", "", `OPTIONAL Scnnr MODE
    a comma delimited list of file extensions to scan
    if none are given all files will be searched`)

	var kwd string
	flag.StringVar(&kwd, "k", "", `OPTIONAL Scnnr MODE and REQUIRED FingerprintFinder
    scnnr: this is a comma delimited list of characters to look for in a file
        if no keywords are given - all file paths of given file extensions will be returned
        if keywords are given - only filepaths of matches will be returned
    FingerprintFinder: this is a comma delimited list of SHA2-256 hashes to find files by`)

	var rgx bool
	flag.BoolVar(&rgx, "r", false, `OPTIONAL Scnnr MODE
    if you want to use the regex engine or not
    defaults to false and will not use the regex engine for scans unless set to a truthy value
    truthy values are: 1, t, T, true, True, TRUE
    falsy values are: 0, f, F, false, False, FALSE`)

	var paths string
	flag.StringVar(&paths, "p", "", `REQUIRED NameFinder MODE
    any absolute path - can be comma delimited: Example: $HOME or '/tmp,/usr'`)

	var fuzzy string
	flag.StringVar(&fuzzy, "f", "", `REQUIRED NameFinder MODE
    fuzzy find the filename(s) contain(s) - can be comma delimited: Example 'wow' or 'wow,omg,lol'`)

	var size string
	flag.StringVar(&size, "s", "", `REQUIRED SizeFinder MODE
    size: 1MB,10MB,100MB,1GB,10GB,100GB,1TB`)

	flag.Parse()

	directory = dir

	if mode == "fnf" {
		scanFuzzy := strings.Split(fuzzy, ",")
		scanPaths := strings.Split(paths, ",")

		nfnf := scnnr.NewFileNameFinder(scanFuzzy)

		for _, path := range scanPaths {
			nfnf.Scan(path)
		}

		for _, file := range nfnf.Files {
			fmt.Println(file)
		}
	} else if mode == "scn" {
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
	} else if mode == "fsf" {
		nfsf := scnnr.NewFileSizeFinder(size)

		nfsf.Scan(directory)

		for _, file := range nfsf.Files {
			fmt.Println(file)
		}
	} else if mode == "fff" {
		keywords = strings.Split(kwd, ",")

		if keywords[0] == "" {
			log.Fatal("-k (known hashes) REQUIRED for FileFingerprintFinder")
		}

		nfff := scnnr.NewFileFingerprintFinder(keywords)

		nfff.Scan(directory)

		for _, file := range nfff.Files {
			fmt.Println(file)
		}
	}
}
