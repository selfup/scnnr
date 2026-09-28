package digest

import (
	"errors"
	"strings"
	"testing"
)

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestSHA256(t *testing.T) {
	for _, test := range []struct{ input, want string }{
		{"abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	} {
		got, err := SHA256(strings.NewReader(test.input))
		if err != nil || got != test.want {
			t.Fatalf("SHA256(%q) = %q, %v; want %q", test.input, got, err, test.want)
		}
	}
	wantErr := errors.New("read failed")
	if _, err := SHA256(failingReader{wantErr}); !errors.Is(err, wantErr) {
		t.Fatalf("read error = %v", err)
	}
}
