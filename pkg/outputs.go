package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type OutputBlock struct {
	Block      *HclBlock
	Attributes Args
	Index      int
}

func BuildOutputBlock(f *hcl.File, b *HclBlock, index int) *OutputBlock {
	r := &OutputBlock{
		Block: b,
		Index: index,
	}
	for _, attribute := range attributesByLines(b.Attributes()) {
		r.Attributes = append(r.Attributes, buildAttrArg(attribute, f))
	}
	return r
}

func (b *OutputBlock) AutoFix() {
	b.removeUnnecessarySensitive()
	b.sortArguments()
	b.write()
}

func (b *OutputBlock) write() {
	attributes := b.Block.WriteBlock.Body().Attributes()
	b.Block.Clear()
	b.Block.writeNewLine()
	b.Block.writeArgs(b.Attributes, attributes)
}

func (b *OutputBlock) removeUnnecessarySensitive() {
	for i := 0; i < len(b.Attributes); i++ {
		attr := b.Attributes[i]
		if attr.Name != "sensitive" {
			continue
		}
		literal, ok := attr.Attribute.Expr.(*hclsyntax.LiteralValueExpr)
		if !ok || !literal.Val.False() {
			continue
		}
		b.Attributes = removeIndex(b.Attributes, i)
		return
	}
}

func (b *OutputBlock) sortArguments() {
	b.Attributes = b.Attributes.SortByName()
}

type OutputsFile struct {
	File *HclFile
}

func BuildOutputsFile(f *HclFile) *OutputsFile {
	return &OutputsFile{
		File: f,
	}
}

func (f *OutputsFile) AutoFix() {
	var blocks []*OutputBlock
	for i := 0; i < len(f.File.WriteFile.Body().Blocks()); i++ {
		b := BuildOutputBlock(f.File.File, f.File.GetBlock(i), i)
		b.AutoFix()
		blocks = append(blocks, b)
	}

	linq.From(blocks).OrderBy(func(i interface{}) interface{} {
		return i.(*OutputBlock).Block.Labels[0]
	}).ToSlice(&blocks)

	f.File.WriteFile.Body().Clear()
	for i, block := range blocks {
		f.File.WriteFile.Body().AppendBlock(block.Block.WriteBlock)
		if i < len(blocks)-1 {
			f.File.WriteFile.Body().AppendNewline()
		}
	}
}
