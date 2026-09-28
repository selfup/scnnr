package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type buildTarget struct {
	goos, arch, output string
}

var targets = []buildTarget{
	{"darwin", "amd64", "scnnr_bins/mac/intel/scnnr"},
	{"darwin", "arm64", "scnnr_bins/mac/arm/scnnr"},
	{"linux", "amd64", "scnnr_bins/linux/intel/scnnr"},
	{"linux", "arm64", "scnnr_bins/linux/arm/scnnr"},
	{"windows", "386", "scnnr_bins/windows/scnnr.exe"},
}

func main() {
	if err := buildAll(".", os.Environ(), os.Stdout, os.Stderr, (*exec.Cmd).Run); err != nil {
		log.Fatal(err)
	}
}

// buildAll accepts command execution and environment as inputs so build plans
// and failures can be tested without compiling binaries or changing os.Environ.
func buildAll(root string, environment []string, stdout, stderr io.Writer, execute func(*exec.Cmd) error) error {
	for _, target := range targets {
		output := filepath.Join(root, filepath.FromSlash(target.output))
		if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
			return err
		}
		cmd := exec.Command("go", "build", "-o", filepath.FromSlash(target.output), "main.go")
		cmd.Dir = root
		cmd.Env = targetEnv(environment, target)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, stdout, stderr
		if err := execute(cmd); err != nil {
			return fmt.Errorf("build %s/%s: %w", target.goos, target.arch, err)
		}
		if target.goos != "windows" {
			if err := os.Chmod(output, 0755); err != nil {
				return err
			}
		}
	}
	if envValue(environment, "CI") != "" && envValue(environment, "VERSION") != "" {
		if err := os.WriteFile(filepath.Join(root, "scnnr_bins/version"), []byte(envValue(environment, "VERSION")+"\n"), 0644); err != nil {
			return err
		}
	}
	for _, name := range []string{"README.md", "LICENSE"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, "scnnr_bins", name), data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func targetEnv(environment []string, target buildTarget) []string {
	var result []string
	for _, entry := range environment {
		key, _, _ := strings.Cut(entry, "=")
		if key != "CGO_ENABLED" && key != "GOOS" && key != "GOARCH" {
			result = append(result, entry)
		}
	}
	return append(result, "CGO_ENABLED=0", "GOOS="+target.goos, "GOARCH="+target.arch)
}

func envValue(environment []string, name string) string {
	for _, e := range slices.Backward(environment) {
		key, value, _ := strings.Cut(e, "=")
		if key == name {
			return value
		}
	}
	return ""
}
