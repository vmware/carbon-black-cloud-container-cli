package layers

import (
	"archive/tar"
	"crypto/sha256"
	"fmt"
	"io"
)

var ELFStart = [4]byte{127, 69, 76, 70} // this is { 0x7f, 'E', 'L', 'F' }

type ExecutableFile struct {
	Digest string `json:"digest"` // the file's SHA26
	Path   string `json:"path"`
	Size   uint64  `json:"size"`
}

// ReadExecutableFromTar creates an ExecutableFile object if the file is an ELF
func ReadExecutableFromTar(reader *tar.Reader, header *tar.Header, path string) (*ExecutableFile, error) {
	if header.Typeflag == tar.TypeDir {
		return nil, nil
	}

	var hash string
	isELF := false

	isELF, hash, err := calculateELFMetadata(reader)
	if !isELF { // returns nil if the file is not ELF
		return nil, err
	}

	return &ExecutableFile{
		Digest: hash,
		Path:   path,
		Size:   uint64(header.Size),
	}, nil
}

// calculateELFMetadata returns whether the file is an ELF and if so, returns its SHA256
func calculateELFMetadata(reader io.Reader) (bool, string, error) {
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
