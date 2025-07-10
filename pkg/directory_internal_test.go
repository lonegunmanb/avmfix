package pkg

import (
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_EnsureModulesShouldRunTerraformAlways(t *testing.T) {
	mockFs := afero.NewMemMapFs()
	require.NoError(t, mockFs.Mkdir("/tmp/.terraform", 0644))
	getInvoked := false
	stub := gostub.Stub(&Fs, mockFs).Stub(&terraformGetFunc, func(string) error {
		getInvoked = true
		return nil
	})
	defer stub.Reset()
	require.NoError(t, DirectoryAutoFix("/tmp"))
	assert.True(t, getInvoked)
}
