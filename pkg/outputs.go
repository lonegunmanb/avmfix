package pkg

import (
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
	for i := 0; i < len(f.File.WriteFile.Body().Blocks()); i++ {
		b := BuildOutputBlock(f.File.File, f.File.GetBlock(i), i)
		b.AutoFix()
	}
}
