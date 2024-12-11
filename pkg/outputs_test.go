package pkg_test

import (
	"testing"

	"github.com/lonegunmanb/avmfix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputsFile_SortOutputAttribute(t *testing.T) {
	output := `output "test" {
  value = "test"
  sensitive = true
  description = "test"
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	outputBlock := pkg.BuildOutputsFile(f)
	outputBlock.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `output "test" {
  description = "test"
  sensitive = true
  value = "test"
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestOutputsFile_RemoveUnnecessarySensitive(t *testing.T) {
	output := `output "test" {
  sensitive = false
  description = "test"
  value = "test"
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	outputBlock := pkg.BuildOutputsFile(f)
	outputBlock.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `output "test" {
  description = "test"
  value = "test"
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestOutputsFile_SortOutputsByName(t *testing.T) {
	output := `# output test2
output "test2" {
  # test2 value
  value = "test2"
  # test2 description
  description = "test2"
}
output "test" {
  value = "test"
  description = "test"
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	outputsFile := pkg.BuildOutputsFile(f)
	outputsFile.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `output "test" {
  description = "test"
  value = "test"
}

# output test2
output "test2" {
  # test2 description
  description = "test2"
  # test2 value
  value = "test2"
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestOutputsFile_MustPreservePotentialSeperatedFirstLineComment(t *testing.T) {
	output := `# tfint-ignore-file: terraform-standard_module_structure

output "image_id" {
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	outputBlock := pkg.BuildOutputsFile(f)
	outputBlock.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	assert.Equal(t, formatHcl(output), formatHcl(fixed))
}
