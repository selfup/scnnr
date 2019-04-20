# Scannr

Scans all files in a given directory for a keyword. Can be any file, or can be just `.js` or `".js,.html,.jsx"`.

Prints out a comma sperated list of each file containing one of the keywords.

## Example

#### Single Keyword

Scan this repo for markdown files with `cache=` in them.

```bash
$ go run main.go ./ .md cache=
README.md
```

#### Multiple Keywords and Multiple File Extensions

scnnr (master) $ go run main.go ./ .md,.go fileData,cache
README.md,cmd/scanner.go,main.go
