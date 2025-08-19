package pkg

import (
	"github.com/hashicorp/hcl/v2"
)

type RemovedBlock struct {
	HclBlock *HclBlock
	File     *hcl.File
}

func BuildRemovedBlock(b *HclBlock, f *HclFile) AutoFixBlock {
	return &RemovedBlock{
		HclBlock: b,
		File:     f.File,
	}
}

func (r *RemovedBlock) AutoFix() error {
	from, ok := r.HclBlock.Attributes()["from"]
	if !ok {
		return nil
	}
	var lifecycle *HclBlock
	var provisioners []*HclBlock
	for _, nb := range r.HclBlock.NestedBlocks() {
		if nb.Type == "lifecycle" {
			lifecycle = nb
			continue
		}
		if nb.Type == "provisioner" {
			provisioners = append(provisioners, nb)
		}
	}
	if lifecycle == nil {
		return nil
	}
	writeAttrs := r.HclBlock.WriteBlock.Body().Attributes()
	hb := r.HclBlock.Clear().
		appendNewline().
		writeArgs(Args{&Arg{
			Name:         "from",
			File:         r.File,
			HclAttribute: from,
		}}, writeAttrs).
		appendNewline().
		appendBlock(lifecycle.WriteBlock)
	if len(provisioners) > 0 {
		hb = hb.appendNewline()
	}
	for _, pb := range provisioners {
		hb.appendBlock(pb.WriteBlock)
	}
	return nil
}
