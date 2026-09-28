<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Scnnr](#scnnr)
    - [Help](#help)
    - [No Keywords](#no-keywords)
    - [Single Keyword](#single-keyword)
    - [Multiple Keywords and Multiple File Extensions](#multiple-keywords-and-multiple-file-extensions)
    - [Keywords with line numbers](#keywords-with-line-numbers)
    - [Keywords with both line numbers and column numbers](#keywords-with-both-line-numbers-and-column-numbers)
    - [Keywords while excluding dirs and file extensions](#keywords-while-excluding-dirs-and-file-extensions)
- [File Name Finder (NameFinder) (fnf)](#file-name-finder-namefinder-fnf)
- [File Fingerprint Finder (FingerprintFinder) (fff)](#file-fingerprint-finder-fingerprintfinder-fff)
- [File Size Finder (SizeFinder) (fsf)](#file-size-finder-sizefinder-fsf)
- [Back to Scnnr](#back-to-scnnr)
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
- [Performance (scn)](#performance-scn)
- [Tests](#tests)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Scnnr

Scans files (by extension) in a given directory for a keyword. Can be any file, or can be just `.js` or `.js,.html,.jsx`.

Prints out a `\n` delimited string of each file (filepath in artifact) containing one of the keywords.

Content matching uses concurrent batches of up to 1024 files. Match output follows concurrent processing order.

Has 3 additional modes: FileSizeFinder (find a file at or above a certain size), FileNameFinder (fuzzy find files with certain keywords in the filename), and FileFingerprintFinder (find files based on their SHA2-256 hash)

`scn` mode is the default

In scn mode, errors walking directories stop the scan. Errors opening or reading files selected for content scanning are fatal.

Finder modes panic if the root directory cannot be read. Files with unavailable metadata are skipped, and errors reading child directories are ignored. Failed directory metadata reads can still cause a panic.

### Help

Call scnnr with the `-h` flag:

```
$ scnnr -h
  -c    OPTIONAL Scnnr MODE
            show line and column numbers for each match (requires -k)
            when enabled, finds ALL matches in files (no early exit)
  -d string
        OPTIONAL Scnnr MODE and OPTIONAL FingerrintFinder MODE
            directory where scnnr will scan
            default is current directory and all child directories (default ".")
  -e string
        OPTIONAL Scnnr MODE
            a comma delimited list of file extensions to scan
            if none are given all files will be searched
  -f string
        REQUIRED NameFinder MODE
            fuzzy find the filename(s) contain(s) - can be comma delimited: Example 'wow' or 'wow,omg,lol'
  -k string
        OPTIONAL Scnnr MODE and REQUIRED FingerprintFinder
            scnnr: this is a comma delimited list of characters to look for in a file
                if no keywords are given - all file paths of given file extensions will be returned
                if keywords are given - only filepaths of matches will be returned
            FingerprintFinder: this is a comma delimited list of SHA2-256 hashes to find files by
  -l    OPTIONAL Scnnr MODE
            show line numbers for each match (requires -k)
            when enabled, finds ALL matches in files (no early exit)
  -m string
        OPTIONAL
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
    
         (default "scn")
  -p string
        REQUIRED NameFinder MODE
            any absolute path - can be comma delimited: Example: $HOME or '/tmp,/usr'
  -r    OPTIONAL Scnnr MODE
            if you want to use the regex engine or not
            defaults to false and will not use the regex engine for scans unless set to a truthy value
            truthy values are: 1, t, T, true, True, TRUE
            falsy values are: 0, f, F, false, False, FALSE
  -s string
        REQUIRED SizeFinder MODE
            size: 1MB,10MB,100MB,1GB,10GB,100GB,1TB
  -xd string
        OPTIONAL Scnnr MODE
            comma-delimited list of directory names to exclude from scanning
            ex: -xd ".git,node_modules,.venv"
  -xe string
        OPTIONAL Scnnr MODE
            comma-delimited list of file extensions to exclude from scanning
            ex: -xe ".log,.tmp,.json"
```

### No Keywords

If you just want to scan for file paths

```bash
$ scnnr -e .md -d .
README.md
```

You can also use equal signs for the flags:

```bash
$ scnnr -e=.md -d=.
README.md
```

Do not provide any keywords and scnnr will return all given filepaths matching given extensions.

If no extensions are given, all filepaths will be returned in the scanned directory. This will walk all dirs!

### Single Keyword

Scan this repo for markdown files with the keyword `cache=` in them.

_With quotes_

```bash
$ scnnr -e ".md" -d "." -k "cache="
README.md
```

_Without quotes (if no need to escape anything)_

```bash
scnnr -e .md -d . -k cache=
```

### Multiple Keywords and Multiple File Extensions

_using -k_

These examples use the files in `pkg/testdata/tree`. Matching paths can appear in a different order.

```bash
$ scnnr -d pkg/testdata/tree -e .txt,.go -k needle,token -xd ignored
pkg/testdata/tree/alpha.txt
pkg/testdata/tree/nested/beta.go
```

### Keywords with line numbers

_using -l_

Prints the first occurrence of each keyword on each matching line. With multiple keywords, the matching keyword is appended to each result.

```bash
$ scnnr -d pkg/testdata/tree -e .txt,.go -k needle,token -l -xd ignored
pkg/testdata/tree/alpha.txt:1:needle
pkg/testdata/tree/alpha.txt:2:token
pkg/testdata/tree/nested/beta.go:2:needle
```

### Keywords with both line numbers and column numbers

_using -c_

Line and column numbers start at 1. Columns count bytes.

```bash
$ scnnr -d pkg/testdata/tree -e .txt,.go -k needle,token -c -xd ignored
pkg/testdata/tree/alpha.txt:1:7:needle
pkg/testdata/tree/alpha.txt:2:8:token
pkg/testdata/tree/nested/beta.go:2:4:needle
```

### Keywords while excluding dirs and file extensions

_using -xd and -xe_

```bash
$ scnnr -d pkg/testdata/tree -k needle,token -c -xd ignored -xe .log
pkg/testdata/tree/alpha.txt:1:7:needle
pkg/testdata/tree/alpha.txt:2:8:token
pkg/testdata/tree/nested/beta.go:2:4:needle
```

# File Name Finder (NameFinder) (fnf)

```
  -p string
        REQUIRED NameFinder MODE
            any absolute path - can be comma delimited: Example: $HOME or '/tmp,/usr'
  -f string
        REQUIRED NameFinder MODE
            fuzzy find the filename(s) contain(s) - can be comma delimited: Example 'wow' or 'wow,omg,lol'
```

Example use to search `/tmp`, `/etc`, `/usr`, and `$HOME/Documents` for filenames that contain:

1. DEFCON
1. Tax
1. Return
1. Finance

```bash
scnnr -m fnf -f DEFCON,Finance,Tax,Return -p /tmp,/usr,/etc,$HOME/Documents
```

# File Fingerprint Finder (FingerprintFinder) (fff)

```
  -d string
        OPTIONAL Scnnr MODE and OPTIONAL FingerrintFinder MODE
            directory where scnnr will scan
            default is current directory and all child directories (default ".")
  -k string
        OPTIONAL Scnnr MODE and REQUIRED FingerprintFinder
            scnnr: this is a comma delimited list of characters to look for in a file
                if no keywords are given - all file paths of given file extensions will be returned
                if keywords are given - only filepaths of matches will be returned
            FingerprintFinder: this is a comma delimited list of SHA2-256 hashes to find files by
```

Example use to find a file with a known hash:

```bash
$ scnnr -m fff -d pkg/testdata/tree -k e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
pkg/testdata/tree/empty.txt
```

Please refer to the release notes for more details:

GitHub: https://github.com/selfup/scnnr/releases

Gitlab: https://gitlab.com/selfup/scnnr/-/releases

# File Size Finder (SizeFinder) (fsf)

```
  -s string
        REQUIRED SizeFinder MODE
            size: 1MB,10MB,100MB,1GB,10GB,100GB,1TB
```

Example use to find any file at least 100MB in E:/LotsOfStuff

```bash
scnnr -m fsf -s 100MB -d E:/LotsOfStuff
```

Sizes use decimal bytes: 1MB is 1,000,000 bytes. Files equal to the threshold are included.

# Back to Scnnr

### Using the package github.com/selfup/scnnr/pkg

```go
import (
  "log"

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

`Scan` prints to stdout and appends to the scanner's result fields. Reusing a scanner also reprocesses its accumulated candidate files. Use `[]string{""}` for no keywords or all extensions; nil or empty keyword and extension slices can panic.

With position tracking, each file parse records the first occurrence of each keyword per line. Without position tracking, each parse records at most one matching path and continues reading the file. Duplicate extensions and repeated scans can produce duplicate results.

The current batching code can omit files: a full final 1024-file batch is not parsed, and the file that flushes a full batch is skipped. Listing paths without keywords does not use this batching.

`Scan` returns directory traversal errors. Errors opening or reading a file call `log.Fatal`. Invalid regexes can panic when a line is examined. `ScanTo(writer)` directs the same output to a writer. Both methods ignore output write errors.

The working tree also includes `Search`, `ScanTo`, `WriteResults`, and `ParseSize`. These methods and helpers are absent from v1.3.0.

`Search` returns results without printing or changing the scanner's result fields. It removes empty strings from keyword and extension lists, returns read and regex errors, and orders results by file path. With `ShowLines` or `ShowCols` enabled, it requires a nonempty keyword and collects all non-overlapping occurrences of each keyword. Without those flags, it stops reading each file after its first match.

With the scanner above and a writer:

```go
scanner.ShowCols = true

result, err := scanner.Search()

if err != nil {
  log.Fatal(err)
}

err = scnnr.WriteResults(writer, result, scnnr.OutputOptions{
  ShowCols:     true,
  ShowKeywords: len(scanner.Keywords) > 1,
})

if err != nil {
  log.Fatal(err)
}
```

`WriteResults` returns output write errors.

All three finder methods use `Scan(directory string)` with no return value and append results to `Files`. NameFinder and SizeFinder construct paths as `directory + directory + filename`. FingerprintFinder uses `directory + Direction + filename`, matches digest substrings, and ignores file read errors.

`NewFileSizeFinder` panics on invalid sizes. `scnnr.ParseSize` validates the same thresholds and returns an error for an invalid size.

## Regex

### Using Regex Patterns

`scnnr -e ".js" -d "artifact" -k "cons?" -r=T > .results`

According to the godoc for `flag.BoolVar` you can use a few things for boolean flag values:

`t, T, 1, true, True, TRUE`

```bash
scnnr -r=1 -d artifact -e .js,.ts,.md -k 'cons*,let?,var?, impor*, expor*' > .results
```

### Using the package github.com/selfup/scnnr/pkg

```go
import (
  "log"

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

# Install

### If you have Go

GitHub repo:

```bash
go install github.com/selfup/scnnr@latest
```

GitLab repo:

```bash
go install gitlab.com/selfup/scnnr@latest
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

_the sha2-256 sum is provided in the artifact zip of the `scnnr_bins.zip`_

_you can also verify the provided sum matches the output in CI (output for verification)_

1. Unzip `artifacts.zip`
1. Unzip `scnnr_bins.zip`

From here pick your arch (mac/windows/linux) and appropriate binary and move to needed path! Mac and Linux builds have both intel and arm builds.

```
scnnr_bins/linux/intel:
scnnr

scnnr_bins/linux/arm:
scnnr

scnnr_bins/mac/intel:
scnnr

scnnr_bins/mac/arm:
scnnr

scnnr_bins/windows:
scnnr.exe
```

### Docker

1. Clone repo: `git clone https://github.com/selfup/scnnr`
1. `cd` into repo

   - Shell: `./scripts/dind.build.sh`
   - Powershell: `./scripts/dind.build.ps1`

1. Unzip `scnnr_bins.zip`

From here pick your arch (mac/windows/linux) and appropriate binary and move to needed path!

```
scnnr_bins/linux/intel:
scnnr

scnnr_bins/linux/arm:
scnnr

scnnr_bins/mac/intel:
scnnr

scnnr_bins/mac/arm:
scnnr

scnnr_bins/windows:
scnnr.exe
```

## Performance (scn)

Use of goroutines, buffers, streams, mutexes, and simple checks.

The following are historical measurements from a JavaScript project. They have not been rerun for the current working tree. The original memory measurement reported usage below 5.5MB.

To measure the current build, compile once and time the binary using your own directory:

```bash
go build -o /tmp/scnnr-bench main.go
time /tmp/scnnr-bench -d path/to/files -e .js,.md -k cache
```

Timing `go run` also includes build and launch overhead. Timings depend on the files, filesystem cache, and machine used.

No matches on 33k files after `npm i` for a JavaScript project as the `artifact`:

```
$ time scnnr -d artifact -e .kt -k cache

real    0m0.121s
user    0m0.053s
sys     0m0.076s
```

33k files, two file types, one keyword, and 567 matches:

```
$ time scnnr -d artifact -e .md,.js -k cache > .results

real    0m0.232s
user    0m0.574s
sys     0m0.210s
$ ls -lahg .results
-rw-r--r-- 1 selfup 33K Jul 21 00:55 .results
```

33k files, two file types, 5 keywords, and 360 matches:

```
$ time scnnr -d artifact -e .js,.md -k stuff,things,wow,lol,omg > .results

real    0m0.266s
user    0m0.813s
sys     0m0.174s
$ ls -lahg .results
-rw-r--r-- 1 selfup 22K Jul 21 00:53 .results
```

33k files, 4 file types, 5 common keywords, and 18866 matches:

Results are piped into a file to reduce noise.

The amount of file paths results in 1.2MB of text data..

```
$ time scnnr -d artifact -e .js,.ts,.md,.css -k const,let,var,import,export > .results

real    0m0.344s
user    0m0.755s
sys     0m0.351s
$ ls -lahg .results
-rw-r--r-- 1 selfup 1.2M Jul 21 00:57 .results
```

## Tests

```bash
go test -race ./...
go vet ./...
./scripts/e2e.sh
```

Tests cover reader and writer failures, build execution, filesystem fixtures, matching, accumulated results, and CLI exit statuses. Compile-time checks cover the existing finder and scanner signatures and selected helpers.

The E2E script creates its own temporary tree and checks output for all four modes. Set `ETE_DIR` to choose the parent directory for those fixtures.

`go vet` checks for suspicious code patterns. It does not detect unused function or method declarations.
