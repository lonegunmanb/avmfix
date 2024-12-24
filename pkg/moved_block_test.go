package pkg_test

import (
	"testing"

	"github.com/lonegunmanb/avmfix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MovedBlock_AutoFix(t *testing.T) {
	expected := `
moved {
  from = "aws_instance.old_name"
  to   = "aws_instance.new_name"
}
`
	inputs := map[string]string{
		"need_sort": `
moved {
  to   = "aws_instance.new_name"
  from = "aws_instance.old_name"
}
`,
		"wel_formatted": expected,
	}
	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			file, diag := pkg.ParseConfig([]byte(input), "")
			require.False(t, diag.HasErrors())
			movedBlock := pkg.BuildMovedBlock(file.GetBlock(0), file.File)
			movedBlock.AutoFix()
			fixed := string(file.WriteFile.Bytes())
			assert.Equal(t, formatHcl(expected), formatHcl(fixed))
		})
	}
}
