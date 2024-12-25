package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type OutputBlock struct {
	Block      *HclBlock
	Attributes Args
}

func BuildOutputBlock(f *hcl.File, b *HclBlock) *OutputBlock {
	r := &OutputBlock{
		Block: b,
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
	b.Block.appendNewline()
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
			return
		}
		b.Attributes = removeIndex(b.Attributes, i)
		return
	}
}

func (b *OutputBlock) sortArguments() {
	b.Attributes = b.Attributes.SortByName()
}

type OutputsFile struct {
	dir  *directory
	File *HclFile
}

func BuildOutputsFile(f *HclFile) *OutputsFile {
	return &OutputsFile{
		dir:  f.dir,
		File: f,
	}
}

func (f *OutputsFile) AutoFix() {
	var blocks []*OutputBlock
	for i := 0; i < len(f.File.WriteFile.Body().Blocks()); i++ {
		block := f.File.GetBlock(i)
		if block.Type != "output" {
			f.dir.AppendBlockToFile("main.tf", block)
			continue
		}
		b := BuildOutputBlock(f.File.File, block)
		b.AutoFix()
		blocks = append(blocks, b)
	}

	linq.From(blocks).OrderBy(func(i interface{}) interface{} {
		return i.(*OutputBlock).Block.Labels[0]
	}).ToSlice(&blocks)

	f.File.ClearWriteFile()

	for i, block := range blocks {
		if i != 0 {
			f.File.appendNewline()
		}
		f.File.appendBlock(block.Block)
		if !endWithNewLine(block.Block.WriteBlock) {
			f.File.appendNewline()
		}
	}
}
