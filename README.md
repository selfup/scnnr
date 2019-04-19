# Scannr

```bash
go run main.go ".js,.html" "initial-scale=1, cache= "
```

```bash
$ go run main.go ".js,.html" "initial-scale=1, cache= "
[.js .html] [initial-scale=1  cache= ]
2019/04/18 20:26:20 [artifact\hyperapp-0\src\index.html]
```

```bash
go run main.go "$(cat patterns.csv)" "$(cat keywords.csv)"
```

```bash
$ go run main.go "$(cat patterns.csv)" "$(cat keywords.csv)"
[.html .yml .js] [initial-scale=1  Cache:   cache:   Cache=   cache= ]
2019/04/18 20:26:53 [artifact\hyperapp-0\src\index.html]
```
