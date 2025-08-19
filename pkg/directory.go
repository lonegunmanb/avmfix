package pkg

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
)

var Fs = afero.NewOsFs()

func DirectoryAutoFix(dirPath string) error {
	d := newDirectory(dirPath)
	if err := d.ensureModules(); err != nil {
		return err
	}
	if err := d.parseTerraformLockFile(); err != nil {
		return fmt.Errorf("failed to parse .terraform.lock.hcl: %w", err)
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
	path             string
	tfFiles          map[string]*HclFile
	dirEntries       map[string]fileMode
	providerVersions map[string]map[string]string
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

var terraformInitFunc = func(path string) error {
	initCmd := exec.Command("terraform", "init", "-backend=false")
	initCmd.Dir = path
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	return initCmd.Run()
}

func (d *directory) ensureModules() error {
	return terraformInitFunc(d.path)
}

func newDirectory(path string) *directory {
	return &directory{
		path:       path,
		tfFiles:    make(map[string]*HclFile),
		dirEntries: make(map[string]fileMode),
	}
}

// parseTerraformLockFile parses the .terraform.lock.hcl file in the given directory
// and returns a nested map structure: namespace -> provider name -> version.
// Example: "hashicorp" -> "azurerm" -> "4.37.0"
func (d *directory) parseTerraformLockFile() error {
	lockFilePath := filepath.Join(d.path, ".terraform.lock.hcl")
	var err error
	d.providerVersions, err = parseTerraformLockFile(lockFilePath)
	return err
}

func parseTerraformLockFileStub(lockFilePath string) (map[string]map[string]string, error) {
	// Check if the lock file exists
	exists, err := afero.Exists(Fs, lockFilePath)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Return empty map if lock file doesn't exist
		return nil, fmt.Errorf("lock file %s does not exist", lockFilePath)
	}

	// Read the lock file content
	content, err := afero.ReadFile(Fs, lockFilePath)
	if err != nil {
		return nil, err
	}

	// Parse the HCL content
	file, diags := hclsyntax.ParseConfig(content, lockFilePath, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, diags
	}

	providerVersions := make(map[string]map[string]string)

	// Iterate through the body to find provider blocks
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("failed to parse `.terraform.lock.hcl` body")
	}

	for _, block := range body.Blocks {
		if block.Type != "provider" || len(block.Labels) == 0 {
			continue
		}

		fullProviderName := block.Labels[0]
		// Remove quotes from the provider name if present
		fullProviderName = strings.Trim(fullProviderName, `"`)

		// Parse the provider name to extract namespace and provider name
		// Format: "registry.terraform.io/namespace/provider"
		parts := strings.Split(fullProviderName, "/")
		if len(parts) < 3 {
			continue
		}

		namespace := parts[len(parts)-2]    // e.g., "hashicorp"
		providerName := parts[len(parts)-1] // e.g., "azurerm"

		// Look for the version attribute in the block
		versionAttr, exists := block.Body.Attributes["version"]
		if !exists {
			continue
		}

		var version string
		if literal, ok := versionAttr.Expr.(*hclsyntax.LiteralValueExpr); ok {
			version = literal.Val.AsString()
		} else if template, ok := versionAttr.Expr.(*hclsyntax.TemplateExpr); ok {
			// Handle template expressions (quoted strings)
			if len(template.Parts) == 1 {
				if literal, ok := template.Parts[0].(*hclsyntax.LiteralValueExpr); ok {
					version = literal.Val.AsString()
				}
			}
		}

		if version == "" {
			continue
		}

		// Initialize namespace map if it doesn't exist
		if providerVersions[namespace] == nil {
			providerVersions[namespace] = make(map[string]string)
		}

		providerVersions[namespace][providerName] = version
	}

	return providerVersions, nil
}

var parseTerraformLockFile = parseTerraformLockFileStub

func (d *directory) resolveNamespace(resourceType string) (string, error) {
	providerType := providerName(resourceType)
	for space, providers := range d.providerVersions {
		if _, ok := providers[providerType]; ok {
			return space, nil
		}
	}
	return "", fmt.Errorf("namespace for resource %s not found", resourceType)
}

func providerName(resourceType string) string {
	return strings.Split(resourceType, "_")[0]
}

func (d *directory) resolveProviderVersion(namespace, resourceType string) (string, error) {
	providerType := providerName(resourceType)
	if providers, ok := d.providerVersions[namespace]; ok {
		if version, ok := providers[providerType]; ok {
			return version, nil
		}
	}
	return "", fmt.Errorf("version for resource %s in namespace %s not found", resourceType, namespace)

}
