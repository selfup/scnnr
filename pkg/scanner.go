package scnnr

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Scanner scans files and based on pattern stores in array for goroutine processing
type Scanner struct {
	sync.Mutex
	Regex            bool
	ShowLines        bool
	ShowCols         bool
	Directory        string
	FileExtensions   []string
	Keywords         []string
	KeywordMatches   []string
	AllMatches       []Match
	MatchedFilePaths []FileData
}

// FileData contains the path of the file as well as relevant metadata
type FileData struct {
	Path string
	Info os.FileInfo
}

// Match represents a single keyword match with position information
type Match struct {
	Path    string
	Line    int
	Column  int
	Keyword string
}

// Scan walks the given directory tree and stores all matching files into a slice
func (s *Scanner) Scan() error {
	err := filepath.Walk(s.Directory, s.scan)
	if err != nil {
		return err
	}

	if s.Keywords[0] == "" {
		var foundFiles []string

		for _, fileInfo := range s.MatchedFilePaths {
			foundFiles = append(foundFiles, fileInfo.Path)
		}

		fmt.Println(strings.Join(foundFiles, "\n"))
	} else {
		var wg sync.WaitGroup

		for _, chunk := range eachSlice(s.MatchedFilePaths) {
			matchedCount := len(chunk)

			wg.Add(matchedCount)

			for _, match := range chunk {
				go func(m FileData) {
					s.parse(m)

					wg.Done()
				}(match)
			}

			wg.Wait()
		}

		// Output results based on position tracking settings
		if s.ShowLines || s.ShowCols {
			s.outputPositionMatches()
		} else {
			// Original behavior
			fmt.Println(strings.Join(s.KeywordMatches, "\n"))
		}
	}

	return nil
}

// outputPositionMatches formats and outputs matches with position information
func (s *Scanner) outputPositionMatches() {
	for _, match := range s.AllMatches {
		output := match.Path

		if s.ShowCols {
			// -c flag shows both line and column
			output = fmt.Sprintf("%s:%d:%d", output, match.Line, match.Column)
		} else if s.ShowLines {
			// -l flag shows only line
			output = fmt.Sprintf("%s:%d", output, match.Line)
		}

		// If multiple keywords, append the matching keyword
		if len(s.Keywords) > 1 {
			output = fmt.Sprintf("%s:%s", output, match.Keyword)
		}

		fmt.Println(output)
	}
}

func (s *Scanner) scan(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !info.IsDir() {
		if s.FileExtensions[0] == "" {
			s.MatchedFilePaths = append(s.MatchedFilePaths, FileData{path, info})
		} else {
			for _, pattern := range s.FileExtensions {
				fileExtension := filepath.Ext(path)

				if fileExtension == pattern {
					s.MatchedFilePaths = append(s.MatchedFilePaths, FileData{path, info})
				}
			}
		}
	}

	return nil
}

// If Regex is true the parser will switch to regex mode.
// Otherwise strings.Contains will be used.
func (s *Scanner) parse(match FileData) {
	file, err := os.Open(match.Path)

	check(err)

	scanner := bufio.NewScanner(file)

	buf := make([]byte, 0, 1024)

	// only extend buffer to file size
	scanner.Buffer(buf, 2*int(match.Info.Size()))

	found := false

	lineNumber := 0

	positionTracking := s.ShowLines || s.ShowCols

	for scanner.Scan() {
		line := scanner.Text()

		lineNumber++

		for i := 0; i < len(s.Keywords); i++ {
			if !positionTracking && found {
				break
			}

			matchFound := false
			var column int

			if s.Regex {
				re := regexp.MustCompile(s.Keywords[i])

				if loc := re.FindStringIndex(line); loc != nil {
					matchFound = true
					column = loc[0] + 1
				}
			} else {
				if idx := strings.Index(line, s.Keywords[i]); idx != -1 {
					matchFound = true
					column = idx + 1
				}
			}

			if matchFound {
				s.Lock()

				if positionTracking {
					match := Match{
						Path:    match.Path,
						Line:    lineNumber,
						Column:  column,
						Keyword: s.Keywords[i],
					}

					s.AllMatches = append(s.AllMatches, match)
				} else {
					s.KeywordMatches = append(s.KeywordMatches, match.Path)

					found = true
				}

				s.Unlock()

				if !positionTracking {
					break
				}
			}
		}
	}

	check(scanner.Err())

	file.Close()
}

func eachSlice(files []FileData) [][]FileData {
	var chunks [][]FileData
	var chunk []FileData

	for _, fileData := range files {
		switch len(chunk) {
		case 1024:
			var newChunk []FileData
			chunks = append(chunks, chunk)
			chunk = newChunk
		default:
			chunk = append(chunk, fileData)
		}
	}

	if len(chunk) < 1024 {
		chunks = append(chunks, chunk)
	}

	return chunks
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
