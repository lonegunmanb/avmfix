package pkg

import (
	"github.com/gobwas/glob"
	"github.com/spf13/afero"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var Fs = afero.NewOsFs()

func DirectoryAutoFix(dirPath string, excludePattern ...string) error {
	pattern := ""
	if len(excludePattern) > 0 {
		pattern = excludePattern[0]
	}
	d := newDirectory(dirPath, pattern)
	if err := d.ensureModules(); err != nil {
		return err
	}
	// variables and outputs files might move blocks into main.tf without fix, so we need run AutoFix twice
	for i := 0; i < 2; i++ {
		if err := d.AutoFix(); err != nil {
			return err
		}
	}
	return nil
}

type fileMode interface {
	Mode() os.FileMode
}

type directory struct {
	path        string
	excludeGlob glob.Glob
	tfFiles     map[string]*HclFile
	dirEntries  map[string]fileMode
}

func (d *directory) AutoFix() error {
	if err := d.loadTfFiles(); err != nil {
		return err
	}
	// Use clone here since d.tfFile might be changed during AutoFix, while the content hasn't been updated.
	for _, hclFile := range maps.Clone(d.tfFiles) {
		if err := hclFile.AutoFix(); err != nil {
			return err
		}

		if err := d.writeFileToDisk(hclFile); err != nil {
			return err
		}
	}
	return nil
}

func (d *directory) AppendBlockToFile(destFileName string, block *HclBlock) {
	if err := d.ensureDestFile(destFileName); err != nil {
		return
	}

	hclFile := d.tfFiles[destFileName]

	if !hclFile.endWithNewLine() {
		hclFile.appendNewline()
	}
	hclFile.appendBlock(block)
	_ = d.writeFileToDisk(hclFile)
}

func (d *directory) writeFileToDisk(hclFile *HclFile) error {
	baseName := filepath.Base(hclFile.FileName)
	mode := d.dirEntries[baseName].Mode()
	err := afero.WriteFile(Fs, hclFile.FileName, hclFile.WriteFile.Bytes(), mode)
	if err != nil {
		return err
	}
	return nil
}

func (d *directory) loadTfFiles() error {
	files, err := afero.ReadDir(Fs, d.path)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tf") {
			// Check if file should be excluded
			if d.shouldExclude(file.Name()) {
				continue
			}
			
			path := filepath.Join(d.path, file.Name())
			content, err := afero.ReadFile(Fs, path)
			if err != nil {
				return err
			}

			hclFile, diags := ParseConfig(content, path)
			if diags.HasErrors() {
				return diags
			}
			hclFile.dir = d
			d.tfFiles[file.Name()] = hclFile
			d.dirEntries[file.Name()] = file
		}
	}
	return nil
}

func (d *directory) ensureDestFile(destFileName string) error {
	destFilePath := filepath.Join(d.path, destFileName)
	exist, err := afero.Exists(Fs, destFilePath)
	if err != nil {
		return err
	}
	if !exist {
		file, err := Fs.Create(destFilePath)
		if err != nil {
			// handle error
			return err
		}
		defer func() {
			_ = file.Close()
		}()
		d.tfFiles[destFileName], _ = ParseConfig([]byte(""), destFilePath)
		fi, err := file.Stat()
		if err != nil {
			return err
		}
		d.dirEntries[destFileName] = fi
	}
	return nil
}

var terraformGetFunc = func(path string) error {
	initCmd := exec.Command("terraform", "get")
	initCmd.Dir = path
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	return initCmd.Run()
}

func (d *directory) shouldExclude(fileName string) bool {
	if d.excludeGlob == nil {
		return false
	}
	return d.excludeGlob.Match(fileName)
}

func (d *directory) ensureModules() error {
	return terraformGetFunc(d.path)
}

func newDirectory(path, excludePattern string) *directory {
	var excludeGlob glob.Glob
	if excludePattern != "" {
		excludeGlob = glob.MustCompile(excludePattern)
	}
	return &directory{
		path:        path,
		excludeGlob: excludeGlob,
		tfFiles:     make(map[string]*HclFile),
		dirEntries:  make(map[string]fileMode),
	}
}
