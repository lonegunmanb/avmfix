package pkg

import (
	"os"
	"path/filepath"
	"strings"
)

func DirectoryAutoFix(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tf") {
			path := filepath.Join(dirPath, file.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			hclFile, diags := ParseConfig(content, path)
			if diags.HasErrors() {
				return diags
			}

			hclFile.AutoFix()

			fileInfo, err := file.Info()
			if err != nil {
				return err
			}

			err = os.WriteFile(path, hclFile.WriteFile.Bytes(), fileInfo.Mode())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
