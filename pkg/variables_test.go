package pkg_test

import (
	"github.com/lonegunmanb/azure-verified-module-fix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVariablesFile_VariableBlockAttributeSort(t *testing.T) {
	output := `variable "image_id" {
  validation {
    condition     = length(var.image_id) > 4 && substr(var.image_id, 0, 4) == "ami-"
    error_message = "The image_id value must be a valid AMI id, starting with \"ami-\"."
  }
  nullable    = false
  sensitive   = true
  default     = "ami-123456"
  description = "The id of the machine image (AMI) to use for the server."
  type        = string
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	variablesFile := pkg.BuildVariablesFile(f)
	variablesFile.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `variable "image_id" {
  type        = string
  default     = "ami-123456"
  description = "The id of the machine image (AMI) to use for the server."
  nullable    = false
  sensitive   = true
  
  validation {
    condition     = length(var.image_id) > 4 && substr(var.image_id, 0, 4) == "ami-"
    error_message = "The image_id value must be a valid AMI id, starting with \"ami-\"."
  }
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestVariablesFile_RequiredVariableShouldHavePriority(t *testing.T) {
	output := `variable "test" {
  description = "test"
  type        = string
}

variable "test2" {
  description = "test2"
  type        = string
  nullable    = false
}

variable "test3" {
  description = "test3"
  type        = string
}

variable "test4" {
  description = "test4"
  type        = string
  nullable    = false
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	variablesFile := pkg.BuildVariablesFile(f)
	variablesFile.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `variable "test2" {
  type        = string
  description = "test2"
  nullable    = false
}

variable "test4" {
  type        = string
  description = "test4"
  nullable    = false
}

variable "test" {
  type        = string
  description = "test"
}

variable "test3" {
  type        = string
  description = "test3"
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestVariablesFile_RemoveUnnecessaryNullable(t *testing.T) {
	output := `variable "image_id" {
  type        = string
  description = "The id of the machine image (AMI) to use for the server."
  nullable    = true
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	variablesFile := pkg.BuildVariablesFile(f)
	variablesFile.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `variable "image_id" {
  type        = string
  description = "The id of the machine image (AMI) to use for the server."
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestVariablesFile_RemoveUnnecessarySensitive(t *testing.T) {
	output := `variable "image_id" {
  type        = string
  description = "The id of the machine image (AMI) to use for the server."
  sensitive   = false
}
`
	f, diag := pkg.ParseConfig([]byte(output), "outputs.tf")
	require.False(t, diag.HasErrors())
	variablesFile := pkg.BuildVariablesFile(f)
	variablesFile.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `variable "image_id" {
  type        = string
  description = "The id of the machine image (AMI) to use for the server."
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}
