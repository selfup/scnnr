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
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	scnnr "github.com/selfup/scnnr/pkg"
)

type options struct {
	mode, directory, size                                        string
	extensions, keywords, paths, fuzzy, excludeDirs, excludeExts []string
	regex, showLines, showCols                                   bool
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		var parseErr *flagError

		if errors.As(err, &parseErr) {
			os.Exit(2)
		}

		log.Fatal(err)
	}
}

type flagError struct{ error }

func (e *flagError) Unwrap() error { return e.error }

func parseOptions(args []string, stderr io.Writer) (options, error) {
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)

	var mode string

	flags.StringVar(&mode, "m", "scn", `OPTIONAL
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
	flags.StringVar(&dir, "d", ".", `OPTIONAL Scnnr MODE and OPTIONAL FingerrintFinder MODE
    directory where scnnr will scan
    default is current directory and all child directories`)

	var ext string
	flags.StringVar(&ext, "e", "", `OPTIONAL Scnnr MODE
    a comma delimited list of file extensions to scan
    if none are given all files will be searched`)

	var kwd string
	flags.StringVar(&kwd, "k", "", `OPTIONAL Scnnr MODE and REQUIRED FingerprintFinder
    scnnr: this is a comma delimited list of characters to look for in a file
        if no keywords are given - all file paths of given file extensions will be returned
        if keywords are given - only filepaths of matches will be returned
    FingerprintFinder: this is a comma delimited list of SHA2-256 hashes to find files by`)

	var rgx bool
	flags.BoolVar(&rgx, "r", false, `OPTIONAL Scnnr MODE
    if you want to use the regex engine or not
    defaults to false and will not use the regex engine for scans unless set to a truthy value
    truthy values are: 1, t, T, true, True, TRUE
    falsy values are: 0, f, F, false, False, FALSE`)

	var showLines bool
	flags.BoolVar(&showLines, "l", false, `OPTIONAL Scnnr MODE
    show line numbers for each match (requires -k)
    when enabled, finds ALL matches in files (no early exit)`)

	var showCols bool
	flags.BoolVar(&showCols, "c", false, `OPTIONAL Scnnr MODE
    show line and column numbers for each match (requires -k)
    when enabled, finds ALL matches in files (no early exit)`)

	var paths string
	flags.StringVar(&paths, "p", "", `REQUIRED NameFinder MODE
    any absolute path - can be comma delimited: Example: $HOME or '/tmp,/usr'`)

	var fuzzy string
	flags.StringVar(&fuzzy, "f", "", `REQUIRED NameFinder MODE
    fuzzy find the filename(s) contain(s) - can be comma delimited: Example 'wow' or 'wow,omg,lol'`)

	var size string
	flags.StringVar(&size, "s", "", `REQUIRED SizeFinder MODE
    size: 1MB,10MB,100MB,1GB,10GB,100GB,1TB`)

	var excludeDirs string
	flags.StringVar(&excludeDirs, "xd", "", `OPTIONAL Scnnr MODE
    comma-delimited list of directory names to exclude from scanning
    ex: -xd ".git,node_modules,.venv"`)

	var excludeExts string
	flags.StringVar(&excludeExts, "xe", "", `OPTIONAL Scnnr MODE
    comma-delimited list of file extensions to exclude from scanning
    ex: -xe ".log,.tmp,.json"`)

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return options{}, err
		}

		return options{}, &flagError{err}
	}

	return options{
		mode: mode, directory: dir, size: size,
		extensions: splitList(ext), keywords: splitList(kwd),
		paths: splitList(paths), fuzzy: splitList(fuzzy),
		excludeDirs: optionalList(excludeDirs), excludeExts: optionalList(excludeExts),
		regex: rgx, showLines: showLines, showCols: showCols,
	}, nil
}

func splitList(value string) []string {
	return strings.Split(value, ",")
}

func optionalList(value string) []string {
	if value == "" {
		return nil
	}

	return splitList(value)
}

func (o options) validate() error {
	switch o.mode {
	case "scn":
		if (o.showLines || o.showCols) && o.keywords[0] == "" {
			return errors.New("Position tracking flags (-l, -c) require keywords (-k)")
		}
	case "fff":
		if o.keywords[0] == "" {
			return errors.New("-k (known hashes) REQUIRED for FileFingerprintFinder")
		}
	}

	return nil
}

// run keeps flag state, output, and errors local to each invocation.
func run(args []string, stdout, stderr io.Writer) error {
	o, err := parseOptions(args, stderr)

	if errors.Is(err, flag.ErrHelp) {
		return nil
	}

	if err != nil {
		return err
	}

	if err := o.validate(); err != nil {
		return err
	}

	var paths []string

	switch o.mode {
	case "scn":
		scanner := scnnr.Scanner{
			Directory: o.directory, FileExtensions: o.extensions, Keywords: o.keywords,
			Regex: o.regex, ShowLines: o.showLines, ShowCols: o.showCols,
			ExcludeDirs: o.excludeDirs, ExcludeExts: o.excludeExts,
		}

		return scanner.ScanTo(stdout)
	case "fnf":
		finder := scnnr.NewFileNameFinder(o.fuzzy)

		for _, path := range o.paths {
			finder.Scan(path)
		}

		paths = finder.Files
	case "fsf":
		finder := scnnr.NewFileSizeFinder(o.size)
		finder.Scan(o.directory)

		paths = finder.Files
	case "fff":
		finder := scnnr.NewFileFingerprintFinder(o.keywords)
		finder.Scan(o.directory)

		paths = finder.Files
	}

	for _, path := range paths {
		io.WriteString(stdout, path+"\n")
	}

	return nil
}
