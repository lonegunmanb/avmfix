package pkg

import (
	"os"
	"path/filepath"
	"strings"
)

func DirectoryAutoFix(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".tf") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			hclFile, diags := ParseConfig(content, path)
			if diags.HasErrors() {
				return diags
			}

			hclFile.AutoFix()

			err = os.WriteFile(path, hclFile.WriteFile.Bytes(), info.Mode())
			if err != nil {
				return err
			}
		}
		return nil
	})
}
