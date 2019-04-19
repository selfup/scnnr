# Scannr

```bash
go run main.go ".js,.html" "initial-scale=1, cache= "
```

```
$ go run main.go ".js,.html" "initial-scale=1, cache= "
[.js .html] [initial-scale=1  cache= ]
2019/04/18 20:26:20 [artifact\hyperapp-0\src\index.html]
```

```bash
export PATTERNS=$(cat patterns.csv) \
  && export KEYWORDS=$(cat keywords.csv) \
  && go run main.go "$PATTERNS" "$KEYWORDS"
```

```
$ export PATTERNS=$(cat patterns.csv) && export KEYWORDS=$(cat keywords.csv) && go run main.go "$PATTERNS" "$KEYWORDS"
[.html .yml .js] [initial-scale=1  Cache:   cache:   Cache=   cache= ]
2019/04/18 20:26:53 [artifact\hyperapp-0\src\index.html]
```
