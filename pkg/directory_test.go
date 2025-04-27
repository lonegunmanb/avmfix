package pkg_test

import (
	"testing"

	"github.com/lonegunmanb/avmfix/pkg"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FileAutoFix(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"main.tf": `resource "azurerm_container_group" "example" {
  name                = "example-continst"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  ip_address_type     = "Public"
  dns_name_label      = "aci-label"
  os_type             = "Linux"

  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }

  tags = {
    environment = "testing"
  }
}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()
	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)
	fixedFile, err := afero.ReadFile(mockFs, "main.tf")
	require.NoError(t, err)
	expected := `resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location
  name                = "example-continst"
  os_type             = "Linux"
  resource_group_name = azurerm_resource_group.example.name
  dns_name_label      = "aci-label"
  ip_address_type     = "Public"
  tags = {
    environment = "testing"
  }

  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }
}`
	assert.Equal(t, formatHcl(expected), formatHcl(string(fixedFile)))
}

func TestNonVariableBlockInVariablesDotTfFileShouldBeMovedIntoMainDotTfFile(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"main.tf": "",
		"variables.tf": `locals {
}

variable "test" {}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()
	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)
	mainContent, err := afero.ReadFile(mockFs, "main.tf")
	require.NoError(t, err)
	expectMain := `locals {
}
`
	assert.Equal(t, formatHcl(expectMain), formatHcl(string(mainContent)))
	expectVariable := `variable "test" {
}
`
	variablesContent, err := afero.ReadFile(mockFs, "variables.tf")
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectVariable), formatHcl(string(variablesContent)))
}

func TestNonVariableBlockInVariablesDotTfFileShouldBeMovedIntoMainDotTfFileAndFixed(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"main.tf": "",
		"variables.tf": `locals {
  b = "b"
  a = "a"
}

variable "test" {}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()
	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)
	mainContent, err := afero.ReadFile(mockFs, "main.tf")
	require.NoError(t, err)
	expectMain := `locals {
  a = "a"
  b = "b"
}
`
	assert.Equal(t, formatHcl(expectMain), formatHcl(string(mainContent)))
	expectVariable := `variable "test" {
}
`
	variablesContent, err := afero.ReadFile(mockFs, "variables.tf")
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectVariable), formatHcl(string(variablesContent)))
}

func TestNonVariableBlockInVariablesDotTfFileShouldBeMovedIntoNewCreatedMainDotTf(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"variables.tf": `locals {
}

variable "test" {}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()
	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)
	mainContent, err := afero.ReadFile(mockFs, "main.tf")
	require.NoError(t, err)
	expectMain := `locals {
}
`
	assert.Equal(t, formatHcl(expectMain), formatHcl(string(mainContent)))
	expectVariable := `variable "test" {
}
`
	variablesContent, err := afero.ReadFile(mockFs, "variables.tf")
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectVariable), formatHcl(string(variablesContent)))
}

func TestNonOutputBlockInOutputsDotTfFileShouldBeMovedIntoMainDotTfFile(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"main.tf": "",
		"outputs.tf": `locals {
}

output "test" {}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()
	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)
	mainContent, err := afero.ReadFile(mockFs, "main.tf")
	require.NoError(t, err)
	expectMain := `locals {
}
`
	assert.Equal(t, formatHcl(expectMain), formatHcl(string(mainContent)))
	expectOutput := `output "test" {
}
`
	outputsContent, err := afero.ReadFile(mockFs, "outputs.tf")
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectOutput), formatHcl(string(outputsContent)))
}

func TestNonOutputBlockInOutputsDotTfFileShouldBeMovedIntoMainDotTfFileAndFixed(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"main.tf": "",
		"outputs.tf": `locals {
  b = "b"
  a = "a"
}

output "test" {}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()
	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)
	mainContent, err := afero.ReadFile(mockFs, "main.tf")
	require.NoError(t, err)
	expectMain := `locals {
  a = "a"
  b = "b"
}
`
	assert.Equal(t, formatHcl(expectMain), formatHcl(string(mainContent)))
	expectOutput := `output "test" {
}
`
	outputsContent, err := afero.ReadFile(mockFs, "outputs.tf")
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectOutput), formatHcl(string(outputsContent)))
}

func TestNonOutputBlockInOutputsDotTfFileShouldBeMovedIntoNewCreatedMainDotTf(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"outputs.tf": `locals {
}

output "test" {}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()
	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)
	mainContent, err := afero.ReadFile(mockFs, "main.tf")
	require.NoError(t, err)
	expectMain := `locals {
}
`
	assert.Equal(t, formatHcl(expectMain), formatHcl(string(mainContent)))
	expectOutput := `output "test" {
}
`
	outputsContent, err := afero.ReadFile(mockFs, "outputs.tf")
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectOutput), formatHcl(string(outputsContent)))
}

func TestVariableBlockInMainDotTfFileShouldBeMovedIntoVariableDotTf(t *testing.T) {
	variableBlock := `variable "test" {}`
	mockFs := fakeFs(map[string]string{
		"main.tf": variableBlock,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()
	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)
	file, err := afero.ReadFile(mockFs, "variables.tf")
	require.NoError(t, err)
	assert.Equal(t, `variable "test" {
}
`, formatHcl(string(file)))
}

func TestOutputBlockInMainDotTfFileShouldBeMovedIntoOutputsDotTf(t *testing.T) {
	outputBlock := `output "example_output" {}`
	mockFs := fakeFs(map[string]string{
		"main.tf": outputBlock,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()

	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)

	file, err := afero.ReadFile(mockFs, "outputs.tf")
	require.NoError(t, err)
	assert.Equal(t, `output "example_output" {
}
`, formatHcl(string(file)))
}

func TestVariableBlockInVariablesDotTfFileShouldNotBeMoved(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"variables.tf":            `variable "test" {}`,
		"deprecated_variables.tf": `variable "test2" {}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()

	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)

	variablesContent, err := afero.ReadFile(mockFs, "variables.tf")
	require.NoError(t, err)
	assert.Equal(t, `variable "test" {
}
`, formatHcl(string(variablesContent)))

	variablesContent, err = afero.ReadFile(mockFs, "deprecated_variables.tf")
	require.NoError(t, err)
	assert.Equal(t, `variable "test2" {
}
`, formatHcl(string(variablesContent)))

	mainExist, err := afero.Exists(mockFs, "main.tf")
	require.NoError(t, err)
	assert.False(t, mainExist)
}

func TestOutputBlockInOutputsDotTfFileShouldNotBeMoved(t *testing.T) {
	mockFs := fakeFs(map[string]string{
		"outputs.tf":            `output "example_output" {}`,
		"deprecated_outputs.tf": `output "example_output2" {}`,
	})
	stub := gostub.Stub(&pkg.Fs, mockFs)
	defer stub.Reset()

	err := pkg.DirectoryAutoFix("")
	require.NoError(t, err)

	outputsContent, err := afero.ReadFile(mockFs, "outputs.tf")
	require.NoError(t, err)
	assert.Equal(t, `output "example_output" {
}
`, formatHcl(string(outputsContent)))

	outputsContent, err = afero.ReadFile(mockFs, "deprecated_outputs.tf")
	require.NoError(t, err)
	assert.Equal(t, `output "example_output2" {
}
`, formatHcl(string(outputsContent)))

	exists, err := afero.Exists(mockFs, "main.tf")
	require.NoError(t, err)
	assert.False(t, exists)
}

func fakeFs(files map[string]string) afero.Fs {
	fs := afero.NewMemMapFs()
	for path, content := range files {
		_ = afero.WriteFile(fs, path, []byte(content), 0644)
	}
	return fs
}
