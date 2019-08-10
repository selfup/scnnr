<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Scannr](#scannr)
  - [Examples](#examples)
    - [Help](#help)
    - [No Keywords](#no-keywords)
    - [Single Keyword](#single-keyword)
    - [Multiple Keywords and Multiple File Extensions](#multiple-keywords-and-multiple-file-extensions)
    - [Using the package github.com/selfup/scnnr/pkg](#using-the-package-githubcomselfupscnnrpkg)
  - [Regex](#regex)
    - [Using Regex Patterns](#using-regex-patterns)
    - [Using the package github.com/selfup/scnnr/pkg](#using-the-package-githubcomselfupscnnrpkg-1)
  - [Install](#install)
    - [If you have Go](#if-you-have-go)
    - [If you do not have Go](#if-you-do-not-have-go)
      - [Release Binaries](#release-binaries)
      - [Direct Download Link](#direct-download-link)
      - [cURL](#curl)
      - [wget](#wget)
      - [Docker](#docker)
  - [Performance](#performance)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Scannr

Scans files (by extension) in a given directory for a keyword. Can be any file, or can be just `.js` or `.js,.html,.jsx`.

Prints out a `\n` delimited string of each file (filepath in artifact) containing one of the keywords.

Max file descriptors is set to 1024 (linux default).

## Examples

### Help

Call scnnr with the `-h` flag:

```
$ scnnr -h
  -d string
        OPTIONAL
            directory where scnnr will scan
            default is current directory and all child directories (default ".")
  -e string
        OPTIONAL
            a comma delimted list of file extensions to scan
            if none are given all files will be searched
  -k string
        OPTIONAL
            a comma delimted list of characters to look for in a file
            if no keywords are given - all file paths of given file extensions will be returned
            if keywords are given only filepaths of matches will be returned
  -r    OPTIONAL
            wether to use the regex engine or not
            defaults to false and will not use the regex engine for scans unless set to a truthy value
            truthy values are: 1, t, T, true, True, TRUE
            flasey values are: 0, f, F, false, False, FALSE
```

### No Keywords

If you just want to scan for file paths

```bash
$ scnnr -e=.md -d=.
README.md
```

Do not provide any keywords and scnnr will return all given filepaths matching given extensions.

If no extensions are given, all filepaths will be returned in the scanned directory.

### Single Keyword

Scan this repo for markdown files with the keyword `cache=` in them.

_With quotes_

```bash
$ scnnr -e=".md" -d="." -k="cache="
README.md
```

_Without quotes (if no need to escape anything)_

```bash
scnnr -e=.md -d=. -k=cache=
```

### Multiple Keywords and Multiple File Extensions

```bash
$ scnnr -d=. -e=.md,.go -k=fileData,cache
README.md
cmd/scanner.go
```

### Using the package github.com/selfup/scnnr/pkg

```go
import (
  scnnr "github.com/selfup/scnnr/pkg"
)

directory := "./artifact"
keywords := []string{"something","something else", "another thing"}
patterns := []string{".js",".go",".md"}

scanner := scnnr.Scanner{
  Directory:      directory,
  FileExtensions: patterns,
  Keywords:       keywords,
}

err := scanner.Scan()

if err != nil {
  log.Fatal(err)
}
```

## Regex

### Using Regex Patterns

`scnnr -e=".js" -d="artifact" -k="cons?" -r=T > .results`

According to the godoc for `flag.BoolVar` you can use a few things for boolean flag values:

`t, T, 1, true, True, TRUE`

```
scnnr $ time scnnr -r=1 -d=artifact -e=.js,.ts,.md -k='cons*,let?,var?, impor*, expor*' > .results

real    0m0.748s
user    0m2.398s
sys     0m0.311s
```

### Using the package github.com/selfup/scnnr/pkg

```go
import (
  scnnr "github.com/selfup/scnnr/pkg"
)

rgx := true
directory := "./artifact"
keywords := []string{"const PASSW*","password?", "export PASS?"}
extensions := []string{".js",".ts"}

scanner := scnnr.Scanner{
  Regex:          rgx,
  Directory:      directory,
  FileExtensions: extensions,
  Keywords:       keywords,
}

err := scanner.Scan()

if err != nil {
  log.Fatal(err)
}
```

## Install

### If you have Go

```bash
go get -u github.com/selfup/scnnr
go install github.com/selfup/scnnr
```

### If you do not have Go

#### Release Binaries

I have a [GitLab Release Repo](https://gitlab.com/selfup/scnnr) that builds the needed artifacts using [GitLabCI](https://docs.gitlab.com/ee/ci/quick_start/)

#### Direct Download Link

https://gitlab.com/selfup/scnnr/-/jobs/artifacts/master/download?job=release

#### cURL

```bash
curl -L https://gitlab.com/selfup/scnnr/-/jobs/artifacts/master/download?job=release > artifacts.zip
```

#### wget

```bash
wget https://gitlab.com/selfup/scnnr/-/jobs/artifacts/master/download?job=release -O artifacts.zip
```

1. Unzip `artifacts.zip`
1. Unzip `scnnr_bins.zip`

From here pick your arch (mac/windows/linux) and appropriate binary and move to needed path!

```
scnnr_bins/linux:
scnnr

scnnr_bins/mac:
scnnr

scnnr_bins/windows:
scnnr.exe
```

#### Docker

1. Clone repo: `git clone https://github.com/selfup/scnnr`
1. `cd` into repo

   - Shell: `./scripts/dind.build.sh`
   - Powershell: `./scripts/dind.build.ps1`

1. Unzip `scnnr_bins.zip`

From here pick your arch (mac/windows/linux) and appropriate binary and move to needed path!

```
scnnr_bins/linux:
scnnr

scnnr_bins/mac:
scnnr

scnnr_bins/windows:
scnnr.exe
```

## Performance

Use of goroutines, buffers, streams, mutexes, and simple checks.

Memory in the following example never went above 5.5MB for the entire program.

No matches on 33k files after `npm i` for a JavaScript project as the `artifact`:

```
$ time scnnr -d=artifact -e=.kt -k=cache

real    0m0.289s
user    0m0.241s
sys     0m0.138s
```

33k files, two file types, one keyword, and 567 matches. _Not all 567 matches displayed in README_:

```
$ time scnnr -d=artifact -e=.md,.js -k=cache > .results

real    0m0.435s
user    0m0.843s
sys     0m0.258s
$ ls -lahg .results
-rw-r--r-- 1 selfup 33K Jul 21 00:55 .results
```

33k files, two file types, 5 keywords, and 360 matches. _Not all 360 matches displayed in README_:

```
$ time scnnr -d=artifact -e=.js,.md -k=stuff,things,wow,lol,omg > .results

real    0m0.450s
user    0m1.016s
sys     0m0.263s
$ ls -lahg .results
-rw-r--r-- 1 selfup 22K Jul 21 00:53 .results
```

33k files, 4 file types, 5 common keywords, and 18866 matches. _Not all 18866 matches displayed in README_:

Results are piped into a file to reduce noise.

The amount of file paths results in 1.2MB of text data..

```
$ time scnnr -d=artifact -e=.js,.ts,.md,.css -k=const,let,var,import,export > .results

real    0m0.445s
user    0m0.924s
sys     0m0.304s
$ ls -lahg .results
-rw-r--r-- 1 selfup 1.2M Jul 21 00:57 .results
```
