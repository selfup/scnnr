package scnnr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type contentMatcher struct {
	keywords  []string
	patterns  []*regexp.Regexp
	regex     bool
	positions bool
}

func newContentMatcher(keywords []string, useRegex, positions bool) (*contentMatcher, error) {
	matcher := &contentMatcher{keywords: nonEmpty(keywords), regex: useRegex, positions: positions}
	if useRegex {
		for _, keyword := range matcher.keywords {
			pattern, err := regexp.Compile(keyword)

			if err != nil {
				return nil, fmt.Errorf("invalid regex %q: %w", keyword, err)
			}

			matcher.patterns = append(matcher.patterns, pattern)
		}
	}

	return matcher, nil
}

// parseLegacy is the reader seam for the original Scan behavior. It reports the
// first occurrence of each keyword on each line, keeps reading after a match,
// retains empty keywords, and compiles regexes only when a line is examined.
func (m *contentMatcher) parseLegacy(r io.Reader, path string, size int64, record func(Match)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 2*int(size))
	found := false
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		for _, keyword := range m.keywords {
			if !m.positions && found {
				break
			}

			column := -1

			if m.regex {
				if location := regexp.MustCompile(keyword).FindStringIndex(line); location != nil {
					column = location[0]
				}
			} else {
				column = strings.Index(line, keyword)
			}

			if column >= 0 {
				record(Match{Path: path, Line: lineNumber, Column: column + 1, Keyword: keyword})

				if !m.positions {
					found = true
					break
				}
			}
		}
	}
	return scanner.Err()
}

// parse uses a buffered reader so line length does not depend on file metadata
// or bufio.Scanner's token limit. Without positions, the first match ends a read.
func (m *contentMatcher) parse(r io.Reader, path string) ([]Match, error) {
	reader := bufio.NewReader(r)

	var matches []Match

	for lineNumber := 1; ; lineNumber++ {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		if len(line) > 0 {
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")

			for i, keyword := range m.keywords {
				var columns []int

				if len(m.patterns) > 0 {
					limit := 1

					if m.positions {
						limit = -1
					}

					for _, location := range m.patterns[i].FindAllStringIndex(line, limit) {
						columns = append(columns, location[0]+1)
					}
				} else {
					for offset := 0; offset < len(line); {
						index := strings.Index(line[offset:], keyword)

						if index < 0 {
							break
						}

						columns = append(columns, offset+index+1)
						offset += index + len(keyword)

						if !m.positions {
							break
						}
					}
				}
				for _, column := range columns {
					matches = append(matches, Match{Path: path, Line: lineNumber, Column: column, Keyword: keyword})

					if !m.positions {
						return matches, nil
					}
				}
			}
		}

		if errors.Is(err, io.EOF) {
			return matches, nil
		}
	}
}

func nonEmpty(values []string) []string {
	var result []string

	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}

	return result
}
