package scnnr

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// Scanner configures a content search. Search returns results without printing
// or changing the result fields; Scan also populates those fields and prints.
type Scanner struct {
	sync.Mutex
	Regex            bool
	ShowLines        bool
	ShowCols         bool
	Directory        string
	FileExtensions   []string
	Keywords         []string
	KeywordMatches   []string
	ExcludeDirs      []string
	ExcludeExts      []string
	AllMatches       []Match
	MatchedFilePaths []FileData
}

// FileData contains a file's path and metadata.
type FileData struct {
	Path string
	Info os.FileInfo
}

// Match identifies a keyword occurrence. Lines and byte columns are one-based.
type Match struct {
	Path    string
	Line    int
	Column  int
	Keyword string
}

// ScanResult contains candidate files, matching paths, and keyword occurrences.
// Without keywords, Paths contains every candidate and Matches is empty.
type ScanResult struct {
	Files   []FileData
	Paths   []string
	Matches []Match
}

// Search walks Directory and returns results in file-path order. Empty keyword
// and extension lists mean list files and include all extensions, respectively.
func (s *Scanner) Search() (ScanResult, error) {
	keywords := nonEmpty(s.Keywords)

	positions := s.ShowLines || s.ShowCols
	if positions && len(keywords) == 0 {
		return ScanResult{}, errors.New("position tracking flags (-l, -c) require keywords (-k)")
	}

	matcher, err := newContentMatcher(keywords, s.Regex, positions)
	if err != nil {
		return ScanResult{}, err
	}

	extensions := nonEmpty(s.FileExtensions)

	var result ScanResult

	err = walkFiles(s.Directory, s.ExcludeDirs, func(path string) bool {
		return includesExtension(filepath.Ext(path), extensions, s.ExcludeExts)
	}, func(file FileData) error {
		result.Files = append(result.Files, file)

		return nil
	})

	if err != nil {
		return ScanResult{}, err
	}

	if len(keywords) == 0 {
		for _, file := range result.Files {
			result.Paths = append(result.Paths, file.Path)
		}

		return result, nil
	}

	// Keep file I/O bounded independently of the number of discovered files.
	slots := make(chan struct{}, concurrentFiles)

	for _, chunk := range eachSlice(result.Files) {
		matches := make([][]Match, len(chunk))
		errs := make([]error, len(chunk))

		var wg sync.WaitGroup

		wg.Add(len(chunk))

		for i, file := range chunk {
			slots <- struct{}{}

			go func() {
				defer func() { <-slots; wg.Done() }()
				matches[i], errs[i] = parseFile(file.Path, matcher)
			}()
		}

		wg.Wait()

		for i, fileMatches := range matches {
			if errs[i] != nil {
				return ScanResult{}, errs[i]
			}

			if len(fileMatches) > 0 {
				result.Paths = append(result.Paths, chunk[i].Path)
				result.Matches = append(result.Matches, fileMatches...)
			}
		}
	}

	return result, nil
}

// Scan preserves the original API: result fields accumulate and output goes to
// stdout. Walk errors are returned; file-reading errors use the original fatal
// handling. Search remains an independent, additive result-oriented API.
func (s *Scanner) Scan() error {
	return s.ScanTo(os.Stdout)
}

// ScanTo supplies the output seam for testing Scan without changing its behavior.
// Like Scan's original fmt.Println calls, output errors are not returned.
func (s *Scanner) ScanTo(w io.Writer) error {
	if err := filepath.Walk(s.Directory, s.scan); err != nil {
		return err
	}

	if s.Keywords[0] == "" {
		var paths []string

		for _, file := range s.MatchedFilePaths {
			paths = append(paths, file.Path)
		}

		fmt.Fprintln(w, strings.Join(paths, "\n"))

		return nil
	}
	for _, chunk := range legacySlices(s.MatchedFilePaths) {
		var wg sync.WaitGroup

		wg.Add(len(chunk))

		for _, file := range chunk {
			go func() {
				defer wg.Done()
				s.parse(file)
			}()
		}

		wg.Wait()
	}

	if s.ShowLines || s.ShowCols {
		options := OutputOptions{
			ShowLines: s.ShowLines, ShowCols: s.ShowCols,
			ShowKeywords: len(s.Keywords) > 1,
		}

		for _, match := range s.AllMatches {
			fmt.Fprintln(w, formatMatch(match, options))
		}
	} else {
		fmt.Fprintln(w, strings.Join(s.KeywordMatches, "\n"))
	}

	return nil
}

func (s *Scanner) scan(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() && slices.Contains(s.ExcludeDirs, filepath.Base(path)) {
		return filepath.SkipDir
	}

	if info.IsDir() || slices.Contains(s.ExcludeExts, filepath.Ext(path)) {
		return nil
	}

	if s.FileExtensions[0] == "" {
		s.MatchedFilePaths = append(s.MatchedFilePaths, FileData{path, info})
	} else {
		for _, extension := range s.FileExtensions {
			if filepath.Ext(path) == extension {
				s.MatchedFilePaths = append(s.MatchedFilePaths, FileData{path, info})
			}
		}
	}

	return nil
}

func (s *Scanner) parse(fileData FileData) {
	file, err := os.Open(fileData.Path)

	check(err)

	matcher := contentMatcher{
		keywords: s.Keywords, regex: s.Regex,
		positions: s.ShowLines || s.ShowCols,
	}

	err = matcher.parseLegacy(file, fileData.Path, fileData.Info.Size(), func(match Match) {
		s.Lock()
		defer s.Unlock()
		if matcher.positions {
			s.AllMatches = append(s.AllMatches, match)
		} else {
			s.KeywordMatches = append(s.KeywordMatches, match.Path)
		}
	})

	check(err)

	file.Close()
}

// legacySlices preserves Scan's existing chunk-boundary behavior. The additive
// Search API uses eachSlice, independently of this compatibility path.
func legacySlices(files []FileData) [][]FileData {
	var chunks [][]FileData
	var chunk []FileData

	for _, file := range files {
		if len(chunk) == filesPerChunk {
			chunks = append(chunks, chunk)
			chunk = nil
		} else {
			chunk = append(chunk, file)
		}
	}

	if len(chunk) < filesPerChunk {
		chunks = append(chunks, chunk)
	}

	return chunks
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func includesExtension(ext string, extensions, excluded []string) bool {
	if slices.Contains(excluded, ext) {
		return false
	}

	return len(extensions) == 0 || slices.Contains(extensions, ext)
}

func parseFile(path string, matcher *contentMatcher) (matches []Match, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}

	defer func() { err = errors.Join(err, file.Close()) }()

	matches, err = matcher.parse(file, path)

	if err != nil {
		err = fmt.Errorf("read %q: %w", path, err)
	}

	return matches, err
}

const (
	filesPerChunk   = 1024
	concurrentFiles = 128
)

func eachSlice(files []FileData) [][]FileData {
	var chunks [][]FileData

	for start := 0; start < len(files); start += filesPerChunk {
		chunks = append(chunks, files[start:min(start+filesPerChunk, len(files))])
	}

	return chunks
}

// OutputOptions controls the CLI-compatible result format.
type OutputOptions struct {
	ShowLines    bool
	ShowCols     bool
	ShowKeywords bool
}

// WriteResults writes one path or occurrence per line and reports write errors.
func WriteResults(w io.Writer, result ScanResult, options OutputOptions) error {
	if options.ShowLines || options.ShowCols {
		for _, match := range result.Matches {
			if _, err := fmt.Fprintln(w, formatMatch(match, options)); err != nil {
				return err
			}
		}

		return nil
	}

	for _, path := range result.Paths {
		if _, err := fmt.Fprintln(w, path); err != nil {
			return err
		}
	}

	return nil
}

func formatMatch(match Match, options OutputOptions) string {
	output := match.Path

	if options.ShowCols {
		output = fmt.Sprintf("%s:%d:%d", output, match.Line, match.Column)
	} else if options.ShowLines {
		output = fmt.Sprintf("%s:%d", output, match.Line)
	}

	if options.ShowKeywords {
		output += ":" + match.Keyword
	}

	return output
}
