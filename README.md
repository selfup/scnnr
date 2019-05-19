# Scannr

Scans all files in a given directory for a keyword. Can be any file, or can be just `.js` or `".js,.html,.jsx"`.

Prints out a comma sperated list of each file containing one of the keywords.

This is extremely helpful for dealing with multiple large files.

Can be used in a serverless lambda/function to scan hundreds of artifacts at once (one artifact per lambda).

Max file descriptors is set to 1024 (linux default).

## Example

#### Single Keyword

Scan this repo for markdown files with `cache=` in them.

```bash
$ go run main.go ./ .md cache=
README.md
```

#### Multiple Keywords and Multiple File Extensions

```bash
scnnr (master) $ go run main.go ./ .md,.go fileData,cache
README.md,cmd/scanner.go,main.go
```

#### Using the package github.com/selfup/scnnr/scnnr

```go
import "github.com/selfup/scnnr/scnnr"

directory := "./artifact"
keywords := []string{"something","something else", "another thing"}
patterns := []string{".js",".go",".md"}

scanner := scnnr.Scanner{
  Directory:    directory,
  FilePatterns: patterns,
  Keywords:     keywords,
}

err := scanner.Scan()

if err != nil {
  log.Fatal(err)
}
```

## Regex

`SCNNR_REGEX=1 go run main.go artifact/ .md,.js cach? > .results`

If you add the `SCNNR_REGEX=1` ENV variable you can then use regex statements instead of raw keywords.

```
scnnr (regex_support) $ time SCNNR_REGEX=1 go run main.go artifact/ .js,.ts,.md 'cons*,let?,var?, impor*, expor*' > .results

real    0m0.748s
user    0m2.398s
sys     0m0.311s
```

## Performance

Use of goroutines, buffers, streams, mutexes, and simple checks.

Memory in the following example never went above 5.5MB for the entire program.

No matches on 33k files after `npm i`:

```
scnnr (master) $ time go run main.go artifact/ .kt cache
found 0 matches:
real    0m0.287s
user    0m0.235s
sys     0m0.131s
```

33k files, two file types, one keyword, and 567 matches. _Not all 517 matches displayed in README_:

```
scnnr (master) $ time go run main.go artifact/ .md,.js cache > .results

real    0m0.435s
user    0m0.909s
sys     0m0.287s
scnnr (master) $ ls -lahg .results
-rw-r--r-- 1 selfup 30K May 18 08:33 .results
```

33k files, two file types, 5 keywords, and 360 matches. _Not all 313 matches displayed in README_:

```
scnnr (master) $ time go run main.go artifact/ .js,.md stuff,things,wow,lol,omg > .results

real    0m0.516s
user    0m1.085s
sys     0m0.358s
scnnr (master) $ ls -lahg .results
-rw-r--r-- 1 selfup 20K May 18 08:33 .results
```

33k files, 4 file types, 5 common keywords, and 18866 matches. _Not all 18506 matches displayed in README_:

Results are piped into a file to reduce noise.

The amount of file paths results in 1.2MB of text data..

```
scnnr (master) $ time go run main.go artifact/ .js,.ts,.md,.css const,let,var,import,export > .results

real    0m0.504s
user    0m1.008s
sys     0m0.308s
scnnr (master) $ ls -lahg .results
-rw-r--r-- 1 selfup 1.2M May 17 17:22 .results
```
