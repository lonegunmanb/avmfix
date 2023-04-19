package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type HclBlock struct {
	*hclsyntax.Block
	WriteBlock *hclwrite.Block
	tokens     hclwrite.Tokens
}

func NewHclBlock(rb *hclsyntax.Block, wb *hclwrite.Block) *HclBlock {
	r := &HclBlock{Block: rb, WriteBlock: wb}
	r.tokens = r.WriteBlock.BuildTokens(hclwrite.Tokens{})
	return r
}

func (b *HclBlock) Attributes() map[string]*HclAttribute {
	attributes := b.Block.Body.Attributes
	r := make(map[string]*HclAttribute, len(attributes))
	for name, attribute := range attributes {
		r[name] = NewHclAttribute(attribute, b.WriteBlock.Body().GetAttribute(name))
	}
	return r
}

func (b *HclBlock) NestedBlocks() []*HclBlock {
	blocks := b.Block.Body.Blocks
	r := make([]*HclBlock, len(blocks))
	for i, block := range blocks {
		r[i] = NewHclBlock(block, b.WriteBlock.Body().Blocks()[i])
	}
	return r
}

func (b *HclBlock) Clear() {
	b.WriteBlock.Body().Clear()
}
