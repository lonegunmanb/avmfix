package pkg

import (
	"github.com/hashicorp/hcl/v2"
)

type TerraformBlock struct {
	HclBlock               *HclBlock
	RequiredProvidersBlock *HclBlock
	RequiredVersion        *HclAttribute
	providers              Args
	File                   *hcl.File
}

func BuildTerraformBlock(block *HclBlock, file *hcl.File) *TerraformBlock {
	r := &TerraformBlock{
		HclBlock: block,
		File:     file,
	}
	if requiredVersionAttr, ok := block.Attributes()["required_version"]; ok {
		r.RequiredVersion = requiredVersionAttr
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

func (b *TerraformBlock) AutoFix() error {
	if b.RequiredProvidersBlock == nil {
		return nil
	}
	attributes := b.RequiredProvidersBlock.WriteBlock.Body().Attributes()
	b.RequiredProvidersBlock.Clear()
	b.RequiredProvidersBlock.appendNewline()
	b.RequiredProvidersBlock.writeArgs(b.providers.SortByName(), attributes)
	if b.RequiredVersion == nil {
		return nil
	}
	b.HclBlock.WriteBlock.Body().Clear()
	b.HclBlock.appendNewline()
	b.HclBlock.writeArgs([]*Arg{buildAttrArg(b.RequiredVersion, b.File)}, b.HclBlock.WriteBlock.Body().Attributes())
	b.HclBlock.appendNewline()
	b.HclBlock.appendBlock(b.RequiredProvidersBlock.WriteBlock)
	return nil
}
