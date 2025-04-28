package pkg_test

import (
	"testing"

	"github.com/lonegunmanb/avmfix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RemovedBlock_AutoFix(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "need sort",
			input: `removed {
  lifecycle {
    destroy = false
  }

  from = aws_instance.example
}`,
			expected: `removed {
  from = aws_instance.example

  lifecycle {
    destroy = false
  }
}`,
		},
		{
			name: "with provisioner",
			input: `removed {
  provisioner "local-exec" {
    when    = destroy
    command = "echo 'Instance ${self.id} has been destroyed.'"
  }
  lifecycle {
    destroy = false
  }

  from = aws_instance.example
}`,
			expected: `removed {
  from = aws_instance.example

  lifecycle {
    destroy = false
  }

  provisioner "local-exec" {
    when    = destroy
    command = "echo 'Instance ${self.id} has been destroyed.'"
  }
}`,
		},
		{
			name: "multiple provisioners",
			input: `removed {
  provisioner "local-exec" {
    when    = destroy
    command = "echo 'destroyed'"
  }
  lifecycle {
    destroy = false
  }

  from = aws_instance.example
  provisioner "local-exec" {
    when    = destroy
    command = "echo 'Instance ${self.id} has been destroyed.'"
  }
}`,
			expected: `removed {
  from = aws_instance.example

  lifecycle {
    destroy = false
  }

  provisioner "local-exec" {
    when    = destroy
    command = "echo 'destroyed'"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "echo 'Instance ${self.id} has been destroyed.'"
  }
}`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			file, diag := pkg.ParseConfig([]byte(c.input), "")
			require.False(t, diag.HasErrors())
			movedBlock := pkg.BuildRemovedBlock(file.GetBlock(0), file.File)
			err := movedBlock.AutoFix()
			require.NoError(t, err)
			fixed := string(file.WriteFile.Bytes())
			assert.Equal(t, formatHcl(c.expected), formatHcl(fixed))
		})
	}
}
