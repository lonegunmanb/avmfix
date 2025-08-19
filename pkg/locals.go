package pkg

import "github.com/hashicorp/hcl/v2"

type LocalsBlock struct {
	HclBlock   *HclBlock
	Attributes Args
	File       *hcl.File
}

func BuildLocalsBlock(block *HclBlock, file *HclFile) *LocalsBlock {
	r := &LocalsBlock{
		HclBlock: block,
		File:     file.File,
	}
	for _, attribute := range attributesByLines(block.Attributes()) {
		r.Attributes = append(r.Attributes, buildAttrArg(attribute, file.File))
	}
	return r
}

func (b *LocalsBlock) AutoFix() error {
	attributes := b.HclBlock.WriteBlock.Body().Attributes()
	b.HclBlock.Clear()
	b.HclBlock.appendNewline()
	b.HclBlock.writeArgs(b.Attributes.SortByName(), attributes)
	return nil
}
