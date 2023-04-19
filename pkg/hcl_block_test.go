package pkg_test

import (
	"testing"

	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/lonegunmanb/azure-verified-module-fix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HclWriteBlockClear(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location

  dns_config {
	search_domains = []
    nameservers = []
  }
}
`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	block := file.GetBlock(0)
	block.Clear()
	expected := `resource "azurerm_container_group" "example" {}`
	assert.Equal(t, s2t(expected), s2t(string(block.WriteBlock.BuildTokens(hclwrite.Tokens{}).Bytes())))
}

func s2t(s string) hclwrite.Tokens {
	config, _ := hclwrite.ParseConfig([]byte(s), "", hcl.InitialPos)
	tokens := config.BuildTokens(hclwrite.Tokens{})
	linq.From(tokens).Where(func(t interface{}) bool {
		token := t.(*hclwrite.Token)
		return token.Type != hclsyntax.TokenNewline
	}).ToSlice(&tokens)
	return tokens
}
