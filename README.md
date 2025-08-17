<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Scnnr](#scnnr)
    - [Help](#help)
    - [No Keywords](#no-keywords)
    - [Single Keyword](#single-keyword)
    - [Multiple Keywords and Multiple File Extensions](#multiple-keywords-and-multiple-file-extensions)
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

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Scnnr

Scans files (by extension) in a given directory for a keyword. Can be any file, or can be just `.js` or `.js,.html,.jsx`.

Prints out a `\n` delimited string of each file (filepath in artifact) containing one of the keywords.

Max file descriptors is set to 1024 (linux default) in scnnr scn mode.

Has 3 additional modes: FileSizeFinder (find a file above a certain size), FileNameFinder (fuzzy find files with certain keywords in the filename), and FileFingerprintFinder (find files based on their SHA2-256 hash)

`scn` mode is the default

_Caveat: will not throw an error if a file cannot be read due to permissions. It is assumed that you know what files you can read/have access to. This allows for a clean/parseable output_

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

```bash
$ scnnr -d . -e .md,.go -k fileData,cache
README.md
cmd/scanner.go
```

### Keywords with line numbers

```bash
$ scnnr -d . -e .md,.go -k fileData,cache -l
scnnr_bins/README.md:117:cache
scnnr_bins/README.md:122:cache
scnnr_bins/README.md:129:cache
scnnr_bins/README.md:135:fileData
scnnr_bins/README.md:135:cache
README.md:123:cache
README.md:128:cache
README.md:135:cache
README.md:141:fileData
README.md:141:cache
scnnr_bins/README.md:377:cache
scnnr_bins/README.md:387:cache
README.md:383:cache
README.md:393:cache
pkg/scanner.go:215:fileData
pkg/scanner.go:222:fileData
```

### Keywords with both line numbers and column numbers

```bash
$ scnnr -d . -e .md,.go -k fileData,cache -c
scnnr_bins/README.md:117:53:cache
scnnr_bins/README.md:122:29:cache
scnnr_bins/README.md:129:22:cache
scnnr_bins/README.md:135:28:fileData
scnnr_bins/README.md:135:37:cache
scnnr_bins/README.md:377:36:cache
scnnr_bins/README.md:387:40:cache
pkg/scanner.go:215:9:fileData
pkg/scanner.go:222:26:fileData
README.md:123:53:cache
README.md:128:29:cache
README.md:135:22:cache
README.md:141:28:fileData
README.md:141:37:cache
README.md:383:36:cache
README.md:393:40:cache
```

### Keywords while excluding dirs (-xd) and file extensions (-xe)

```bash
go run main.go -k main,let,for -p . -c -xd ".git,scnnr_bins,pkg,cmd,.zip" -xe ".yml,.sh,.md" 
.editorconfig:11:21:let
.gitignore:3:1:main
.gitignore:4:1:main
.dockerignore:3:1:main
.dockerignore:4:1:main
Dockerfile:14:21:main
Dockerfile:15:25:main
main.go:25:9:main
main.go:36:6:main
main.go:46:8:for
main.go:47:8:for
main.go:48:8:for
main.go:49:8:for
main.go:70:65:for
main.go:78:57:for
main.go:84:23:for
main.go:89:34:for
main.go:125:3:for
main.go:129:3:for
main.go:172:3:for
main.go:179:42:for
main.go:186:3:for
scnnr_bins.zip:2053:93:for
scnnr_bins.zip:26325:8:let
scnnr_bins.zip:32123:53:let
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
$ known_hash="de4f51f97fa690026e225798ff294cd182b93847aaa46fe1e32b848eb9e985bd"
$ go run main.go -m fff -d $HOME/Documents -k $known_hash
/home/selfup/Documents//dotfiles/mac/.bash_profile
```

Please refer to the release notes for more details:

GitHub: https://github.com/selfup/scnnr/releases/tag/v1.1.8

Gitlab: https://gitlab.com/selfup/scnnr/-/releases/v1.1.8

# File Size Finder (SizeFinder) (fsf)

```
  -s string
        REQUIRED SizeFinder MODE
            size: 1MB,10MB,100MB,1GB,10GB,100GB,1TB
```

Example use to find any file over 100MB in E:/LotsOfStuff

```bash
scnnr -m fsf -s 100MB -d E:/LotsOfStuff
```

# Back to Scnnr

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

`scnnr -e ".js" -d "artifact" -k "cons?" -r T > .results`

According to the godoc for `flag.BoolVar` you can use a few things for boolean flag values:

`t, T, 1, true, True, TRUE`

```
scnnr $ time scnnr -r 1 -d artifact -e .js,.ts,.md -k 'cons*,let?,var?, impor*, expor*' > .results

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

Memory in the following example never went above 5.5MB for the entire program.

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
