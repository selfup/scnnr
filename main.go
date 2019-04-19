package main

import "github.com/selfup/scnnr/cmd"

func main() {
	scanner := new(cmd.Scanner)

	scanner.Patterns = []string{".html", ".yml", ".js"}
	scanner.Keywords = []string{"initial-scale=1", "Cache", "cache"}

	scanner.Scan()
}
