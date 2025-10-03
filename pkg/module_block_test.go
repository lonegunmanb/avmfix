package pkg

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadLocalModuleAsBlockSchema(t *testing.T) {
	moduleHclConfig := `module "local" {
	source = "./test_module"
}
`
	syntaxFile, diag := hclsyntax.ParseConfig([]byte(moduleHclConfig), "test.tf", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	writeFile, diag := hclwrite.ParseConfig([]byte(moduleHclConfig), "test.hcl", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	block := syntaxFile.Body.(*hclsyntax.Body).Blocks[0]
	writeBlock := writeFile.Body().Blocks()[0]
	sut := &ModuleBlock{
		dir: filepath.Join("test-fixture", "local_module"),
		HclBlock: &HclBlock{
			Block:      block,
			WriteBlock: writeBlock,
		},
	}
	schemaBlock, err := sut.schemaBlock()
	require.NoError(t, err)
	assert.Len(t, schemaBlock.Attributes, 2)
	assert.True(t, schemaBlock.Attributes["required_variable"].Required)
	assert.True(t, schemaBlock.Attributes["optional_variable"].Optional)
}

func TestLoadRegistryModuleAsBlockSchema(t *testing.T) {
	moduleHclConfig := `module "consul" {
  source = "hashicorp/consul/aws"
  version = "0.1.0"
}
`
	syntaxFile, diag := hclsyntax.ParseConfig([]byte(moduleHclConfig), "test.tf", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	writeFile, diag := hclwrite.ParseConfig([]byte(moduleHclConfig), "test.tf", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	block := syntaxFile.Body.(*hclsyntax.Body).Blocks[0]
	writeBlock := writeFile.Body().Blocks()[0]
	sut := &ModuleBlock{
		dir: filepath.Join("test-fixture", "remote_module"),
		HclBlock: &HclBlock{
			Block:      block,
			WriteBlock: writeBlock,
		},
	}

	schemaBlock, err := sut.schemaBlock()
	require.NoError(t, err)
	assert.Len(t, schemaBlock.Attributes, 2)
	assert.True(t, schemaBlock.Attributes["required_variable"].Required)
	assert.True(t, schemaBlock.Attributes["optional_variable"].Optional)
}

func TestModuleAutoFix(t *testing.T) {
	moduleHclConfig := `module "consul" {
  optional_variable = "value"
  required_variable = "value"
  source = "hashicorp/consul/aws"
  version = "0.1.0"
  for_each = var.for_each
  providers = {}
  depends_on = [null_resource.this]
}
`
	file, diag := ParseConfig([]byte(moduleHclConfig), "test.tf")
	require.False(t, diag.HasErrors())
	sut, err := BuildModuleBlock(file.GetBlock(0), filepath.Join("test-fixture", "remote_module"), file)
	require.NoError(t, err)
	err = sut.AutoFix()
	require.NoError(t, err)
	fixed := string(sut.HclBlock.WriteBlock.BuildTokens(nil).Bytes())
	assert.Equal(t, `module "consul" {
  source = "hashicorp/consul/aws"
  version = "0.1.0"
  providers = {}
  for_each = var.for_each

  required_variable = "value"
  optional_variable = "value"

  depends_on = [null_resource.this]
}
`, fixed)
}
