package pkg_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexPatterns(t *testing.T) {
	var outputsFileRegex = regexp.MustCompile(`.*?outputs.*?\.tf$`)
	var variablesFileRegex = regexp.MustCompile(`.*?variables.*?\.tf$`)

	tests := []struct {
		name     string
		regex    *regexp.Regexp
		filename string
		matches  bool
	}{
		// Tests for outputsFileRegex (line 38)
		{"outputs.tf exact match", outputsFileRegex, "outputs.tf", true},
		{"my_outputs.tf match", outputsFileRegex, "my_outputs.tf", true},
		{"extra_outputs.tf match", outputsFileRegex, "extra_outputs.tf", true},
		{"outputs.tfvars no match", outputsFileRegex, "outputs.tfvars", false},
		{"outputs.tf.backup no match", outputsFileRegex, "outputs.tf.backup", false},
		{"random.tf no match", outputsFileRegex, "random.tf", false},

		// Tests for variablesFileRegex (line 39)
		{"variables.tf exact match", variablesFileRegex, "variables.tf", true},
		{"my_variables.tf match", variablesFileRegex, "my_variables.tf", true},
		{"variables_extra.tf match", variablesFileRegex, "variables_extra.tf", true},
		{"variables.tfvars no match", variablesFileRegex, "variables.tfvars", false},
		{"variables.tf.backup no match", variablesFileRegex, "variables.tf.backup", false},
		{"random.tf no match", variablesFileRegex, "random.tf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.matches, tt.regex.MatchString(tt.filename))
		})
	}
}
