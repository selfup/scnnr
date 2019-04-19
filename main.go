package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	scanner := new(Scanner)

	scanner.FilePatterns = []string{".html", ".yml", ".js"}
	scanner.Keywords = []string{"initial-scale=1", "Cache: ", "cache: ", "Cache=", "cache="}

	scanner.Scan()
}

// Scanner scans files and based on pattern stores in array for goroutine processing
type Scanner struct {
	sync.Mutex
	MatchedFilePaths []string
	FilePatterns     []string
	Keywords         []string
	KeywordMatches   []string
}

// Scan walks the given directory tree and stores all matching files into a slice
func (s *Scanner) Scan() error {
	walkingError := filepath.Walk(".", s.scan)

	if walkingError != nil {
		log.Fatal(walkingError)
		return walkingError
	}

	for _, match := range s.MatchedFilePaths {
		s.parse(match)
	}

	log.Println(s.KeywordMatches)

	return nil
}

func (s *Scanner) scan(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	for _, pattern := range s.FilePatterns {
		if strings.Contains(path, pattern) {
			s.MatchedFilePaths = append(s.MatchedFilePaths, path)
		}
	}

	return nil
}

func (s *Scanner) parse(match string) {
	file, err := os.Open(match)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)

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
