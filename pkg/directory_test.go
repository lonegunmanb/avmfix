package pkg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lonegunmanb/avmfix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FileAutoFix(t *testing.T) {
	temp, err := os.MkdirTemp("", "autofix*")
	require.NoError(t, err)
	file, err := os.Create(filepath.Join(temp, "main.tf"))
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(temp)
	}()
	_, err = file.WriteString(`resource "azurerm_container_group" "example" {
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
}`)
	require.NoError(t, err)
	err = pkg.DirectoryAutoFix(temp)
	require.NoError(t, err)
	fixedFile, err := os.ReadFile(filepath.Join(temp, "main.tf"))
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
	temp, err := os.MkdirTemp("", "autofix*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(temp)
	}()
	mainFilePath := filepath.Join(temp, "main.tf")
	_, err = os.Create(mainFilePath)
	require.NoError(t, err)
	variablesPath := filepath.Join(temp, "variables.tf")
	variablesFile, err := os.Create(variablesPath)
	require.NoError(t, err)
	_, err = variablesFile.WriteString(`locals {
}

variable "test" {}`)
	require.NoError(t, err)
	err = pkg.DirectoryAutoFix(temp)
	require.NoError(t, err)
	mainContent, err := os.ReadFile(mainFilePath)
	require.NoError(t, err)
	expectMain := `locals {
}
`
	assert.Equal(t, formatHcl(expectMain), formatHcl(string(mainContent)))
	expectVariable := `variable "test" {
}
`
	variablesContent, err := os.ReadFile(variablesPath)
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectVariable), formatHcl(string(variablesContent)))
}

func TestNonVariableBlockInVariablesDotTfFileShouldBeMovedIntoMainDotTfFileAndFixed(t *testing.T) {
	temp, err := os.MkdirTemp("", "autofix*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(temp)
	}()
	mainFilePath := filepath.Join(temp, "main.tf")
	_, err = os.Create(mainFilePath)
	require.NoError(t, err)
	variablesPath := filepath.Join(temp, "variables.tf")
	variablesFile, err := os.Create(variablesPath)
	require.NoError(t, err)
	_, err = variablesFile.WriteString(`locals {
  b = "b"
  a = "a"
}

variable "test" {}`)
	require.NoError(t, err)
	err = pkg.DirectoryAutoFix(temp)
	require.NoError(t, err)
	mainContent, err := os.ReadFile(mainFilePath)
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
	variablesContent, err := os.ReadFile(variablesPath)
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectVariable), formatHcl(string(variablesContent)))
}

func TestNonVariableBlockInVariablesDotTfFileShouldBeMovedIntoNewCreatedMainDotTf(t *testing.T) {
	temp, err := os.MkdirTemp("", "autofix*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(temp)
	}()
	variablesPath := filepath.Join(temp, "variables.tf")
	variablesFile, err := os.Create(variablesPath)
	require.NoError(t, err)
	_, err = variablesFile.WriteString(`locals {
}

variable "test" {}`)
	require.NoError(t, err)
	err = pkg.DirectoryAutoFix(temp)
	require.NoError(t, err)
	mainFilePath := filepath.Join(temp, "main.tf")
	mainContent, err := os.ReadFile(mainFilePath)
	require.NoError(t, err)
	expectMain := `locals {
}
`
	assert.Equal(t, formatHcl(expectMain), formatHcl(string(mainContent)))
	expectVariable := `variable "test" {
}
`
	variablesContent, err := os.ReadFile(variablesPath)
	require.NoError(t, err)
	assert.Equal(t, formatHcl(expectVariable), formatHcl(string(variablesContent)))
}
