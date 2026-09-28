package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestChecksumFormatAndFailures(t *testing.T) {
	const sum = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	var stdout bytes.Buffer
	line, err := checksum(strings.NewReader("abc"), "sample.zip", &stdout)
	if err != nil {
		t.Fatal(err)
	}
	if stdout.String() != sum+"\n" || line != sum+"  sample.zip\n" {
		t.Fatalf("stdout=%q checksum=%q", stdout.String(), line)
	}
	wantErr := errors.New("I/O failed")
	if _, err := checksum(failingReader{wantErr}, "sample.zip", io.Discard); !errors.Is(err, wantErr) {
		t.Fatalf("read error=%v", err)
	}
	if _, err := checksum(strings.NewReader("abc"), "sample.zip", failingWriter{wantErr}); !errors.Is(err, wantErr) {
		t.Fatalf("write error=%v", err)
	}
}

func TestRunChecksumFiles(t *testing.T) {
	root := t.TempDir()
	source, destination := filepath.Join(root, "sample.zip"), filepath.Join(root, "sample.zip.sha256")
	if err := os.WriteFile(source, []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run(source, destination, io.Discard); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad  " + source + "\n"
	if string(got) != want {
		t.Fatalf("checksum file=%q; want %q", got, want)
	}
	if err := run(filepath.Join(root, "missing"), destination, io.Discard); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing source error=%v", err)
	}
	if err := run(source, filepath.Join(root, "missing", "sum"), io.Discard); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination error=%v", err)
	}
	if err := run(root, destination, io.Discard); err == nil {
		t.Fatal("directory source accepted")
	}
}
