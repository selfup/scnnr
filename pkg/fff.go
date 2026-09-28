package scnnr

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/selfup/scnnr/internal/digest"
)

// FileFingerprintFinder matches SHA-256 digests, including digest substrings.
type FileFingerprintFinder struct {
	mutex        sync.Mutex
	Direction    string
	Files        []string
	FingerPrints []string
}

func NewFileFingerprintFinder(fingerPrints []string) *FileFingerprintFinder {
	return &FileFingerprintFinder{Direction: PathDirection(), FingerPrints: fingerPrints}
}

// Scan preserves the original void API and best-effort reads. Read failures do
// not stop the scan; the bytes returned by ReadFile are hashed, as before.
func (f *FileFingerprintFinder) Scan(directory string) {
	CheckDirOrPanic(directory)

	f.findFiles(directory)
}

func (f *FileFingerprintFinder) findFiles(directory string) {
	walkFinderFiles(directory, f.Direction, func(directory string, file os.FileInfo) {
		path := FullFilePath(directory, f.Direction, file)
		data, _ := os.ReadFile(path)
		sum, _ := digest.SHA256(bytes.NewReader(data))

		for _, fingerprint := range f.FingerPrints {
			if strings.Contains(sum, fingerprint) {
				f.mutex.Lock()
				f.Files = append(f.Files, path)
				f.mutex.Unlock()
			}
		}
	})
}

func fileDigest(path string) (sum string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %q: %w", path, err)
	}

	defer func() { err = errors.Join(err, file.Close()) }()

	sum, err = digest.SHA256(file)
	if err != nil {
		err = fmt.Errorf("hash %q: %w", path, err)
	}

	return sum, err
}
