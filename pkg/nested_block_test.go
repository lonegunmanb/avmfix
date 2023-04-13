package pkg

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildNestedBlock_OneOptionalNestedBlock(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location
  dns_config {
    nameservers = []
	search_domains = []
  }
}
`
	file, diag := hclsyntax.ParseConfig([]byte(code), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	resourceBlock := BuildResourceBlock(file.Body.(*hclsyntax.Body).Blocks[0], file, func(block Block) error { return nil })
	assert.Equal(t, 1, len(resourceBlock.OptionalNestedBlocks.Blocks))
	assert.Nil(t, resourceBlock.RequiredNestedBlocks)
	dnsConfigBlock := resourceBlock.OptionalNestedBlocks.Blocks[0]
	assert.Equal(t, "dns_config", dnsConfigBlock.Name)
	assert.Equal(t, "dns_config", dnsConfigBlock.SortField)
	assert.Equal(t, file, dnsConfigBlock.File)
	assert.Equal(t, resourceBlock.Block.Body.Blocks[0], dnsConfigBlock.Block)
	assert.Equal(t, 1, len(dnsConfigBlock.RequiredArgs.Args))
	assert.Equal(t, "nameservers", dnsConfigBlock.RequiredArgs.Args[0].Name)
	assert.Equal(t, 1, len(dnsConfigBlock.OptionalArgs.Args))
	assert.Equal(t, "search_domains", dnsConfigBlock.OptionalArgs.Args[0].Name)
	assert.Equal(t, []string{"resource", "azurerm_container_group", "dns_config"}, dnsConfigBlock.Path)
}

func TestBuildNestedBlock_OneRequiredNestedBlock(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location
  container {
    name   = "hello-world"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = "0.5"
    memory = "1.5"
	memory_limit = 1.5
	
	gpu_limit {
		count = 1
		sku = "K80"
	}
  }
}
`
	file, diag := hclsyntax.ParseConfig([]byte(code), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	resourceBlock := BuildResourceBlock(file.Body.(*hclsyntax.Body).Blocks[0], file, func(block Block) error { return nil })
	assert.Equal(t, 1, len(resourceBlock.RequiredNestedBlocks.Blocks))
	assert.Nil(t, resourceBlock.OptionalNestedBlocks)
	containerBlock := resourceBlock.RequiredNestedBlocks.Blocks[0]
	assert.Equal(t, "container", containerBlock.Name)
	assert.Equal(t, "container", containerBlock.SortField)
	assert.Equal(t, 4, len(containerBlock.RequiredArgs.Args))
	assert.Equal(t, "name", containerBlock.RequiredArgs.Args[0].Name)
	assert.Equal(t, "image", containerBlock.RequiredArgs.Args[1].Name)
	assert.Equal(t, "cpu", containerBlock.RequiredArgs.Args[2].Name)
	assert.Equal(t, "memory", containerBlock.RequiredArgs.Args[3].Name)
	assert.Equal(t, 1, len(containerBlock.OptionalArgs.Args))
	assert.Equal(t, "memory_limit", containerBlock.OptionalArgs.Args[0].Name)
	assert.Equal(t, 1, len(containerBlock.OptionalNestedBlocks.Blocks))
	gpuLimitBlock := containerBlock.OptionalNestedBlocks.Blocks[0]
	assert.Equal(t, "gpu_limit", gpuLimitBlock.Name)
	assert.Equal(t, "gpu_limit", gpuLimitBlock.SortField)
	assert.Nil(t, gpuLimitBlock.RequiredArgs)
	assert.Equal(t, 2, len(gpuLimitBlock.OptionalArgs.Args))
	assert.Equal(t, "count", gpuLimitBlock.OptionalArgs.Args[0].Name)
	assert.Equal(t, "sku", gpuLimitBlock.OptionalArgs.Args[1].Name)
}

func TestBuildNestedBlock_DynamicNestedBlock(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location
  dynamic "dns_config" {
	for_each = var.dns_config ? [1] : []

	content {
    	nameservers = []
		search_domains = []
	}
  }
}
`
	file, diag := hclsyntax.ParseConfig([]byte(code), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	resourceBlock := BuildResourceBlock(file.Body.(*hclsyntax.Body).Blocks[0], file, func(block Block) error { return nil })
	assert.Equal(t, 1, len(resourceBlock.OptionalNestedBlocks.Blocks))
	assert.Nil(t, resourceBlock.RequiredNestedBlocks)
	dnsConfigBlock := resourceBlock.OptionalNestedBlocks.Blocks[0]
	assert.Equal(t, "dns_config", dnsConfigBlock.Name)
	assert.Equal(t, "dns_config", dnsConfigBlock.SortField)
	assert.Equal(t, 1, len(dnsConfigBlock.HeadMetaArgs.Args))
	assert.Equal(t, 1, len(dnsConfigBlock.RequiredArgs.Args))
	assert.Equal(t, "nameservers", dnsConfigBlock.RequiredArgs.Args[0].Name)
	assert.Equal(t, 1, len(dnsConfigBlock.OptionalArgs.Args))
	assert.Equal(t, "search_domains", dnsConfigBlock.OptionalArgs.Args[0].Name)
	assert.Equal(t, []string{"resource", "azurerm_container_group", "dns_config"}, dnsConfigBlock.Path)
}

func TestBuildNestedBlock_BothRequiredAndOptionalNestedBlocks(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location
  container {
    name   = "hello-world"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = "0.5"
    memory = "1.5"
  }
  dns_config {
    nameservers = []
  }
}
`
	file, diag := hclsyntax.ParseConfig([]byte(code), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	resourceBlock := BuildResourceBlock(file.Body.(*hclsyntax.Body).Blocks[0], file, func(block Block) error { return nil })
	assert.Equal(t, 1, len(resourceBlock.RequiredNestedBlocks.Blocks))
	assert.Equal(t, 1, len(resourceBlock.OptionalNestedBlocks.Blocks))
	assert.Equal(t, "container", resourceBlock.RequiredNestedBlocks.Blocks[0].Name)
	assert.Equal(t, "dns_config", resourceBlock.OptionalNestedBlocks.Blocks[0].Name)
}

func TestBuildNestedBlock_DynamicInnerNestedBlock(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location
  container {
    name   = "hello-world"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = "0.5"
    memory = "1.5"
	memory_limit = 1.5
	
	dynamic "gpu_limit" {
		for_each = var.gpu_limit ? [1] : []
	
		content {
			count = 1
			sku = "K80"
		}
	}
  }
}
`
	file, diag := hclsyntax.ParseConfig([]byte(code), "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	resourceBlock := BuildResourceBlock(file.Body.(*hclsyntax.Body).Blocks[0], file, func(block Block) error { return nil })
	assert.Equal(t, 1, len(resourceBlock.RequiredNestedBlocks.Blocks))
	assert.Nil(t, resourceBlock.OptionalNestedBlocks)
	containerBlock := resourceBlock.RequiredNestedBlocks.Blocks[0]
	assert.Equal(t, 1, len(containerBlock.OptionalNestedBlocks.Blocks))
	gpuLimitBlock := containerBlock.OptionalNestedBlocks.Blocks[0]
	assert.Equal(t, "gpu_limit", gpuLimitBlock.Name)
	assert.Equal(t, "gpu_limit", gpuLimitBlock.SortField)
	assert.Nil(t, gpuLimitBlock.RequiredArgs)
	assert.Equal(t, 2, len(gpuLimitBlock.OptionalArgs.Args))
	assert.Equal(t, "count", gpuLimitBlock.OptionalArgs.Args[0].Name)
	assert.Equal(t, "sku", gpuLimitBlock.OptionalArgs.Args[1].Name)
}
