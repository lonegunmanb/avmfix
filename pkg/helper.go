package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"sort"
)

var headMetaArgPriority = map[string]int{"for_each": 0, "count": 0, "provider": 1}
var tailMetaArgPriority = map[string]int{"lifecycle": 0, "depends_on": 1}

// IsHeadMeta checks whether a name represents a type of head Meta arg
func (b *ResourceBlock) isHeadMeta(argName string) bool {
	_, isHeadMeta := headMetaArgPriority[argName]
	return isHeadMeta
}

// IsTailMeta checks whether a name represents a type of tail Meta arg
func (b *ResourceBlock) isTailMeta(argName string) bool {
	_, isTailMeta := tailMetaArgPriority[argName]
	return isTailMeta
}

func attributesByLines(attributes map[string]*HclAttribute) []*HclAttribute {
	var attrs []*HclAttribute
	for _, attr := range attributes {
		attrs = append(attrs, attr)
	}
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].Range().Start.Line < attrs[j].Range().Start.Line
	})
	return attrs
}

func removeIndex[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}

func endWithNewLine(b *hclwrite.Block) bool {
	tokens := b.BuildTokens(hclwrite.Tokens{})
	return tokens[len(tokens)-1].Type == hclsyntax.TokenNewline
}
