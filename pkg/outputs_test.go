package pkg_test

import (
	"github.com/lonegunmanb/azure-verified-module-fix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOutputsFile_SortOutputAttribute(t *testing.T) {
	output := `
output "test" {
  value = "test"
  sensitive = false
  description = "test"
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	outputBlock := pkg.BuildOutputsFile(f)
	outputBlock.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `
output "test" {
  description = "test"
  sensitive = false
  value = "test"
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}
