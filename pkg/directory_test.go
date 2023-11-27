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
