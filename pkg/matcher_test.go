package scnnr

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestContentMatcher(t *testing.T) {
	tests := []struct {
		name, input string
		keywords    []string
		regex, all  bool
		want        []Match
	}{
		{
			name: "first match only", input: "hello needle needle\nneedle\n",
			keywords: []string{"needle"},
			want:     []Match{{"sample", 1, 7, "needle"}},
		},
		{
			name: "every literal occurrence", input: "hello needle needle\nsecond token\n",
			keywords: []string{"needle", "token"}, all: true,
			want: []Match{{"sample", 1, 7, "needle"}, {"sample", 1, 14, "needle"}, {"sample", 2, 8, "token"}},
		},
		{
			name: "every regex occurrence", input: "id12 id345\nid6",
			keywords: []string{`id\d+`}, regex: true, all: true,
			want: []Match{{"sample", 1, 1, `id\d+`}, {"sample", 1, 6, `id\d+`}, {"sample", 2, 1, `id\d+`}},
		},
		{
			name: "regex first match", input: "id12 id345",
			keywords: []string{`id\d+`}, regex: true,
			want: []Match{{"sample", 1, 1, `id\d+`}},
		},
		{
			name: "CRLF blank lines and final unterminated line", input: "\r\nneedle\r\n\r\nneedle",
			keywords: []string{"needle"}, all: true,
			want: []Match{{"sample", 2, 1, "needle"}, {"sample", 4, 1, "needle"}},
		},
		{
			name: "byte columns", input: "é needle",
			keywords: []string{"needle"}, all: true,
			want: []Match{{"sample", 1, 4, "needle"}},
		},
		{
			name: "literal metacharacters and spaces", input: "a+b a+b",
			keywords: []string{"a+b", " "}, all: true,
			want: []Match{{"sample", 1, 1, "a+b"}, {"sample", 1, 5, "a+b"}, {"sample", 1, 4, " "}},
		},
		{
			name: "zero width regex", input: "abc",
			keywords: []string{"^", "$"}, regex: true, all: true,
			want: []Match{{"sample", 1, 1, "^"}, {"sample", 1, 4, "$"}},
		},
		{
			name: "long line", input: strings.Repeat("x", 100000) + "needle",
			keywords: []string{"needle"}, all: true,
			want: []Match{{"sample", 1, 100001, "needle"}},
		},
		{name: "empty input", keywords: []string{"needle"}, all: true},
		{name: "no match", input: "unrelated\n", keywords: []string{"needle"}, all: true},
		{name: "empty keyword entries", input: "hello", keywords: []string{"", ""}, all: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matcher, err := newContentMatcher(test.keywords, test.regex, test.all)
			if err != nil {
				t.Fatal(err)
			}
			got, err := matcher.parse(strings.NewReader(test.input), "sample")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("matches = %#v; want %#v", got, test.want)
			}
		})
	}
}

func TestContentMatcherErrors(t *testing.T) {
	if _, err := newContentMatcher([]string{"["}, true, true); err == nil {
		t.Fatal("invalid regex accepted")
	}
	wantErr := errors.New("reader failed")
	for _, all := range []bool{false, true} {
		matcher, err := newContentMatcher([]string{"needle"}, false, all)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := matcher.parse(failingReader{wantErr}, "sample"); !errors.Is(err, wantErr) {
			t.Fatalf("read error = %v; want %v", err, wantErr)
		}
		input := io.MultiReader(strings.NewReader("needle\n"), failingReader{wantErr})
		matches, err := matcher.parse(input, "sample")
		if all && !errors.Is(err, wantErr) {
			t.Fatalf("position scan lost reader failure: %v", err)
		}
		if !all && (err != nil || len(matches) != 1) {
			t.Fatalf("first-match scan kept reading: matches=%v, err=%v", matches, err)
		}
	}
}

func TestLegacyReaderMatching(t *testing.T) {
	for _, test := range []struct {
		name, input      string
		keywords         []string
		regex, positions bool
		want             []Match
	}{
		{"first occurrence per keyword per line", "needle needle\ntoken token\n", []string{"needle", "token"}, false, true, []Match{{"sample", 1, 1, "needle"}, {"sample", 2, 1, "token"}}},
		{"one result per file", "needle needle\nneedle\n", []string{"needle"}, false, false, []Match{{"sample", 1, 1, "needle"}}},
		{"regex first occurrence per line", "id1 id2\nid3", []string{"id[0-9]"}, true, true, []Match{{"sample", 1, 1, "id[0-9]"}, {"sample", 2, 1, "id[0-9]"}}},
		{"empty keyword retained", "nothing\n\n", []string{""}, false, true, []Match{{"sample", 1, 1, ""}, {"sample", 2, 1, ""}}},
		{"duplicate keywords retained", "needle", []string{"needle", "needle"}, false, true, []Match{{"sample", 1, 1, "needle"}, {"sample", 1, 1, "needle"}}},
		{"byte columns and CRLF", "é needle\r\n", []string{"needle"}, false, true, []Match{{"sample", 1, 4, "needle"}}},
		{"empty input does not compile invalid regex", "", []string{"["}, true, true, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			matcher := contentMatcher{keywords: test.keywords, regex: test.regex, positions: test.positions}
			var got []Match
			err := matcher.parseLegacy(strings.NewReader(test.input), "sample", int64(len(test.input)), func(match Match) { got = append(got, match) })
			if err != nil || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("legacy matches=%#v error=%v; want %#v", got, err, test.want)
			}
		})
	}
	invalid := contentMatcher{keywords: []string{"["}, regex: true}
	expectPanic(t, func() { invalid.parseLegacy(strings.NewReader("line"), "sample", 4, func(Match) {}) })
}

func TestLegacyReaderKeepsReadingAfterMatch(t *testing.T) {
	wantErr := errors.New("failure after matching line")
	for _, positions := range []bool{false, true} {
		matcher := contentMatcher{keywords: []string{"needle"}, positions: positions}
		var got []Match
		input := io.MultiReader(strings.NewReader("needle\n"), failingReader{wantErr})
		err := matcher.parseLegacy(input, "sample", 100, func(match Match) { got = append(got, match) })
		if !errors.Is(err, wantErr) || len(got) != 1 {
			t.Fatalf("original reader behavior positions=%v matches=%v error=%v", positions, got, err)
		}
	}
}
