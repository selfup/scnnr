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
	Directory        string
	FilePatterns     []string
	Keywords         []string
	RegexKeywords    []string
	KeywordMatches   []string
	MatchedFilePaths []FileData
}

// FileData contains the path of the file as well as relevant metadata
type FileData struct {
	Path string
	Info os.FileInfo
}

// Scan walks the given directory tree and stores all matching files into a slice
func (s *Scanner) Scan() error {
	err := filepath.Walk(s.Directory, s.scan)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	regex := os.Getenv("SCNNR_REGEX")

	if regex == "1" || len(os.Args) > 4 && os.Args[4] == "rgx" {
		s.RegexKeywords = s.Keywords
	}

	for _, chunk := range eachSlice(s.MatchedFilePaths) {
		matchedCount := len(chunk)
		wg.Add(matchedCount)

		for _, match := range chunk {
			go func(m FileData) {
				s.parse(m, regex)
				wg.Done()
			}(match)
		}

		wg.Wait()
	}

	fmt.Printf("found %d matches: \n%s", len(s.KeywordMatches), strings.Join(s.KeywordMatches, ","))

	return nil
}

func (s *Scanner) scan(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	for _, pattern := range s.FilePatterns {
		if !info.IsDir() {
			fileExtension := filepath.Ext(path)

			if fileExtension == pattern {
				s.MatchedFilePaths = append(s.MatchedFilePaths, FileData{path, info})
			}
		}
	}

	return nil
}

// If RegexKeywords has a length greater than 0 the parser will switch to regex mode.
// Otherwise strings.Contains will be used.
// RegexKeywords will be populated if the 4th argument is "rgx" or if the SCNNR_REGEX ENV is set to "1"
func (s *Scanner) parse(match FileData, regex string) {
	file, err := os.Open(match.Path)
	check(err)

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024)

	// only extend buffer to file size
	scanner.Buffer(buf, 2*int(match.Info.Size()))

	found := false

	for scanner.Scan() {
		line := scanner.Text()

		for i := 0; i < len(s.Keywords); i++ {
			if found {
				break
			}

			if len(s.RegexKeywords) > 0 {
				re := regexp.MustCompile(s.RegexKeywords[i])

				if re.Match([]byte(line)) {
					s.Lock()
					s.KeywordMatches = append(s.KeywordMatches, match.Path)
					found = true
					s.Unlock()
					break
				}
			} else {
				if strings.Contains(line, s.Keywords[i]) {
					s.Lock()
					s.KeywordMatches = append(s.KeywordMatches, match.Path)
					found = true
					s.Unlock()
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

	for i, fileData := range files {
		// 1024 is default max files open for linux
		if i != 0 && i%1023 == 0 {
			chunk = append(chunk, fileData)
			chunks = append(chunks, chunk)

			var newChunk []FileData

			chunk = newChunk
		} else {
			chunk = append(chunk, fileData)
		}
	}

	if len(chunk) < 1025 {
		chunks = append(chunks, chunk)
	}

	return chunks
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
