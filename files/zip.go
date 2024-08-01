package files

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnzipArchive Unzips an archive to an output path
func UnzipArchive(zipPath string, outputPath string) error {
	archive, err := zip.OpenReader(zipPath)

	if err != nil {
		return err
	}

	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(outputPath, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(outputPath)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path for: %v", filePath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())

		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()

		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	return nil
}
