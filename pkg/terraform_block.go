package pkg

import (
	"github.com/hashicorp/hcl/v2"
)

type TerraformBlock struct {
	HclBlock               *HclBlock
	RequiredProvidersBlock *HclBlock
	RequiredVersion        *HclAttribute
	Experiments            *HclAttribute
	Backend                *HclBlock
	Cloud                  *HclBlock
	ProviderMeta           *HclBlock
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
	if experimentsAttr, ok := block.Attributes()["experiments"]; ok {
		r.Experiments = experimentsAttr
	}
	for _, nb := range block.NestedBlocks() {
		if nb.Type == "required_providers" {
			r.RequiredProvidersBlock = nb
			for _, attribute := range attributesByLines(nb.Attributes()) {
				r.providers = append(r.providers, buildAttrArg(attribute, file))
			}
			continue
		}
		if nb.Type == "backend" {
			r.Backend = NewHclBlock(nb.Block, nb.WriteBlock)
			continue
		}
		if nb.Type == "cloud" {
			r.Cloud = NewHclBlock(nb.Block, nb.WriteBlock)
			continue
		}
		if nb.Type == "provider_meta" {
			r.ProviderMeta = NewHclBlock(nb.Block, nb.WriteBlock)
			continue
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
	args := []*Arg{buildAttrArg(b.RequiredVersion, b.File)}
	if b.Experiments != nil {
		args = append(args, buildAttrArg(b.Experiments, b.File))
	}
	b.HclBlock.writeArgs(args, b.HclBlock.WriteBlock.Body().Attributes())
	b.HclBlock.appendNewline()
	if b.Backend != nil {
		b.HclBlock.appendBlock(b.Backend.WriteBlock)
	}
	if b.Cloud != nil {
		b.HclBlock.appendBlock(b.Cloud.WriteBlock)
	}
	if b.ProviderMeta != nil {
		b.HclBlock.appendBlock(b.ProviderMeta.WriteBlock)
	}
	b.HclBlock.appendBlock(b.RequiredProvidersBlock.WriteBlock)
	return nil
}
