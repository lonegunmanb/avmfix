package pkg

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func unzip(source, destination string) error {
	r, err := zip.OpenReader(source)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer func() {
		_ = r.Close()
	}()

	for _, f := range r.File {
		if err := unzipFile(f, destination); err != nil {
			return fmt.Errorf("failed to extract file from zip: %w", err)
		}
	}

	return nil
}

func unzipFile(f *zip.File, destination string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	path := filepath.Join(destination, f.Name)
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(path, f.Mode()); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		return nil
	}

	outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	if _, err := io.Copy(outFile, rc); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
