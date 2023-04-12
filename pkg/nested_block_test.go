package pkg

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BuildNestedBlock(t *testing.T) {
	inputs := []struct {
		name                        string
		code                        string
		expectedRequiredNestedBlock []*NestedBlock
		expectedOptionalNestedBlock []*NestedBlock
	}{
		{
			name: "1. one optional nested block",
			code: `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location
  dns_config {
    nameservers = []
	search_domains = []
  }
}
`,
			expectedOptionalNestedBlock: []*NestedBlock{
				{
					Name:      "dns_config",
					SortField: "dns_config",
					RequiredArgs: &Args{
						Args: []*Arg{
							{
								Name: "nameservers",
							},
						},
					},
					OptionalArgs: &Args{
						Args: []*Arg{
							{
								Name: "search_domains",
							},
						},
					},
					Path: []string{"resource", "azurerm_container_group", "dns_config"},
				},
			},
		},
		{
			name: "2. one optional nested block, one required nested block",
			code: `
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
`,
			expectedRequiredNestedBlock: []*NestedBlock{
				{
					Name:      "container",
					SortField: "container",
					RequiredArgs: &Args{
						Args: []*Arg{
							{
								Name: "name",
							},
							{
								Name: "image",
							},
							{
								Name: "cpu",
							},
							{
								Name: "memory",
							},
						},
					},
					Path: []string{"resource", "azurerm_container_group", "container"},
				},
			},
			expectedOptionalNestedBlock: []*NestedBlock{
				{
					Name:      "dns_config",
					SortField: "dns_config",
					RequiredArgs: &Args{
						Args: []*Arg{
							{
								Name: "nameservers",
							},
						},
					},
					Path: []string{"resource", "azurerm_container_group", "dns_config"},
				},
			},
		},
	}

	for i := 0; i < len(inputs); i++ {
		input := inputs[i]
		t.Run(input.name, func(t *testing.T) {
			file, diag := hclsyntax.ParseConfig([]byte(input.code), "", hcl.InitialPos)
			require.False(t, diag.HasErrors())
			resourceBlock := BuildResourceBlock(file.Body.(*hclsyntax.Body).Blocks[0], file, func(block Block) error { return nil })
			processNestedBlockForCompare(resourceBlock.RequiredNestedBlocks)
			processNestedBlockForCompare(resourceBlock.OptionalNestedBlocks)

			var actual []*NestedBlock = nil
			if resourceBlock.RequiredNestedBlocks != nil {
				actual = resourceBlock.RequiredNestedBlocks.Blocks
			}
			assert.Equal(t, input.expectedRequiredNestedBlock, actual)
			if resourceBlock.OptionalNestedBlocks != nil {
				actual = resourceBlock.OptionalNestedBlocks.Blocks
			}
			assert.Equal(t, input.expectedOptionalNestedBlock, actual)
		})
	}
}

func processNestedBlockForCompare(nbs *NestedBlocks) {
	if nbs == nil {
		return
	}
	for _, nb := range nbs.Blocks {
		nb.Block = nil
		nb.Range = hcl.Range{}
		nb.File = nil
		nb.writeBlock = nil
		nb.emit = nil
		if nb.HeadMetaArgs != nil {
			nb.HeadMetaArgs.Range = nil
			for _, arg := range nb.HeadMetaArgs.Args {
				processArgForCompare(arg)
			}
		}
		if nb.RequiredArgs != nil {
			nb.RequiredArgs.Range = nil
			for _, arg := range nb.RequiredArgs.Args {
				processArgForCompare(arg)
			}
		}
		if nb.OptionalArgs != nil {
			nb.OptionalArgs.Range = nil
			for _, arg := range nb.OptionalArgs.Args {
				processArgForCompare(arg)
			}
		}
		processNestedBlockForCompare(nb.RequiredNestedBlocks)
		processNestedBlockForCompare(nb.OptionalNestedBlocks)
	}
}

func processArgForCompare(arg *Arg) {
	arg.Range = hcl.Range{}
	arg.File = nil
	arg.sAttr = nil
	arg.wAttr = nil
}
