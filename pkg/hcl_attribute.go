package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type HclAttribute struct {
	*hclsyntax.Attribute
	WriteAttribute *hclwrite.Attribute
	tokens         hclwrite.Tokens
}

func NewHclAttribute(attribute *hclsyntax.Attribute, writeAttribute *hclwrite.Attribute) *HclAttribute {
	r := &HclAttribute{
		Attribute:      attribute,
		WriteAttribute: writeAttribute,
	}
	r.tokens = r.WriteAttribute.BuildTokens(hclwrite.Tokens{})
	return r
}
