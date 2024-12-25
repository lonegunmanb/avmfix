package pkg

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

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
