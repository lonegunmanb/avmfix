package pkg

import (
	"path/filepath"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_EnsureModulesShouldRunTerraformAlways(t *testing.T) {
	mockFs := afero.NewMemMapFs()
	require.NoError(t, mockFs.Mkdir("/tmp/.terraform", 0644))
	getInvoked := false
	stub := gostub.Stub(&Fs, mockFs).Stub(&terraformInitFunc, func(string) error {
		getInvoked = true
		return nil
	})
	defer stub.Reset()
	require.NoError(t, DirectoryAutoFix("/tmp"))
	assert.True(t, getInvoked)
}

func TestParseTerraformLockFile(t *testing.T) {
	tests := []struct {
		name        string
		lockContent string
		expected    map[string]map[string]string
		shouldError bool
	}{
		{
			name: "valid lock file with single provider",
			lockContent: `provider "registry.terraform.io/hashicorp/azurerm" {
  version     = "4.37.0"
  constraints = ">= 3.74.0, ~> 4.0"
  hashes = [
    "h1:MfFA2dyXwJlMi4p7PBjQzyRDLm0vcpnVeMPedvUT6BE=",
    "zh:10acb818823a0319215beb796af1a7a97820be5d40ec1779deb8c2bdb1cac6d0",
  ]
}`,
			expected: map[string]map[string]string{
				"hashicorp": {
					"azurerm": "4.37.0",
				},
			},
		},
		{
			name: "valid lock file with multiple providers",
			lockContent: `provider "registry.terraform.io/hashicorp/azurerm" {
  version     = "4.37.0"
  constraints = ">= 3.74.0, ~> 4.0"
  hashes = [
    "h1:MfFA2dyXwJlMi4p7PBjQzyRDLm0vcpnVeMPedvUT6BE=",
  ]
}

provider "registry.terraform.io/hashicorp/random" {
  version = "3.6.0"
  hashes = [
    "h1:I8MBeauYA8J8yheLJ8oSMWqB0kovn16dF/wKZ1QTdkk=",
  ]
}`,
			expected: map[string]map[string]string{
				"hashicorp": {
					"azurerm": "4.37.0",
					"random":  "3.6.0",
				},
			},
		},
		{
			name: "valid lock file with multiple namespaces",
			lockContent: `provider "registry.terraform.io/hashicorp/azurerm" {
  version     = "4.37.0"
  hashes = [
    "h1:MfFA2dyXwJlMi4p7PBjQzyRDLm0vcpnVeMPedvUT6BE=",
  ]
}

provider "registry.terraform.io/microsoft/azuredevops" {
  version = "1.3.0"
  hashes = [
    "h1:I8MBeauYA8J8yheLJ8oSMWqB0kovn16dF/wKZ1QTdkk=",
  ]
}`,
			expected: map[string]map[string]string{
				"hashicorp": {
					"azurerm": "4.37.0",
				},
				"microsoft": {
					"azuredevops": "1.3.0",
				},
			},
		},
		{
			name:        "no lock file exists",
			lockContent: "",
			shouldError: true,
		},
		{
			name: "lock file with no version attribute",
			lockContent: `provider "registry.terraform.io/hashicorp/azurerm" {
  constraints = ">= 3.74.0, ~> 4.0"
  hashes = [
    "h1:MfFA2dyXwJlMi4p7PBjQzyRDLm0vcpnVeMPedvUT6BE=",
  ]
}`,
			expected: map[string]map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := gostub.Stub(&parseTerraformLockFile, parseTerraformLockFileStub)
			defer stub.Reset()
			// Setup filesystem
			mockFs := afero.NewMemMapFs()
			testDir := "/test"
			err := mockFs.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			// Write lock file if content is provided
			if tt.lockContent != "" {
				lockPath := filepath.Join(testDir, ".terraform.lock.hcl")
				err = afero.WriteFile(mockFs, lockPath, []byte(tt.lockContent), 0644)
				require.NoError(t, err)
			}

			// Stub the filesystem
			defer gostub.Stub(&Fs, mockFs).Reset()

			sut := &directory{path: testDir}
			// Call the function
			err = sut.parseTerraformLockFile()
			result := sut.providerVersions

			// Check results
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDirectoryContainsModuleBlockShouldRunTerraformInitFirst(t *testing.T) {
	called := false
	gostub.Stub(&parseTerraformLockFile, func(lockFilePath string) (map[string]map[string]string, error) {
		called = true
		return nil, nil
	})
	_ = DirectoryAutoFix(filepath.Join("test-fixture", "local_module"))
	assert.True(t, called)
}
