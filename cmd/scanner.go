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
	MatchedFilePaths []string
	FilePatterns     []string
	Keywords         []string
	KeywordMatches   []string
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
		go func(m string) {
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
			if strings.Contains(filepath.Ext(path), pattern) {
				s.MatchedFilePaths = append(s.MatchedFilePaths, path)
			}
		}
	}

	return nil
}

func (s *Scanner) parse(match string) {
	file, err := os.Open(match)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*3096)

	for scanner.Scan() {
		line := scanner.Text()

		for i := 0; i < len(s.Keywords); i++ {
			if contains(s.KeywordMatches, match) {
				break
			}

			if strings.Contains(line, s.Keywords[i]) {
				s.Lock()
				s.KeywordMatches = append(s.KeywordMatches, match)
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
