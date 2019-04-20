package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Scanner scans files and based on pattern stores in array for goroutine processing
type Scanner struct {
	sync.Mutex
	Directory        string
	FilePatterns     []string
	Keywords         []string
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

	matchedCount := len(s.MatchedFilePaths)
	wg.Add(matchedCount)

	for _, match := range s.MatchedFilePaths {
		go func(m FileData) {
			defer wg.Done()

			s.parse(m)
		}(match)
	}

	wg.Wait()
	fmt.Println("Files matching pattern(s): ", s.KeywordMatches)

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

func (s *Scanner) parse(match FileData) {
	file, err := os.Open(match.Path)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 0)

	// only extend buffer to file size
	scanner.Buffer(buf, int(match.Info.Size()))

	for scanner.Scan() {
		line := scanner.Text()

		for i := 0; i < len(s.Keywords); i++ {
			// avoid duplicating results when iterating through keywords
			if contains(s.KeywordMatches, match.Path) {
				break
			}

			if strings.Contains(line, s.Keywords[i]) {
				// utilize Mutex while parse gets called as a goroutine
				s.Lock()
				s.KeywordMatches = append(s.KeywordMatches, match.Path)
				s.Unlock()
			}
		}
	}

	check(scanner.Err())
}

func contains(list []string, match string) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == match {
			return true
		}
	}

	return false
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
