package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func buildFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range map[string]string{"README.md": "readme\n", "LICENSE": "license\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func fakeBuild(cmd *exec.Cmd) error {
	return os.WriteFile(filepath.Join(cmd.Dir, cmd.Args[3]), []byte("binary"), 0600)
}

func TestBuildPlanEnvironmentAndArtifacts(t *testing.T) {
	root := buildFixture(t)
	environment := []string{"CGO_ENABLED=1", "GOOS=old", "GOARCH=old", "KEEP=value", "CI=true", "VERSION=42"}
	before := append([]string(nil), environment...)
	processEnv := []string{os.Getenv("GOOS"), os.Getenv("GOARCH"), os.Getenv("CGO_ENABLED")}
	want := []struct{ goos, arch, output string }{
		{"darwin", "amd64", "scnnr_bins/mac/intel/scnnr"},
		{"darwin", "arm64", "scnnr_bins/mac/arm/scnnr"},
		{"linux", "amd64", "scnnr_bins/linux/intel/scnnr"},
		{"linux", "arm64", "scnnr_bins/linux/arm/scnnr"},
		{"windows", "386", "scnnr_bins/windows/scnnr.exe"},
	}
	calls := 0
	execute := func(cmd *exec.Cmd) error {
		if calls >= len(want) {
			t.Fatal("unexpected extra build")
		}
		target := want[calls]
		calls++
		args := []string{"go", "build", "-o", filepath.FromSlash(target.output), "main.go"}
		if cmd.Dir != root || !reflect.DeepEqual(cmd.Args, args) {
			t.Fatalf("build command: dir=%q args=%q; want %q", cmd.Dir, cmd.Args, args)
		}
		if envValue(cmd.Env, "GOOS") != target.goos || envValue(cmd.Env, "GOARCH") != target.arch ||
			envValue(cmd.Env, "CGO_ENABLED") != "0" || envValue(cmd.Env, "KEEP") != "value" {
			t.Fatalf("wrong target environment for %s/%s", target.goos, target.arch)
		}
		return fakeBuild(cmd)
	}
	if err := buildAll(root, environment, io.Discard, io.Discard, execute); err != nil {
		t.Fatal(err)
	}
	if calls != len(want) || !reflect.DeepEqual(environment, before) ||
		!reflect.DeepEqual(processEnv, []string{os.Getenv("GOOS"), os.Getenv("GOARCH"), os.Getenv("CGO_ENABLED")}) {
		t.Fatal("build plan incomplete or environment mutated")
	}
	for name, want := range map[string]string{"README.md": "readme\n", "LICENSE": "license\n", "version": "42\n"} {
		got, err := os.ReadFile(filepath.Join(root, "scnnr_bins", name))
		if err != nil || string(got) != want {
			t.Fatalf("artifact %q = %q, %v; want %q", name, got, err, want)
		}
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(root, "scnnr_bins/mac/intel/scnnr"))
		if err != nil || info.Mode().Perm() != 0755 {
			t.Fatalf("binary permissions: info=%v err=%v", info, err)
		}
	}
}

func TestBuildCommandFailureStopsPlan(t *testing.T) {
	root := buildFixture(t)
	wantErr := errors.New("compiler failed")
	calls := 0
	err := buildAll(root, nil, io.Discard, io.Discard, func(*exec.Cmd) error {
		calls++
		return wantErr
	})
	if !errors.Is(err, wantErr) || calls != 1 {
		t.Fatalf("build failure: calls=%d error=%v", calls, err)
	}
	if _, err := os.Stat(filepath.Join(root, "scnnr_bins/README.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("failed build continued to copy artifacts")
	}
}

func TestBuildFilesystemFailures(t *testing.T) {
	t.Run("directory creation", func(t *testing.T) {
		root := buildFixture(t)
		if err := os.WriteFile(filepath.Join(root, "scnnr_bins"), []byte("blocked"), 0600); err != nil {
			t.Fatal(err)
		}
		called := false
		err := buildAll(root, nil, io.Discard, io.Discard, func(*exec.Cmd) error { called = true; return nil })
		if err == nil || called {
			t.Fatalf("directory creation failure: err=%v compilerCalled=%v", err, called)
		}
	})
	t.Run("binary missing", func(t *testing.T) {
		err := buildAll(buildFixture(t), nil, io.Discard, io.Discard, func(*exec.Cmd) error { return nil })
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("missing compiler output error=%v", err)
		}
	})
	t.Run("documentation missing", func(t *testing.T) {
		err := buildAll(t.TempDir(), nil, io.Discard, io.Discard, fakeBuild)
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("missing documentation error=%v", err)
		}
	})
	t.Run("version output blocked", func(t *testing.T) {
		root := buildFixture(t)
		if err := os.MkdirAll(filepath.Join(root, "scnnr_bins/version"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := buildAll(root, []string{"CI=1", "VERSION=42"}, io.Discard, io.Discard, fakeBuild); err == nil {
			t.Fatal("version write failure ignored")
		}
	})
	t.Run("documentation output blocked", func(t *testing.T) {
		root := buildFixture(t)
		if err := os.MkdirAll(filepath.Join(root, "scnnr_bins/README.md"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := buildAll(root, nil, io.Discard, io.Discard, fakeBuild); err == nil {
			t.Fatal("documentation write failure ignored")
		}
	})
}

func TestBuildVersionRequiresCIAndVersion(t *testing.T) {
	for _, environment := range [][]string{nil, {"CI=1"}, {"VERSION=42"}} {
		root := buildFixture(t)
		if err := buildAll(root, environment, io.Discard, io.Discard, fakeBuild); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(root, "scnnr_bins/version")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("wrote version without both CI and VERSION")
		}
	}
}
