package scnnr

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func writeFixture(t *testing.T, root, name, content string) string {
	t.Helper()

	path := filepath.Join(root, name)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	return path
}

func expectPanic(t *testing.T, call func()) any {
	t.Helper()

	var value any

	func() {
		defer func() { value = recover() }()
		call()
	}()

	if value == nil {
		t.Fatal("expected the original API to panic")
	}

	return value
}

func TestContainsAny(t *testing.T) {
	for _, test := range []struct {
		value    string
		keywords []string
		want     bool
	}{
		{"needle.txt", []string{"needle"}, true},
		{"needle.txt", []string{"absent", ".txt"}, true},
		{"needle.txt", []string{"NEEDLE"}, false},
		{"needle.txt", nil, false},
		{"needle.txt", []string{""}, false},
	} {
		if got := containsAny(test.value, test.keywords); got != test.want {
			t.Errorf("containsAny(%q, %q)=%v; want %v", test.value, test.keywords, got, test.want)
		}
	}
}

func TestFileNameFinderLegacyPathsDuplicatesAndAccumulation(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()

	writeFixture(t, first, "needle.txt", "abc")
	writeFixture(t, first, filepath.Join("nested", "needle.go"), "abc")
	writeFixture(t, second, "needle.txt", "abc")
	writeFixture(t, first, "unrelated.txt", "needle")

	finder := NewFileNameFinder([]string{"needle", ".go"})
	finder.Scan(first)
	finder.Scan(second)

	nested := filepath.Join(first, "nested")

	// These strings pin the original path construction, including its quirks.
	want := []string{first + first + "needle.txt", nested + nested + "needle.go", nested + nested + "needle.go", second + second + "needle.txt"}

	if !reflect.DeepEqual(finder.Files, want) {
		t.Fatalf("files=%q; want original results %q", finder.Files, want)
	}

	expectPanic(t, func() { finder.Scan(filepath.Join(first, "missing")) })

	if !reflect.DeepEqual(finder.Files, want) {
		t.Fatal("root panic changed accumulated results")
	}
}

func TestFinderConstructorsAndInvalidSizePanic(t *testing.T) {
	if finder := NewFileNameFinder([]string{"needle"}); finder.Direction != PathDirection() || !reflect.DeepEqual(finder.Keywords, []string{"needle"}) {
		t.Fatalf("name finder defaults=%#v", finder)
	}

	if finder := NewFileFingerprintFinder([]string{"hash"}); finder.Direction != PathDirection() || !reflect.DeepEqual(finder.FingerPrints, []string{"hash"}) {
		t.Fatalf("fingerprint finder defaults=%#v", finder)
	}

	for size, want := range map[string]int64{
		"1MB": 1000000, "10MB": 10000000, "100MB": 100000000,
		"1GB": 1000000000, "10GB": 10000000000, "100GB": 100000000000, "1TB": 1000000000000,
	} {
		got, err := ParseSize(size)

		if err != nil || got != want {
			t.Errorf("ParseSize(%q)=%d, %v; want %d", size, got, err, want)
		}

		finder := NewFileSizeFinder(size)

		if finder.Size != want || finder.Direction != PathDirection() {
			t.Errorf("constructor threshold for %q=%d", size, finder.Size)
		}
	}
	for _, size := range []string{"", "0MB", "1MiB", "1mb", "garbage"} {
		if _, err := ParseSize(size); err == nil {
			t.Errorf("additive ParseSize accepted %q", size)
		}

		value := expectPanic(t, func() { NewFileSizeFinder(size) })

		if value != "please provide a size 1MB 10MB 100MB 1GB 10GB 100GB 1TB" {
			t.Fatalf("constructor panic=%v", value)
		}
	}
}

func TestFileSizeFinderLegacyThresholdsAndPaths(t *testing.T) {
	root := t.TempDir()

	writeFixture(t, root, "below", "ab")
	writeFixture(t, root, "equal", "abc")
	writeFixture(t, root, filepath.Join("nested", "above"), "abcd")

	finder := NewFileSizeFinder("1MB")
	finder.Size = 3
	finder.Scan(root)

	nested := filepath.Join(root, "nested")

	want := []string{root + root + "equal", nested + nested + "above"}

	if !reflect.DeepEqual(finder.Files, want) {
		t.Fatalf("threshold files=%q; want %q", finder.Files, want)
	}

	finder.Scan(root)

	if !reflect.DeepEqual(finder.Files, append(append([]string(nil), want...), want...)) {
		t.Fatal("size finder no longer accumulates")
	}

	expectPanic(t, func() { finder.Scan(filepath.Join(root, "missing")) })

	for _, threshold := range []int64{0, -1} {
		all := NewFileSizeFinder("1MB")

		all.Size = threshold

		all.Scan(root)

		if len(all.Files) != 3 {
			t.Fatalf("original threshold %d should include all files; got %q", threshold, all.Files)
		}
	}
}

func TestFileFingerprintFinderLegacyMatching(t *testing.T) {
	const abcHash = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	const emptyHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	root := t.TempDir()

	abc := writeFixture(t, root, "abc.txt", "abc")
	empty := writeFixture(t, root, filepath.Join("nested", "empty"), "")
	other := writeFixture(t, root, "other", "different")

	for _, test := range []struct {
		name         string
		hashes, want []string
	}{
		{"full digest", []string{abcHash}, []string{abc}},
		{"substring", []string{abcHash[:12]}, []string{abc}},
		{"multiple digests", []string{emptyHash, abcHash}, []string{abc, empty}},
		{"overlapping digests produce duplicates", []string{abcHash, abcHash[:12]}, []string{abc, abc}},
		{"no match", []string{"unknown"}, nil},
		{"empty fingerprint matches every digest", []string{""}, []string{abc, other, empty}},
	} {
		t.Run(test.name, func(t *testing.T) {
			finder := NewFileFingerprintFinder(test.hashes)

			finder.Scan(root)

			if !reflect.DeepEqual(finder.Files, test.want) {
				t.Fatalf("files=%q; want %q", finder.Files, test.want)
			}
		})
	}

	finder := NewFileFingerprintFinder([]string{abcHash})

	// tests accumulation across calls.
	// Scan appends to finder.Files
	// so scanning twice should leave []string{abc, abc}.
	finder.Scan(root)
	finder.Scan(root)

	if !reflect.DeepEqual(finder.Files, []string{abc, abc}) {
		t.Fatal("fingerprint finder no longer accumulates")
	}

	expectPanic(t, func() { finder.Scan(filepath.Join(root, "missing")) })

	// The additive reader helper can still report errors independently.
	if _, err := fileDigest(filepath.Join(root, "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("fileDigest missing file error=%v", err)
	}
}

func TestFindersContinueAfterBrokenEntries(t *testing.T) {
	root := t.TempDir()

	abc := writeFixture(t, root, "a-needle.txt", "abc")

	nested := writeFixture(t, root, filepath.Join("nested", "b-needle.txt"), "abc")

	if err := os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "broken")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	name := NewFileNameFinder([]string{"needle"})

	size := NewFileSizeFinder("1MB")
	size.Size = 1

	hash := NewFileFingerprintFinder([]string{"ba7816"})

	name.Scan(root)

	size.Scan(root)

	hash.Scan(root)

	nestedDir := filepath.Join(root, "nested")

	wantPaths := []string{root + root + "a-needle.txt", nestedDir + nestedDir + "b-needle.txt"}

	if !reflect.DeepEqual(name.Files, wantPaths) || !reflect.DeepEqual(size.Files, wantPaths) {
		t.Fatalf("name/size scans stopped: name=%q size=%q", name.Files, size.Files)
	}

	if !reflect.DeepEqual(hash.Files, []string{abc, nested}) {
		t.Fatalf("hash scan stopped: %q", hash.Files)
	}
}

func TestFingerprintFinderContinuesAfterReadFailure(t *testing.T) {
	root := t.TempDir()

	target := t.TempDir()

	link := filepath.Join(root, "a-unreadable-directory-link")

	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	abc := writeFixture(t, root, "z-abc.txt", "abc")

	const emptyHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	finder := NewFileFingerprintFinder([]string{emptyHash, "ba7816"})

	finder.Scan(root)
	// ReadFile on the directory fails. The original API hashes the returned
	// empty bytes and continues to the following readable file.
	if !reflect.DeepEqual(finder.Files, []string{link, abc}) {
		t.Fatalf("read-failure behavior changed: %q", finder.Files)
	}
}

func TestEmptyNameKeywordsAndCustomDirection(t *testing.T) {
	root := t.TempDir()

	writeFixture(t, root, "plain.txt", "abc")

	all := NewFileNameFinder([]string{""})

	all.Scan(root)

	if !reflect.DeepEqual(all.Files, []string{root + root + "plain.txt"}) {
		t.Fatalf("empty keyword behavior=%q", all.Files)
	}

	missingMetadata := NewFileNameFinder([]string{""})
	missingMetadata.Direction = "|"
	missingMetadata.Scan(root)

	if len(missingMetadata.Files) != 0 {
		t.Fatalf("custom direction's failed stats were not skipped: %q", missingMetadata.Files)
	}
}

func TestPathHelpersKeepSuppliedDirection(t *testing.T) {
	root := t.TempDir()
	path := writeFixture(t, root, "plain.txt", "abc")
	info, err := os.Stat(path)

	if err != nil {
		t.Fatal(err)
	}

	for _, direction := range []string{"/", "|", "", root} {
		want := root + direction + "plain.txt"

		if got := FullFilePath(root, direction, info); got != want {
			t.Fatalf("FullFilePath direction=%q got=%q; want %q", direction, got, want)
		}
	}

	files, dirs := CollectFilesAndDirs(root, PathDirection())

	if len(files) != 1 || len(dirs) != 0 || files[0].Name() != "plain.txt" {
		t.Fatalf("collection files=%v dirs=%v", files, dirs)
	}

	files, dirs = CollectFilesAndDirs(filepath.Join(root, "missing"), PathDirection())

	if len(files)+len(dirs) != 0 {
		t.Fatal("failed directory read should return no entries")
	}
}

func TestAdditiveWalkCallbackAndSymlinks(t *testing.T) {
	root := t.TempDir()
	target := writeFixture(t, root, "target.txt", "abc")
	wantErr := errors.New("visitor failed")

	if err := walkFiles(root, nil, nil, func(FileData) error { return wantErr }); !errors.Is(err, wantErr) {
		t.Fatalf("visitor error=%v", err)
	}

	if err := os.Symlink(target, filepath.Join(root, "file-link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if err := os.Symlink(root, filepath.Join(root, "dir-link")); err != nil {
		t.Fatal(err)
	}

	var paths []string

	if err := walkFiles(root, nil, nil, func(file FileData) error {
		paths = append(paths, file.Path)

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(paths, []string{filepath.Join(root, "file-link"), target}) {
		t.Fatalf("additive traversal=%q", paths)
	}
}
