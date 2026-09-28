package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func cliFixture(parts ...string) string {
	return filepath.Join(append([]string{"pkg", "testdata", "tree"}, parts...)...)
}

func assertOutputLines(t *testing.T, got, want string) {
	t.Helper()
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")
	slices.Sort(gotLines)
	slices.Sort(wantLines)
	if !slices.Equal(gotLines, wantLines) {
		t.Fatalf("output=%q; want the same original lines as %q", got, want)
	}
}

func TestParseOptionsKeepsOriginalDefaultsAndCSV(t *testing.T) {
	defaults, err := parseOptions(nil, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if defaults.mode != "scn" || defaults.directory != "." ||
		!reflect.DeepEqual(defaults.keywords, []string{""}) ||
		!reflect.DeepEqual(defaults.extensions, []string{""}) ||
		defaults.excludeDirs != nil || defaults.excludeExts != nil {
		t.Fatalf("original defaults changed: %#v", defaults)
	}
	opts, err := parseOptions([]string{
		"-d", "sample", "-e", ".go,.txt", "-k", "needle,, token",
		"-r", "-c", "-xd", ".git,node_modules", "-xe", ".log",
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.regex || !opts.showCols ||
		!reflect.DeepEqual(opts.keywords, []string{"needle", "", " token"}) ||
		!reflect.DeepEqual(opts.extensions, []string{".go", ".txt"}) {
		t.Fatalf("CSV or flags changed: %#v", opts)
	}
	ignored, err := parseOptions([]string{"positional", "-m", "fff"}, io.Discard)
	if err != nil || !reflect.DeepEqual(ignored, defaults) {
		t.Fatalf("original positional-argument behavior changed: %#v, %v", ignored, err)
	}
	if _, err := parseOptions([]string{"-h"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help error=%v", err)
	}
	again, err := parseOptions(nil, io.Discard)
	if err != nil || !reflect.DeepEqual(defaults, again) {
		t.Fatal("flag state leaked between invocations")
	}
}

func TestRunOriginalScannerOutput(t *testing.T) {
	root := cliFixture()
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"list paths", []string{"-d", root, "-e", ".go,.log"}, cliFixture("nested", "beta.go") + "\n" + cliFixture("skip.log") + "\n"},
		{"matching paths", []string{"-d", root, "-k", "needle", "-xd", "ignored", "-xe", ".log"}, cliFixture("alpha.txt") + "\n" + cliFixture("nested", "beta.go") + "\n"},
		{"one keyword occurrence per line", []string{"-d", root, "-e", ".txt", "-k", "needle,token", "-l", "-xd", "ignored"}, cliFixture("alpha.txt") + ":1:needle\n" + cliFixture("alpha.txt") + ":2:token\n"},
		{"columns", []string{"-d", root, "-e", ".txt", "-k", "needle", "-c", "-xd", "ignored"}, cliFixture("alpha.txt") + ":1:7\n"},
		{"regex", []string{"-d", root, "-e", ".go", "-k", "n.edle", "-r=true", "-c"}, cliFixture("nested", "beta.go") + ":2:4\n"},
		{"no match retains newline", []string{"-d", root, "-k", "absent"}, "\n"},
		{"empty first CSV element lists files", []string{"-d", root, "-e", ".go", "-k", ",needle"}, cliFixture("nested", "beta.go") + "\n"},
		{"unknown mode succeeds silently", []string{"-m", "unknown"}, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := run(test.args, &stdout, &stderr); err != nil {
				t.Fatal(err)
			}
			assertOutputLines(t, stdout.String(), test.want)
			if stderr.Len() != 0 {
				t.Fatalf("stderr=%q", stderr.String())
			}
		})
	}
}

func TestRunOriginalFinders(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "needle.txt"), []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	second := t.TempDir()
	if err := os.WriteFile(filepath.Join(second, "needle.go"), []byte("other"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"names include duplicate keyword matches", []string{"-m", "fnf", "-p", root + "," + second, "-f", "needle,.txt"}, root + root + "needle.txt\n" + root + root + "needle.txt\n" + second + second + "needle.go\n"},
		{"missing fuzzy keyword retains match-all", []string{"-m", "fnf", "-p", root}, root + root + "needle.txt\n"},
		{"full hash", []string{"-m", "fff", "-d", root, "-k", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}, filepath.Join(root, "needle.txt") + "\n"},
		{"hash substring", []string{"-m", "fff", "-d", root, "-k", "ba7816"}, filepath.Join(root, "needle.txt") + "\n"},
		{"hash no match", []string{"-m", "fff", "-d", root, "-k", "unknown"}, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			if err := run(test.args, &stdout, io.Discard); err != nil {
				t.Fatal(err)
			}
			assertOutputLines(t, stdout.String(), test.want)
		})
	}
	sizeRoot := t.TempDir()
	for name, size := range map[string]int64{"below": 999999, "equal": 1000000, "above": 1000001} {
		file, err := os.Create(filepath.Join(sizeRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		err = file.Truncate(size)
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			t.Fatalf("size fixture: %v, %v", err, closeErr)
		}
	}
	var stdout bytes.Buffer
	if err := run([]string{"-m", "fsf", "-d", sizeRoot, "-s", "1MB"}, &stdout, io.Discard); err != nil {
		t.Fatal(err)
	}
	assertOutputLines(t, stdout.String(), sizeRoot+sizeRoot+"above\n"+sizeRoot+sizeRoot+"equal\n")
}

func TestRunOriginalFailures(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"unknown flag", []string{"-unknown"}, "flag provided but not defined"},
		{"invalid bool", []string{"-r=invalid"}, "invalid"},
		{"lines without keywords", []string{"-l"}, "Position tracking flags (-l, -c) require keywords (-k)"},
		{"columns without keywords", []string{"-c", "-k", ",needle"}, "Position tracking flags (-l, -c) require keywords (-k)"},
		{"fingerprints missing", []string{"-m", "fff"}, "-k (known hashes) REQUIRED for FileFingerprintFinder"},
		{"missing scanner directory", []string{"-d", filepath.Join(t.TempDir(), "missing")}, "missing"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(test.args, &stdout, &stderr)
			if err == nil || !strings.Contains(err.Error(), test.want) || stdout.Len() != 0 {
				t.Fatalf("error=%v stdout=%q; want %q", err, stdout.String(), test.want)
			}
		})
	}
	wantErr := errors.New("stdout failed")
	if err := run([]string{"-d", cliFixture(), "-e", ".go"}, failingWriter{wantErr}, io.Discard); err != nil {
		t.Fatalf("original output errors were ignored; got %v", err)
	}
}

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-h"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "Usage of "+os.Args[0]) {
		t.Fatalf("help stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIOriginalExitStatuses(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "scnnr")
	cmd := exec.Command("go", "build", "-o", binary, "main.go")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	for _, test := range []struct {
		name    string
		args    []string
		code    int
		message string
	}{
		{"help", []string{"-h"}, 0, "Usage of " + binary},
		{"unknown mode", []string{"-m", "unknown"}, 0, ""},
		{"bad flag", []string{"-unknown"}, 2, "flag provided but not defined"},
		{"invalid size panics", []string{"-m", "fsf", "-s", "2MB"}, 2, "panic: please provide a size"},
		{"name root panics", []string{"-m", "fnf", "-f", "needle", "-p", filepath.Join(t.TempDir(), "missing")}, 2, "panic:"},
		{"scanner root returns error", []string{"-d", filepath.Join(t.TempDir(), "missing")}, 1, "missing"},
		{"invalid regex panics", []string{"-d", cliFixture(), "-e", ".go", "-k", "[", "-r"}, 2, "panic: regexp:"},
		{"required positions", []string{"-c"}, 1, "Position tracking flags"},
	} {
		t.Run(test.name, func(t *testing.T) {
			output, err := exec.Command(binary, test.args...).CombinedOutput()

			code := 0

			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatal(err)
				}
				code = exitErr.ExitCode()
			}

			if code != test.code || !strings.Contains(string(output), test.message) {
				t.Fatalf("exit=%d output=%q; want exit=%d containing %q", code, output, test.code, test.message)
			}
		})
	}
}
