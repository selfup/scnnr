package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func archiveFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "bins")
	if err := os.MkdirAll(filepath.Join(root, "nested", "empty"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string][]byte{
		"readme.txt":                      []byte("hello\n"),
		filepath.Join("nested", "binary"): {0, 1, 2, 255},
		"empty":                           {},
	} {
		if err := os.WriteFile(filepath.Join(root, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func archiveContents(t *testing.T, data []byte) map[string]string {
	t.Helper()
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	contents := make(map[string]string)
	for _, entry := range archive.File {
		if _, exists := contents[entry.Name]; exists {
			t.Fatalf("duplicate archive entry %q", entry.Name)
		}
		file, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(file)
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			t.Fatalf("archive read=%v close=%v", err, closeErr)
		}
		contents[entry.Name] = string(content)
	}
	return contents
}

func TestWriteZipContents(t *testing.T) {
	root := archiveFixture(t)
	var output bytes.Buffer
	if err := writeZip(&output, root); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"bins/": "", "bins/empty": "", "bins/nested/": "",
		"bins/nested/binary": string([]byte{0, 1, 2, 255}),
		"bins/nested/empty/": "", "bins/readme.txt": "hello\n",
	}
	if got := archiveContents(t, output.Bytes()); !reflect.DeepEqual(got, want) {
		t.Fatalf("archive contents=%v; want %v", got, want)
	}
	output.Reset()
	if err := writeZip(&output, filepath.Join(root, "readme.txt")); err != nil {
		t.Fatal(err)
	}
	if got := archiveContents(t, output.Bytes()); !reflect.DeepEqual(got, map[string]string{"readme.txt": "hello\n"}) {
		t.Fatalf("single-file archive=%v", got)
	}
}

func TestArchiveMissingSourceAndOutputErrors(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "out.zip")
	if err := zipBins(filepath.Join(root, "missing"), target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing source error=%v", err)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("missing source created an output archive")
	}
	if err := zipBins(archiveFixture(t), filepath.Join(root, "missing", "out.zip")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("output creation error=%v", err)
	}
	wantErr := errors.New("archive writer failed")
	// The small archive fits in zip.Writer's buffer; the error is surfaced by Close.
	if err := writeZip(failingWriter{wantErr}, archiveFixture(t)); !errors.Is(err, wantErr) {
		t.Fatalf("archive close/write error=%v", err)
	}
}

func TestArchiveReadFailure(t *testing.T) {
	root := archiveFixture(t)
	if err := os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "broken")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := writeZip(io.Discard, root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unreadable archive entry error=%v", err)
	}
}

func TestPackBuildFailureAndSuccess(t *testing.T) {
	root := archiveFixture(t)
	target := filepath.Join(t.TempDir(), "out.zip")
	wantErr := errors.New("build failed")
	if err := run(root, target, func() error { return wantErr }); !errors.Is(err, wantErr) {
		t.Fatalf("build error=%v", err)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("failed build created an archive")
	}
	called := false
	if err := run(root, target, func() error { called = true; return nil }); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !called || archiveContents(t, data)["bins/readme.txt"] != "hello\n" {
		t.Fatal("successful build did not produce expected archive")
	}
}
