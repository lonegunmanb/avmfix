package pkg_test

import (
	"testing"

	"github.com/lonegunmanb/azure-verified-module-fix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildResourceGroup_ArgumentsOnly(t *testing.T) {
	code := `
resource "azurerm_resource_group" "example" {
  name     = "example"
  location = "West Europe"
  tags     = {
	environment = "Production"
  }
}`
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	assert.Equal(t, "example", resourceBlock.Name)
	assert.Equal(t, "azurerm_resource_group", resourceBlock.Type)
	assert.Equal(t, 2, len(resourceBlock.RequiredArgs.Args))
	assert.Equal(t, "name", resourceBlock.RequiredArgs.Args[0].Name)
	assert.Equal(t, "location", resourceBlock.RequiredArgs.Args[1].Name)
	assert.Equal(t, 1, len(resourceBlock.OptionalArgs.Args))
	assert.Equal(t, "tags", resourceBlock.OptionalArgs.Args[0].Name)
}

func TestBuildResourceGroup_MetaArguments(t *testing.T) {
	code := `
resource "azurerm_resource_group" "example" {
  count    = var.create_resource_group ? 1 : 0
  provider = azurerm.example

  name     = "example"
  location = "West Europe"
  tags     = {
	environment = "Production"
  }

  depends_on = [var.depends_on]

  lifecycle {
    create_before_destroy = false
	prevent_destroy 	  = false
	ignore_changes 		  = [
		tags,
	]
	replace_triggered_by = [
		"null_resource.trigger",
	]
  }
}`
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	assert.Equal(t, 2, len(resourceBlock.HeadMetaArgs.Args))
	assert.Equal(t, "count", resourceBlock.HeadMetaArgs.Args[0].Name)
	assert.Equal(t, "provider", resourceBlock.HeadMetaArgs.Args[1].Name)
	assert.Equal(t, "depends_on", resourceBlock.TailMetaArgs.Args[0].Name)
	assert.Equal(t, 1, len(resourceBlock.TailMetaNestedBlocks.Blocks))
	lifecycleBlock := resourceBlock.TailMetaNestedBlocks.Blocks[0]
	assert.Equal(t, "lifecycle", lifecycleBlock.Name)
	assert.Nil(t, lifecycleBlock.RequiredArgs)
	assert.Equal(t, 4, len(lifecycleBlock.OptionalArgs.Args))
	assert.Equal(t, "create_before_destroy", lifecycleBlock.OptionalArgs.Args[0].Name)
	assert.Equal(t, "prevent_destroy", lifecycleBlock.OptionalArgs.Args[1].Name)
	assert.Equal(t, "ignore_changes", lifecycleBlock.OptionalArgs.Args[2].Name)
	assert.Equal(t, "replace_triggered_by", lifecycleBlock.OptionalArgs.Args[3].Name)
}

func TestResourceBlockAutoFix(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
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
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_container_group" "example" {
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
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_AutoFix_DynamicNestedBlock(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location

  dynamic "dns_config" {
	for_each = var.nameservers == null ? ["dns_config"] : []

	content {
    	nameservers = var.nameservers
	}
  }
  container {
	cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"
	memory_limit = 1.5

	dynamic "gpu_limit" {
      for_each = var.gpu_limit_enabled ? ["1"] : []

		content {
			sku = "K80"
			count = 1
		}
	}
  }
}
`
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location

  container {
	cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"
	memory_limit = 1.5

	dynamic "gpu_limit" {
      for_each = var.gpu_limit_enabled ? ["1"] : []

		content {
			count = 1
			sku = "K80"
		}
	}
  }
  dynamic "dns_config" {
	for_each = var.nameservers == null ? ["dns_config"] : []

	content {
    	nameservers = var.nameservers
	}
  }
}
`
	s := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(s))
}

func TestResourceBlockAutoFix_CommentsShouldBePreserved(t *testing.T) {
	code := `
# Multi-Line Comments
#   Here
resource "azurerm_container_group" "example" {
  name                = "example-continst"
  os_type             = "Linux"
  resource_group_name = azurerm_resource_group.example.name
  ip_address_type     = "Public"
  # Nested Block Comment
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
	# Multi-Line
	#   Nested Block Comment
    ports {
      port     = 443
      protocol = "TCP"
    }
    name   = "hello-world"
  }

  # Argument comment dns_name_label
  dns_name_label      = "aci-label"
  #Argument Comment location
  location            = azurerm_resource_group.example.location
  
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }
  # tags
  tags = {
    environment = "testing"
  }
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	resourceBlock.AutoFix()
	expected := `
# Multi-Line Comments
#   Here
resource "azurerm_container_group" "example" {
  #Argument Comment location
  location            = azurerm_resource_group.example.location
  name                = "example-continst"
  os_type             = "Linux"
  resource_group_name = azurerm_resource_group.example.name
  # Argument comment dns_name_label
  dns_name_label      = "aci-label"
  ip_address_type     = "Public"
  # tags
  tags = {
    environment = "testing"
  }
  
  # Nested Block Comment
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"

	# Multi-Line
	#   Nested Block Comment
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
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlockAutoFix_Datasource(t *testing.T) {
	code := `
data "azurerm_virtual_network" "example" {
  resource_group_name = "networking"
  name                = "production"
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	resourceBlock.AutoFix()
	expected := `
data "azurerm_virtual_network" "example" {
  name                = "production"
  resource_group_name = "networking"
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}
