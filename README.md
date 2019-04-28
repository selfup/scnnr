# Scannr

Scans all files in a given directory for a keyword. Can be any file, or can be just `.js` or `".js,.html,.jsx"`.

Prints out a comma sperated list of each file containing one of the keywords.

This is extremely helpful for dealing with multiple large files.

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

33k file and 517 matches. _Not all 517 matches displayed in README_:

```
scnnr (master) $ time go run main.go artifact/ .md,.js cache
found 517 matches: artifact/node_modules/@babel/core/node_modules/@babel/traverse/lib/path/index.js,artifact/dist/src.7fec4c36.js,artifact/node_modules/@babel/core/node_modules/@babel/traverse/lib/index.js,artifact/node_modules/symbol-tree/lib/SymbolTree.js,artifact/node_modules/symbol-tree/lib/SymbolTreeNode.js
real    0m0.532s
user    0m1.430s
sys     0m0.279s
```
