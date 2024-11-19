package pkg

import (
	"github.com/hashicorp/hcl/v2"
)

type TerraformBlock struct {
	HclBlock               *HclBlock
	RequiredProvidersBlock *HclBlock
	providers              Args
	File                   *hcl.File
}

func BuildTerraformBlock(block *HclBlock, file *hcl.File) *TerraformBlock {
	r := &TerraformBlock{
		HclBlock: block,
		File:     file,
	}
	for _, nb := range block.NestedBlocks() {
		if nb.Type == "required_providers" {
			r.RequiredProvidersBlock = nb
			for _, attribute := range attributesByLines(nb.Attributes()) {
				r.providers = append(r.providers, buildAttrArg(attribute, file))
			}
			break
		}
	}
	return r
}

func (b *TerraformBlock) AutoFix() {
	if b.RequiredProvidersBlock == nil {
		return
	}
	attributes := b.RequiredProvidersBlock.WriteBlock.Body().Attributes()
	b.RequiredProvidersBlock.Clear()
	b.RequiredProvidersBlock.appendNewline()
	b.RequiredProvidersBlock.writeArgs(b.providers.SortByName(), attributes)
}
