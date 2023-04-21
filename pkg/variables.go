package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var variableAttributePriorities = map[string]int{
	"type":        0,
	"default":     1,
	"description": 2,
	"nullable":    3,
	"sensitive":   4,
}

type VariablesFile struct {
	File *HclFile
}

func BuildVariablesFile(f *HclFile) *VariablesFile {
	return &VariablesFile{
		File: f,
	}
}

func (f *VariablesFile) AutoFix() {
	requiredVariables := []*VariableBlock{}
	optionalVariables := []*VariableBlock{}
	for i := 0; i < len(f.File.WriteFile.Body().Blocks()); i++ {
		b := BuildVariableBlock(f.File.File, f.File.GetBlock(i))
		b.AutoFix()
		nullable, ok := b.Block.Attributes()["nullable"]
		if !ok || nullable.IsNullable() {
			optionalVariables = append(optionalVariables, b)
			continue
		}
		requiredVariables = append(requiredVariables, b)
	}
	linq.From(optionalVariables).OrderBy(func(i interface{}) interface{} {
		return i.(*VariableBlock).Block.Labels[0]
	}).ToSlice(&optionalVariables)
	linq.From(requiredVariables).OrderBy(func(i interface{}) interface{} {
		return i.(*VariableBlock).Block.Labels[0]
	}).ToSlice(&requiredVariables)
	f.File.WriteFile.Body().Clear()
	for i, variableBlock := range requiredVariables {
		f.File.WriteFile.Body().AppendBlock(variableBlock.Block.WriteBlock)
		if i < len(requiredVariables)-1 {
			f.File.WriteFile.Body().AppendNewline()
		}
	}

	if len(requiredVariables) > 0 && len(optionalVariables) > 0 {
		f.File.WriteFile.Body().AppendNewline()
	}

	for i, variableBlock := range optionalVariables {
		f.File.WriteFile.Body().AppendBlock(variableBlock.Block.WriteBlock)
		if i < len(requiredVariables)-1 {
			f.File.WriteFile.Body().AppendNewline()
		}
	}
}

type VariableBlock struct {
	Block      *HclBlock
	Attributes Args
}

func BuildVariableBlock(f *hcl.File, b *HclBlock) *VariableBlock {
	r := &VariableBlock{
		Block: b,
	}
	for _, attribute := range attributesByLines(b.Attributes()) {
		r.Attributes = append(r.Attributes, buildAttrArg(attribute, f))
	}
	return r
}

func (b *VariableBlock) AutoFix() {
	b.sortArguments()
	b.removeUnnecessaryNullable()
	b.removeUnnecessarySensitive()
	b.write()
}

func (b *VariableBlock) sortArguments() {
	linq.From(b.Attributes).OrderBy(func(i interface{}) interface{} {
		attr := i.(*Arg)
		return variableAttributePriorities[attr.Name]
	}).ToSlice(&b.Attributes)
}

func (b *VariableBlock) write() {
	attributes := b.Block.WriteBlock.Body().Attributes()
	blocks := b.Block.WriteBlock.Body().Blocks()
	b.Block.Clear()
	b.Block.writeNewLine()
	b.Block.writeArgs(b.Attributes, attributes)
	if len(blocks) > 0 {
		validationBlock := blocks[0]
		b.Block.writeNewLine()
		b.Block.appendBlock(validationBlock)
	}
}

func (b *VariableBlock) removeUnnecessaryNullable() {
	for i := 0; i < len(b.Attributes); i++ {
		attr := b.Attributes[i]
		if attr.Name != "nullable" {
			continue
		}
		literal, ok := attr.Attribute.Expr.(*hclsyntax.LiteralValueExpr)
		if ok && literal.Val.True() {
			b.Attributes = removeIndex(b.Attributes, i)
		}
		return
	}
}

func (b *VariableBlock) removeUnnecessarySensitive() {
	for i := 0; i < len(b.Attributes); i++ {
		attr := b.Attributes[i]
		if attr.Name != "sensitive" {
			continue
		}
		literal, ok := attr.Attribute.Expr.(*hclsyntax.LiteralValueExpr)
		if ok && literal.Val.False() {
			b.Attributes = removeIndex(b.Attributes, i)
		}
		return
	}
}
