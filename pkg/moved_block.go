package pkg

import "github.com/hashicorp/hcl/v2"

type MovedBlock struct {
	HclBlock *HclBlock
	File     *hcl.File
}

func BuildMovedBlock(block *HclBlock, file *HclFile) *MovedBlock {
	return &MovedBlock{
		HclBlock: block,
		File:     file.File,
	}
}

func (b *MovedBlock) AutoFix() error {
	if !b.isComplete() {
		return nil
	}
	attributes := b.HclBlock.WriteBlock.Body().Attributes()
	b.HclBlock.Clear()
	b.HclBlock.appendNewline()

	b.HclBlock.writeArgs([]*Arg{
		buildAttrArg(b.HclBlock.Attributes()["from"], b.File),
		buildAttrArg(b.HclBlock.Attributes()["to"], b.File),
	}, attributes)
	return nil
}

func (b *MovedBlock) isComplete() bool {
	_, ok := b.HclBlock.Attributes()["from"]
	if !ok {
		return false
	}
	_, ok = b.HclBlock.Attributes()["to"]
	return ok
}
