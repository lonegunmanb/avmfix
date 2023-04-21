package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type HclAttribute struct {
	*hclsyntax.Attribute
	WriteAttribute *hclwrite.Attribute
}

func NewHclAttribute(attribute *hclsyntax.Attribute, writeAttribute *hclwrite.Attribute) *HclAttribute {
	r := &HclAttribute{
		Attribute:      attribute,
		WriteAttribute: writeAttribute,
	}
	return r
}

func (a *HclAttribute) IsNullable() bool {
	expr, ok := a.Expr.(*hclsyntax.LiteralValueExpr)
	return ok && expr.Val.True()
}
