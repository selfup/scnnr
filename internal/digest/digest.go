// Package digest provides the streaming checksum shared by finder and release
// commands.
package digest

import (
	"crypto/sha256"
	"fmt"
	"io"
)

func SHA256(r io.Reader) (string, error) {
	hash := sha256.New()

	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
