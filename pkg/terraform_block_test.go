package pkg_test

import (
	"testing"

	"github.com/lonegunmanb/avmfix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RequiredProvidersSort(t *testing.T) {
	cases := []struct {
		code     string
		expected string
		desc     string
	}{
		{
			desc: "sort by name",
			code: `
terraform {
  required_version = ">= 1.3"

  required_providers {
    foo = {
      source  = "hashicorp/foo"
      version = "1.0.0"
    }
    bar = {
      source  = "hashicorp/bar"
      version = ">= 0.3.2, < 1.0"
    }
  }
}
`,
			expected: `
terraform {
  required_version = ">= 1.3"

  required_providers {
    bar = {
      source  = "hashicorp/bar"
      version = ">= 0.3.2, < 1.0"
    }
    foo = {
      source  = "hashicorp/foo"
      version = "1.0.0"
    }
  }
}
`,
		},
		{
			desc: "single provider",
			code: `
terraform {
  required_version = ">= 1.3"

  required_providers {
    bar = {
      source  = "hashicorp/bar"
      version = ">= 0.3.2, < 1.0"
    }
  }
}
`,
			expected: `
terraform {
  required_version = ">= 1.3"

  required_providers {
    bar = {
      source  = "hashicorp/bar"
      version = ">= 0.3.2, < 1.0"
    }
  }
}
`,
		},
		{
			desc: "empty provider",
			code: `
terraform {
  required_version = ">= 1.3"

  required_providers {
  }
}
`,
			expected: `
terraform {
  required_version = ">= 1.3"

  required_providers {
  }
}
`,
		},
		{
			desc: "no provider",
			code: `
terraform {
  required_version = ">= 1.3"
}
`,
			expected: `
terraform {
  required_version = ">= 1.3"
}
`,
		},
	}
	for _, cc := range cases {
		t.Run(cc.desc, func(t *testing.T) {
			f, diag := pkg.ParseConfig([]byte(cc.code), "")
			require.False(t, diag.HasErrors())
			sut := pkg.BuildTerraformBlock(f.GetBlock(0), f.File)
			sut.AutoFix()
			fixed := string(f.WriteFile.Bytes())
			assert.Equal(t, formatHcl(cc.expected), formatHcl(fixed))

		})
	}
}
