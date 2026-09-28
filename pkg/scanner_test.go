package scnnr

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
	"testing"
)

func fixturePath(parts ...string) string {
	return filepath.Join(append([]string{"testdata", "tree"}, parts...)...)
}

func TestSearchFilters(t *testing.T) {
	all := []string{
		fixturePath("alpha.txt"), fixturePath("empty.txt"), fixturePath("ignored", "hidden.txt"),
		fixturePath("nested", "beta.go"), fixturePath("nested", "quiet.txt"), fixturePath("skip.log"),
	}
	tests := []struct {
		name                    string
		extensions, excludeExts []string
		excludeDirs             []string
		want                    []string
	}{
		{name: "nil lists", want: all},
		{name: "legacy empty sentinel", extensions: []string{""}, want: all},
		{
			name: "duplicate extension", extensions: []string{".txt", ".txt"},
			want: []string{all[0], all[1], all[2], all[4]},
		},
		{
			name:        "excluded directory and extension",
			excludeDirs: []string{"ignored"}, excludeExts: []string{".log"},
			want: []string{all[0], all[1], all[3], all[4]},
		},
		{name: "exclusion wins", extensions: []string{".txt"}, excludeExts: []string{".txt"}},
		{name: "excluded root", excludeDirs: []string{"tree"}},
		{name: "unknown extension", extensions: []string{".missing"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scanner := Scanner{
				Directory: fixturePath(), FileExtensions: test.extensions,
				ExcludeDirs: test.excludeDirs, ExcludeExts: test.excludeExts,
			}
			result, err := scanner.Search()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Paths, test.want) {
				t.Fatalf("paths = %q; want %q", result.Paths, test.want)
			}
			if len(result.Matches) != 0 || len(result.Files) != len(test.want) {
				t.Fatalf("unexpected listing result: %#v", result)
			}
		})
	}
}

func TestSearchFiltersBeforeReadingMetadata(t *testing.T) {
	root := t.TempDir()
	want := writeFixture(t, root, "good.txt", "needle")
	if err := os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "broken.log")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	for _, test := range []struct {
		name                 string
		extensions, excluded []string
	}{
		{"excluded extension", nil, []string{".log"}},
		{"included extensions only", []string{".txt"}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			scanner := Scanner{
				Directory: root, Keywords: []string{"needle"},
				FileExtensions: test.extensions, ExcludeExts: test.excluded,
			}
			result, err := scanner.Search()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Paths, []string{want}) {
				t.Fatalf("paths=%q; want %q", result.Paths, want)
			}
		})
	}
}

func TestSearchMatchesAreStableAndIndependent(t *testing.T) {
	scanner := Scanner{
		Directory: fixturePath(), Keywords: []string{"needle", "token"}, ShowCols: true,
		ExcludeDirs: []string{"ignored"}, ExcludeExts: []string{".log"},
		KeywordMatches: []string{"previous"}, AllMatches: []Match{{Path: "previous"}},
	}
	want := []Match{
		{fixturePath("alpha.txt"), 1, 7, "needle"},
		{fixturePath("alpha.txt"), 1, 14, "needle"},
		{fixturePath("alpha.txt"), 2, 8, "token"},
		{fixturePath("nested", "beta.go"), 2, 4, "needle"},
	}
	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := scanner.Search()
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(result.Matches, want) {
				t.Errorf("matches = %#v; want %#v", result.Matches, want)
			}
			if !reflect.DeepEqual(result.Paths, []string{fixturePath("alpha.txt"), fixturePath("nested", "beta.go")}) {
				t.Errorf("paths = %q", result.Paths)
			}
		}()
	}
	wg.Wait()
	if !reflect.DeepEqual(scanner.KeywordMatches, []string{"previous"}) || scanner.AllMatches[0].Path != "previous" {
		t.Fatal("Search mutated legacy result fields")
	}
}

func TestSearchFirstMatchMode(t *testing.T) {
	scanner := Scanner{
		Directory: fixturePath(), Keywords: []string{"needle"},
		ExcludeDirs: []string{"ignored"}, ExcludeExts: []string{".log"},
	}
	result, err := scanner.Search()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Matches) != 2 || len(result.Paths) != 2 {
		t.Fatalf("first-match results = %#v", result)
	}
}

func TestSearchErrors(t *testing.T) {
	tests := []struct {
		name    string
		scanner *Scanner
	}{
		{"missing directory", &Scanner{Directory: filepath.Join(t.TempDir(), "missing")}},
		{"empty directory", &Scanner{}},
		{"invalid regex", &Scanner{Directory: fixturePath(), Keywords: []string{"["}, Regex: true}},
		{"lines require keywords", &Scanner{Directory: fixturePath(), ShowLines: true}},
		{"columns require keywords", &Scanner{Directory: fixturePath(), Keywords: []string{""}, ShowCols: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.scanner.Search()
			if err == nil {
				t.Fatal("expected error")
			}
			if len(result.Files)+len(result.Paths)+len(result.Matches) != 0 {
				t.Fatalf("failure returned partial results: %#v", result)
			}
		})
	}
	matcher, err := newContentMatcher([]string{"needle"}, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseFile(filepath.Join(t.TempDir(), "missing"), matcher); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file error = %v", err)
	}
	if _, err := parseFile(t.TempDir(), matcher); err == nil {
		t.Fatal("directory was accepted as readable content")
	}
}

func TestScanAccumulatesLegacyResults(t *testing.T) {
	scanner := Scanner{Directory: fixturePath(), FileExtensions: []string{".go"}, Keywords: []string{"needle"}}
	for scan := 1; scan <= 2; scan++ {
		if err := scanner.ScanTo(&bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		if len(scanner.MatchedFilePaths) != scan || len(scanner.KeywordMatches) != scan*(scan+1)/2 {
			t.Fatalf("original accumulation changed: files=%d paths=%q", len(scanner.MatchedFilePaths), scanner.KeywordMatches)
		}
	}
	scanner.ShowCols = true
	if err := scanner.ScanTo(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if len(scanner.KeywordMatches) != 3 || len(scanner.AllMatches) != 3 {
		t.Fatal("position scan did not retain original accumulated results")
	}
	scanner.ShowCols = false
	scanner.Keywords = []string{""}
	if err := scanner.ScanTo(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if len(scanner.AllMatches) != 3 || len(scanner.KeywordMatches) != 3 || len(scanner.MatchedFilePaths) != 4 {
		t.Fatal("listing scan cleared the original result fields")
	}
}

func TestEachSlicePreservesEveryFile(t *testing.T) {
	for _, count := range []int{0, 1, 1023, 1024, 1025, 2048, 2049, 2050, 4096} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			files := make([]FileData, count)
			for i := range files {
				files[i].Path = fmt.Sprint(i)
			}
			var got []FileData
			for _, chunk := range eachSlice(files) {
				if len(chunk) == 0 || len(chunk) > filesPerChunk {
					t.Fatalf("invalid chunk length %d", len(chunk))
				}
				got = append(got, chunk...)
			}
			if !slices.Equal(got, files) {
				t.Fatalf("lost or reordered files: input=%d output=%d", len(files), len(got))
			}
		})
	}
}

func TestLegacyScanContracts(t *testing.T) {
	t.Run("missing root returns walk error", func(t *testing.T) {
		scanner := Scanner{Directory: filepath.Join(t.TempDir(), "missing"), Keywords: []string{""}, FileExtensions: []string{""}}
		if err := scanner.ScanTo(&bytes.Buffer{}); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("walk error=%v", err)
		}
	})
	t.Run("nil keywords retain panic", func(t *testing.T) {
		scanner := Scanner{Directory: t.TempDir(), FileExtensions: []string{""}}
		expectPanic(t, func() { scanner.ScanTo(&bytes.Buffer{}) })
	})
	t.Run("nil extensions retain panic", func(t *testing.T) {
		root := t.TempDir()
		writeFixture(t, root, "sample", "needle")
		scanner := Scanner{Directory: root, Keywords: []string{"needle"}}
		expectPanic(t, func() { scanner.ScanTo(&bytes.Buffer{}) })
	})
	t.Run("duplicate extensions are retained", func(t *testing.T) {
		root := t.TempDir()
		path := writeFixture(t, root, "sample.txt", "needle")
		scanner := Scanner{Directory: root, Keywords: []string{""}, FileExtensions: []string{".txt", ".txt"}}
		var output bytes.Buffer
		if err := scanner.ScanTo(&output); err != nil {
			t.Fatal(err)
		}
		if output.String() != path+"\n"+path+"\n" || len(scanner.MatchedFilePaths) != 2 {
			t.Fatalf("original duplicate extensions changed: %q", output.String())
		}
	})
	t.Run("empty first keyword disables matching", func(t *testing.T) {
		root := t.TempDir()
		path := writeFixture(t, root, "sample", "unrelated")
		scanner := Scanner{Directory: root, Keywords: []string{"", "needle"}, FileExtensions: []string{""}, ShowCols: true}
		var output bytes.Buffer
		if err := scanner.ScanTo(&output); err != nil || output.String() != path+"\n" {
			t.Fatalf("empty sentinel output=%q error=%v", output.String(), err)
		}
	})
	t.Run("empty results retain newline", func(t *testing.T) {
		scanner := Scanner{Directory: t.TempDir(), Keywords: []string{""}, FileExtensions: []string{""}}
		var output bytes.Buffer
		if err := scanner.ScanTo(&output); err != nil || output.String() != "\n" {
			t.Fatalf("empty output=%q error=%v", output.String(), err)
		}
	})
	t.Run("full batch retains original boundary behavior", func(t *testing.T) {
		scanner := Scanner{
			Directory: t.TempDir(), Keywords: []string{"needle"}, FileExtensions: []string{""},
			MatchedFilePaths: make([]FileData, 1024),
		}
		var output bytes.Buffer
		if err := scanner.ScanTo(&output); err != nil || output.String() != "\n" || len(scanner.KeywordMatches) != 0 {
			t.Fatalf("original 1024-file batch changed: output=%q error=%v", output.String(), err)
		}
	})
}

func TestLegacySlicesMatchOriginalBoundaryBehavior(t *testing.T) {
	for _, test := range []struct{ input, output int }{
		{0, 0}, {1, 1}, {1023, 1023}, {1024, 0},
		{1025, 1024}, {2048, 2047}, {2049, 1024}, {2050, 2048},
	} {
		count := 0
		for _, chunk := range legacySlices(make([]FileData, test.input)) {
			count += len(chunk)
		}
		if count != test.output {
			t.Errorf("legacy batch input=%d output=%d; want original %d", test.input, count, test.output)
		}
	}
}

func TestSearchAtChunkBoundary(t *testing.T) {
	root := t.TempDir()
	for i := range 1025 {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("%04d.txt", i)), []byte("needle\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	scanner := Scanner{Directory: root, Keywords: []string{"needle"}}
	result, err := scanner.Search()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Paths) != 1025 {
		t.Fatalf("returned %d of 1025 matching files", len(result.Paths))
	}
	for i, path := range result.Paths {
		if path != filepath.Join(root, fmt.Sprintf("%04d.txt", i)) {
			t.Fatalf("file %d missing or out of order: %q", i, path)
		}
	}
}

func TestWriteResults(t *testing.T) {
	result := ScanResult{
		Paths:   []string{"a", "b"},
		Matches: []Match{{"a", 2, 3, "needle"}, {"b", 4, 5, "token"}},
	}
	tests := []struct {
		name    string
		options OutputOptions
		want    string
	}{
		{"paths", OutputOptions{}, "a\nb\n"},
		{"lines", OutputOptions{ShowLines: true}, "a:2\nb:4\n"},
		{"columns imply lines", OutputOptions{ShowCols: true}, "a:2:3\nb:4:5\n"},
		{"columns take precedence", OutputOptions{ShowLines: true, ShowCols: true}, "a:2:3\nb:4:5\n"},
		{"keyword labels", OutputOptions{ShowLines: true, ShowKeywords: true}, "a:2:needle\nb:4:token\n"},
		{"columns and keywords", OutputOptions{ShowCols: true, ShowKeywords: true}, "a:2:3:needle\nb:4:5:token\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := WriteResults(&out, result, test.options); err != nil {
				t.Fatal(err)
			}
			if out.String() != test.want {
				t.Fatalf("output = %q; want %q", out.String(), test.want)
			}
		})
	}
	var empty bytes.Buffer
	if err := WriteResults(&empty, ScanResult{}, OutputOptions{}); err != nil || empty.Len() != 0 {
		t.Fatalf("empty result output=%q err=%v", empty.String(), err)
	}
	wantErr := errors.New("output failed")
	for _, options := range []OutputOptions{{}, {ShowCols: true}} {
		if err := WriteResults(failingWriter{wantErr}, result, options); !errors.Is(err, wantErr) {
			t.Fatalf("write error = %v", err)
		}
	}
}
