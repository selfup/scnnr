package cmd

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Scanner scans files and based on pattern stores in array for goroutine processing
type Scanner struct {
	Matched  []string
	Patterns []string
	Keywords []string
}

// Scan walks the given directory tree and stores all matching files into a slice
func (s *Scanner) Scan() error {
	walkingError := filepath.Walk(".", s.scan)

	if walkingError != nil {
		log.Fatal(walkingError)
		return walkingError
	}

	parser := new(Parser)
	parser.Keywords = s.Keywords

	for _, match := range s.Matched {
		parser.Parse(match)
	}

	log.Println(parser.KeywordMatches)

	return nil
}

func (s *Scanner) scan(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	for _, pattern := range s.Patterns {
		if strings.Contains(path, pattern) {
			s.Matched = append(s.Matched, path)
		}
	}

	return nil
}

// Parser is a Mutexed struct for concurrent buffio
type Parser struct {
	sync.Mutex
	Keywords       []string
	KeywordMatches []string
}

// Parse will parse each file with a new buffer and try to find a match based on certain keywords
func (p *Parser) Parse(match string) {
	file, err := os.Open(match)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		for i := 0; i < len(p.Keywords); i++ {
			if contains(p.KeywordMatches, match) {
				break
			}

			if strings.Contains(line, p.Keywords[i]) {
				p.Lock()
				p.KeywordMatches = append(p.KeywordMatches, match)
				p.Unlock()
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
