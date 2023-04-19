package pkg_test

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"testing"

	"github.com/lonegunmanb/azure-verified-module-fix/pkg"
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
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	assert.Equal(t, 1, len(resourceBlock.OptionalNestedBlocks.Blocks))
	assert.Nil(t, resourceBlock.RequiredNestedBlocks)
	dnsConfigBlock := resourceBlock.OptionalNestedBlocks.Blocks[0]
	assert.Equal(t, "dns_config", dnsConfigBlock.Name)
	assert.Equal(t, "dns_config", dnsConfigBlock.SortField)
	assert.Equal(t, file.File, dnsConfigBlock.File)
	assert.Equal(t, resourceBlock.HclBlock.Body.Blocks[0], dnsConfigBlock.HclBlock.Block)
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
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	assert.Equal(t, 1, len(resourceBlock.RequiredNestedBlocks.Blocks))
	assert.Nil(t, resourceBlock.OptionalNestedBlocks)
	containerBlock := resourceBlock.RequiredNestedBlocks.Blocks[0]
	assert.Equal(t, "container", containerBlock.Name)
	assert.Equal(t, "container", containerBlock.SortField)
	assert.Equal(t, containerBlock.HclBlock.Range(), containerBlock.Range)
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
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
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
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
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
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
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

func TestNestedBlock_NotWellGaped(t *testing.T) {
	inputs := []string{
		`
resource "azurerm_container_group" "example" {
  location = azurerm_resource_group.example.location

  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"
	memory_limit = 1.5
	gpu_limit {
		count = 1
		sku = "K80"
	}
  }
}
`,
	}

	for i := 0; i < len(inputs); i++ {
		file, diag := pkg.ParseConfig([]byte(inputs[i]), "")
		require.False(t, diag.HasErrors())
		resourceBlock := pkg.BuildResourceBlock(file.GetBlock(i), file.File, func(block pkg.Block) error { return nil })
		assert.False(t, resourceBlock.RequiredNestedBlocks.Blocks[0].CheckOrder())
	}
}

func TestNestedBlock_WellFormatNestedBlock(t *testing.T) {
	inputs := []string{`
resource "azurerm_container_group" "example" {
  location = azurerm_resource_group.example.location

  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"
	memory_limit = 1.5

	dynamic "gpu" {
		for_each = var.gpu == null ? [] : [1]

		content {
			count = var.gpu.count
			sku   = var.gpu.sku
		}
	}
	gpu_limit {
		count = 1
		sku = "K80"
	}
  }
}
`,
		`
resource "azurerm_container_group" "example" {
  location = azurerm_resource_group.example.location

  container {
    cpu    = "0.5"
	# comment
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"
	memory_limit = 1.5

	dynamic "gpu" {
		for_each = var.gpu == null ? [] : [1]

		content {
			count = var.gpu.count
			# comment
			sku   = var.gpu.sku
		}
	}
	gpu_limit {
		count = 1
		sku = "K80"
	}
  }
}
`}
	for i := 0; i < len(inputs); i++ {
		code := inputs[i]
		file, diag := pkg.ParseConfig([]byte(code), "")
		require.False(t, diag.HasErrors())
		resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
		assert.Equal(t, 1, len(resourceBlock.RequiredNestedBlocks.Blocks))
		containerBlock := resourceBlock.RequiredNestedBlocks.Blocks[0]
		assert.True(t, containerBlock.CheckOrder())
		gpuBlock := containerBlock.OptionalNestedBlocks.Blocks[0]
		assert.True(t, gpuBlock.CheckOrder())
		gpuLimitBlock := containerBlock.OptionalNestedBlocks.Blocks[1]
		assert.True(t, gpuLimitBlock.CheckOrder())
	}
}

func TestBuildNestedBlock_AutoFix_ReorderRequiredArgsByNames(t *testing.T) {
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
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File, func(block pkg.Block) error { return nil })
	cb := resourceBlock.RequiredNestedBlocks.Blocks[0]
	cb.AutoFix()
	expected := `container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"
	memory_limit = 1.5
	
	gpu_limit {
		count = 1
		sku = "K80"
	}
  }
`
	s := string(cb.HclBlock.WriteBlock.BuildTokens(hclwrite.Tokens{}).Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(s))
}

func formatHcl(input string) string {
	// Create a new HCL file from the input string
	f, _ := hclwrite.ParseConfig([]byte(input), "", hcl.InitialPos)

	// Format the HCL file
	formatted := f.Bytes()

	return string(formatted)
}
