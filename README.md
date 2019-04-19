# Scannr

Scan all files in a given directory for a keyword. Can be any file, or can be just `.js` or `".js,.html,.jsx"`.

It then searches for the keyword in each file and returns a slice of all files with said keyword.

### Example

Scan this repo for markdown files with `cache=` in them.

```bash
$ go run main.go ./ .md cache=
File patterns:  [.md]
Keywords:  [cache=]
Files matching pattern(s):  [README.md]
```
