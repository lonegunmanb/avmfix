package pkg_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	config, diagnostics := hclsyntax.ParseConfig([]byte(code), "", hcl.InitialPos)
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(config.Body.(*hclsyntax.Body).Blocks[0], config, func(block pkg.Block) error { return nil })
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
	config, diagnostics := hclsyntax.ParseConfig([]byte(code), "", hcl.InitialPos)
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(config.Body.(*hclsyntax.Body).Blocks[0], config, func(block pkg.Block) error { return nil })
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
