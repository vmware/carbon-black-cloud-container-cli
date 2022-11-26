package layers

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// Keep below definitions in sync with whatever is on used on the control plane for consistent conversions.
const (
	// CategoryElf is a linux executable
	CategoryElf FileCategory = "ELF"
)

type FileCategory string

var ELFStart = [4]byte{127, 69, 76, 70} // this is { 0x7f, 'E', 'L', 'F' }

type ExecutableFile struct {
	Digest          string       `json:"digest"` // the file's SHA256
	Path            string       `json:"path"`
	Size            uint64       `json:"size"`
	Category        FileCategory `json:"file_category"`
	InSquashedImage bool         `json:"in_squashed_image"`
}

// CalculateELFMetadata returns whether the file is an ELF and if so, returns its SHA256
func CalculateELFMetadata(reader io.Reader) (bool, string, error) {
	start := make([]byte, 4)
	_, err := reader.Read(start)
	if err != nil && err != io.EOF {
		return false, "", err
	}

	isELF := true
	for i := 0; i < 4; i++ {
		if start[i] != ELFStart[i] {
			isELF = false
			return false, "", nil // if it's not ELF we will not continue to calculate the hash
		}
	}

	hash := sha256.New()
	hash.Write(start)
	_, err = io.Copy(hash, reader)
	if err != nil {
		return false, "", err
	}

	return isELF, fmt.Sprintf("%x", hash.Sum(nil)), nil
}
