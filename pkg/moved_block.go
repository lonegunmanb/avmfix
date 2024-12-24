package pkg

import "github.com/hashicorp/hcl/v2"

type MovedBlock struct {
	HclBlock *HclBlock
	File     *hcl.File
}

func BuildMovedBlock(block *HclBlock, file *hcl.File) *MovedBlock {
	return &MovedBlock{
		HclBlock: block,
		File:     file,
	}
}

func (b *MovedBlock) AutoFix() {
	if !b.isComplete() {
		return
	}
	attributes := b.HclBlock.WriteBlock.Body().Attributes()
	b.HclBlock.Clear()
	b.HclBlock.appendNewline()
	
	b.HclBlock.writeArgs([]*Arg{
		buildAttrArg(b.HclBlock.Attributes()["from"], b.File),
		buildAttrArg(b.HclBlock.Attributes()["to"], b.File),
	}, attributes)
}

func (b *MovedBlock) isComplete() bool {
	_, ok := b.HclBlock.Attributes()["from"]
	if !ok {
		return false
	}
	_, ok = b.HclBlock.Attributes()["to"]
	return ok
}
