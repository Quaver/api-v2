package files

import (
	"bytes"
	"compress/gzip"
	"io/fs"
	"io/ioutil"
)

// Un-gzips a given file
func ungzipFile(buffer *bytes.Buffer, path string) error {
	reader, err := gzip.NewReader(buffer)

	defer func(reader *gzip.Reader) {
		_ = reader.Close()
	}(reader)

	if err != nil {
		return err
	}

	var unpacked bytes.Buffer
	_, err = unpacked.ReadFrom(reader)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, unpacked.Bytes(), fs.ModePerm)

	if err != nil {
		return err
	}

	return nil
}
