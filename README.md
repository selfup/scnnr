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
keywords := "something,something else,another thing"
filepatterns := ".js,.go,.md"

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

33k files, two file types, one keyword, and 517 matches. _Not all 517 matches displayed in README_:

```
scnnr (master) $ time go run main.go artifact/ .md,.js cache
found 517 matches: artifact/node_modules/@babel/core/node_modules/@babel/traverse/lib/path/index.js,artifact/dist/src.7fec4c36.js,artifact/node_modules/@babel/core/node_modules/@babel/traverse/lib/index.js,artifact/node_modules/symbol-tree/lib/SymbolTree.js,artifact/node_modules/symbol-tree/lib/SymbolTreeNode.js
real    0m0.532s
user    0m1.430s
sys     0m0.279s
```

33k files, two file types, 5 keywords, and 313 matches. _Not all 313 matches displayed in README_:

```
scnnr (master) $ time go run main.go artifact/ .js,.md stuff,things,wow,lol,omg
found 313 matches: artifact/README.md,artifact/node_modules/@babel/core/node_modules/@babel/parser/CHANGELOG.md,artifact/node_modules/@babel/helper-call-delegate/node_modules/@babel/parser/CHANGELOG.md,artifact/node_modules/@babel/core/node_modules/json5/CHANGELOG.md,artifact/node_modules/@babel/generator/node_modules/jsesc/README.md,artifact/node_modules/@babel/helper-replace-supers/node_modules/@babel/parser/CHANGELOG.md,artifact/node_modules/@babel/helpers/node_modules/@babel/parser/CHANGELOG.md,artifact/node_modules/static-module/node_modules/source-map/dist/source-map.debug.js
real    0m0.734s
user    0m2.814s
sys     0m0.264s
```
