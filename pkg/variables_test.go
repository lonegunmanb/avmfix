package pkg_test

import (
	"strconv"
	"testing"

	"github.com/lonegunmanb/avmfix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariablesFile_VariableBlockAttributeSort(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input: `variable "image_id" {
  validation {
    condition     = length(var.image_id) > 4 && substr(var.image_id, 0, 4) == "ami-"
    error_message = "The image_id value must be a valid AMI id, starting with \"ami-\"."
  }
  sensitive   = true
  nullable    = false
  description = "The id of the machine image (AMI) to use for the server."
  default     = "ami-123456"
  type        = string
}
`,
			expected: `variable "image_id" {
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
`,
		},
	}
	for i, c := range cases {
		input := c.input
		expected := c.expected
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f, diag := pkg.ParseConfig([]byte(input), "variables.tf")
			require.False(t, diag.HasErrors())
			variablesFile := pkg.BuildVariablesFile(f)
			variablesFile.AutoFix()
			fixed := string(f.WriteFile.Bytes())
			assert.Equal(t, formatHcl(expected), formatHcl(fixed))
		})
	}
}

func TestVariablesFile_RequiredVariableShouldHavePriority(t *testing.T) {
	code := `variable "test" {
  description = "test"
  default     = null
  type        = string
  nullable    = true
}

variable "test2" {
  description = "test2"
  default     = null
  type        = string
  nullable    = true
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
	f, diag := pkg.ParseConfig([]byte(code), "variables.tf")
	require.False(t, diag.HasErrors())
	variablesFile := pkg.BuildVariablesFile(f)
	variablesFile.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `variable "test3" {
  type        = string
  description = "test3"
}

variable "test4" {
  type        = string
  description = "test4"
  nullable    = false
}

variable "test" {
  type        = string
  default     = null
  description = "test"
}

variable "test2" {
  type        = string
  default     = null
  description = "test2"
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestVariablesFile_RemoveUnnecessaryNullable(t *testing.T) {
	code := `variable "image_id" {
  nullable    = true
  type        = string
  description = "The id of the machine image (AMI) to use for the server."
}
`
	f, diag := pkg.ParseConfig([]byte(code), "variables.tf")
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
	code := `variable "image_id" {
  type        = string
  description = "The id of the machine image (AMI) to use for the server."
  sensitive   = false
}
`
	f, diag := pkg.ParseConfig([]byte(code), "variables.tf")
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

func TestVariablesFile_NoGapBetweenTwoBlocks(t *testing.T) {
	code := `variable "rg_so" {
  type        = string
  description = "Name of the Resource group in which to deploy service objects"
}

variable "rg_network" {
  type        = string
  description = "Name of the Resource group in which to deploy network objects"
}

variable "avdLocation" {
  description = "Location of the resource group."
}
variable "prefix" {
  type        = string
  description = "Prefix of the name of the AVD machine(s)"
}
variable "vnet" {
  type        = string
  description = "Name of avd vnet"
}

variable "pesnet" {
  type        = string
  description = "Name of subnet"
}

variable "domain_password" {
  type        = string
  description = "Password of the user to authenticate with the domain"
  sensitive   = true
}
variable "domain_user" {
  type        = string
  description = "Domain user to authenticate with the domain"
}`
	f, diag := pkg.ParseConfig([]byte(code), "variables.tf")
	require.False(t, diag.HasErrors())
	variablesFile := pkg.BuildVariablesFile(f)
	variablesFile.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	expected := `variable "avdLocation" {
  description = "Location of the resource group."
}

variable "domain_password" {
  type        = string
  description = "Password of the user to authenticate with the domain"
  sensitive   = true
}

variable "domain_user" {
  type        = string
  description = "Domain user to authenticate with the domain"
} 

variable "pesnet" {
  type        = string
  description = "Name of subnet"
}

variable "prefix" {
  type        = string
  description = "Prefix of the name of the AVD machine(s)"
}

variable "rg_network" {
  type        = string
  description = "Name of the Resource group in which to deploy network objects"
}

variable "rg_so" {
  type        = string
  description = "Name of the Resource group in which to deploy service objects"
}

variable "vnet" {
  type        = string
  description = "Name of avd vnet"
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestVariablesFile_MultipleValidations(t *testing.T) {
	code := `variable "image_id" {
  type        = string
  default     = "ami-123456"
  description = "The id of the machine image (AMI) to use for the server."
  nullable    = false
  sensitive   = true

  validation {
    condition     = length(var.image_id) > 4 && substr(var.image_id, 0, 4) == "ami-"
    error_message = "The image_id value must be a valid AMI id, starting with \"ami-\"."
  }
  validation {
    condition     = length(var.image_id) > 4 && substr(var.image_id, 0, 4) == "ami-"
    error_message = "The image_id value must be a valid AMI id, starting with \"ami-\"."
  }
}
`
	f, diag := pkg.ParseConfig([]byte(code), "variables.tf")
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
  validation {
    condition     = length(var.image_id) > 4 && substr(var.image_id, 0, 4) == "ami-"
    error_message = "The image_id value must be a valid AMI id, starting with \"ami-\"."
  }
}
`
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestVariablesFile_MustPreservePotentialSeperatedFirstLineComment(t *testing.T) {
	input := `# tfint-ignore-file: terraform-standard_module_structure

variable "image_id" {
}
`
	f, diag := pkg.ParseConfig([]byte(input), "variables.tf")
	require.False(t, diag.HasErrors())
	variablesFile := pkg.BuildVariablesFile(f)
	variablesFile.AutoFix()
	fixed := string(f.WriteFile.Bytes())
	assert.Equal(t, formatHcl(input), formatHcl(fixed))
}
