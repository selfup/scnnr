package scnnr_test

import (
	"os"

	scnnr "github.com/selfup/scnnr/pkg"
)

// These assignments fail to compile if a refactor changes an existing public
// method or constructor signature. In particular, finder Scan has no result.
type originalFinder interface{ Scan(string) }
type originalScanner interface{ Scan() error }

var (
	_ originalFinder                                      = (*scnnr.FileNameFinder)(nil)
	_ originalFinder                                      = (*scnnr.FileSizeFinder)(nil)
	_ originalFinder                                      = (*scnnr.FileFingerprintFinder)(nil)
	_ originalScanner                                     = (*scnnr.Scanner)(nil)
	_ func(string)                                        = (*scnnr.FileNameFinder)(nil).Scan
	_ func(string)                                        = (*scnnr.FileSizeFinder)(nil).Scan
	_ func(string)                                        = (*scnnr.FileFingerprintFinder)(nil).Scan
	_ func([]string) *scnnr.FileNameFinder                = scnnr.NewFileNameFinder
	_ func(string) *scnnr.FileSizeFinder                  = scnnr.NewFileSizeFinder
	_ func([]string) *scnnr.FileFingerprintFinder         = scnnr.NewFileFingerprintFinder
	_ func(string, string, os.FileInfo) string            = scnnr.FullFilePath
	_ func() string                                       = scnnr.PathDirection
	_ func(string)                                        = scnnr.CheckDirOrPanic
	_ func(string, string) ([]os.FileInfo, []os.FileInfo) = scnnr.CollectFilesAndDirs
)
