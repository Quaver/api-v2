package files

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// GetFileMD5 Gets the MD% hash of a given file
func GetFileMD5(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	file, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer func(file *os.File) {
		err := file.Close()

		if err != nil {
			return
		}
	}(file)

	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
