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
	"os"
	"strings"

	scnnr "github.com/selfup/scnnr/pkg"
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
		fmt.Print("\nERROR - scnnr has required arguments - please read above output - exiting..\n\n")
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
