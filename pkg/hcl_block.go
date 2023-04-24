package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type HclBlock struct {
	*hclsyntax.Block
	WriteBlock *hclwrite.Block
}

func NewHclBlock(rb *hclsyntax.Block, wb *hclwrite.Block) *HclBlock {
	r := &HclBlock{Block: rb, WriteBlock: wb}
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

func (b *HclBlock) writeArgs(args Args, attributes map[string]*hclwrite.Attribute) {
	if args == nil {
		return
	}

	for _, arg := range args {
		tokens := attributes[arg.Name].BuildTokens(hclwrite.Tokens{})
		b.WriteBlock.Body().AppendUnstructuredTokens(tokens)
	}
}

func (b *HclBlock) appendBlock(nb *hclwrite.Block) {
	b.WriteBlock.Body().AppendBlock(nb)
}

func (b *HclBlock) appendNestedBlocks(nbs *NestedBlocks, originalBlocks []*hclwrite.Block) {
	if nbs == nil {
		return
	}
	var orderedBlocks []*NestedBlock
	linq.From(nbs.Blocks).OrderBy(func(i interface{}) interface{} {
		return i.(*NestedBlock).SortField
	}).ToSlice(&orderedBlocks)

	for _, ob := range orderedBlocks {
		tokens := originalBlocks[ob.Index].BuildTokens(hclwrite.Tokens{})
		b.WriteBlock.Body().AppendUnstructuredTokens(tokens)
	}
}

func (b *HclBlock) appendNewline() {
	b.WriteBlock.Body().AppendNewline()
}

func (b *HclBlock) isSingleLineBlock() bool {
	tokens := b.WriteBlock.BuildTokens(hclwrite.Tokens{})
	cBrace := false
	for i := len(tokens) - 1; i >= 1; i-- {
		if tokens[i].Type == hclsyntax.TokenCBrace {
			cBrace = true
		}
		if cBrace && tokens[i-1].Type == hclsyntax.TokenNewline {
			return false
		}
	}
	return true
}
