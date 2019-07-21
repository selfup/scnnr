# Scannr

Scans all files in a given directory for a keyword. Can be any file, or can be just `.js` or `".js,.html,.jsx"`.

Prints out a `\n` delimited string of each file (filepath in artifact) containing one of the keywords.

Very easy to parse!

This is extremely helpful for dealing with multiple large files.

Can be used in a serverless lambda/function to scan hundreds of artifacts at once (one artifact per lambda).

Max file descriptors is set to 1024 (linux default).

## Example

#### Help

Either use the `-h` flag or no flags at all to get help info.

Without the `-h` flag and/or all required args, `scnnr` will exit with a status code of 1:

```
$ scnnr
  -d string
        REQUIRED
            directory where scnnr will scan
  -e string
        REQUIRED
            a comma delimted list of file extensions to scan
  -k string
        REQUIRED
            a comma delimted list of keywords to search for in a file
  -r    OPTIONAL
            wether to use the regex engine or not
            defaults to false and will not use the regex engine for scans unless set to a truthy value
            truthy values are: 1, t, T, true, True, TRUE
            flasey values are: 0, f, F, false, False, FALSE

ERROR - scannr has required arguments - please read above output - exiting..
```

#### Single Keyword

Scan this repo for markdown files with `cache=` in them.

```bash
$ go run main.go -e=".md" -d="." -k="cache="
README.md
```

Or without quotes (if no need to escape anything)

```bash
go run main.go -e=.md -d=. -k=cache=
```

#### Multiple Keywords and Multiple File Extensions

```bash
$ go run main.go -d=. -e=.md,.go -k=fileData,cache
README.md
cmd/scanner.go
main.go
```

#### Using the package github.com/selfup/scnnr/scnnr

```go
import "github.com/selfup/scnnr/scnnr"

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

`go run main.go -e=".js" -d="artifact" -k="cons?" -r=T > .results`

According to the godoc for `flag.BoolVar` you can use a few things for boolean flag values:

`t, T, 1, true, True, TRUE`

```
scnnr $ time go run main.go -r=1 -d=artifact -e=.js,.ts,.md -k='cons*,let?,var?, impor*, expor*' > .results

real    0m0.748s
user    0m2.398s
sys     0m0.311s
```

#### Using the package github.com/selfup/scnnr/scnnr

```go
import "github.com/selfup/scnnr/scnnr"

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

## Performance

Use of goroutines, buffers, streams, mutexes, and simple checks.

Memory in the following example never went above 5.5MB for the entire program.

No matches on 33k files after `npm i` for a JavaScript project as the `artifact`:

```
$ time go run main.go -d=artifact -e=.kt -k=cache


real    0m0.289s
user    0m0.241s
sys     0m0.138s
```

33k files, two file types, one keyword, and 567 matches. _Not all 567 matches displayed in README_:

```
$ time go run main.go -d=artifact -e=.md,.js -k=cache > .results

real    0m0.435s
user    0m0.843s
sys     0m0.258s
$ ls -lahg .results
-rw-r--r-- 1 selfup 33K Jul 21 00:55 .results
```

33k files, two file types, 5 keywords, and 360 matches. _Not all 360 matches displayed in README_:

```
$ time go run main.go -d=artifact -e=.js,.md -k=stuff,things,wow,lol,omg > .results

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
$ time go run main.go -d=artifact -e=.js,.ts,.md,.css -k=const,let,var,import,export > .results

real    0m0.445s
user    0m0.924s
sys     0m0.304s
$ ls -lahg .results
-rw-r--r-- 1 selfup 1.2M Jul 21 00:57 .results
```
