package pkg

import (
	"github.com/hashicorp/hcl/v2"
)

// NestedBlock is a wrapper of the nested Block
type NestedBlock struct {
	*resourceBlock
	SortField string
	Index     int
}

var _ block = &NestedBlock{}

// DefRange gets the definition range of the nested Block
func (b *NestedBlock) DefRange() hcl.Range {
	return b.HclBlock.DefRange()
}

// NestedBlocks is the collection of nestedBlocks with the same type
type NestedBlocks struct {
	Blocks []*NestedBlock
	Range  *hcl.Range
}

// GetRange returns the entire range of this type of nestedBlocks
func (b *NestedBlocks) GetRange() *hcl.Range {
	if b == nil {
		return nil
	}
	return b.Range
}

func (b *NestedBlock) BlockType() string {
	return b.HclBlock.Type
}

func (b *NestedBlock) AutoFix() {
	for _, nestedBlock := range b.nestedBlocks() {
		nestedBlock.AutoFix()
	}
	blockToFix := b.HclBlock
	if b.BlockType() == "dynamic" {
		blockToFix = blockToFix.NestedBlocks()[0]
	}
	singleLineBlock := blockToFix.isSingleLineBlock()
	empty := true
	attributes := blockToFix.WriteBlock.Body().Attributes()
	nestedBlocks := blockToFix.WriteBlock.Body().Blocks()
	blockToFix.Clear()
	if b.RequiredArgs != nil || b.OptionalArgs != nil {
		blockToFix.appendNewline()
		blockToFix.writeArgs(b.RequiredArgs.SortByName(), attributes)
		blockToFix.writeArgs(b.OptionalArgs.SortByName(), attributes)
		empty = false
	}
	if len(b.nestedBlocks()) > 0 {
		blockToFix.appendNewline()
		blockToFix.appendNestedBlocks(b.RequiredNestedBlocks, nestedBlocks)
		blockToFix.appendNestedBlocks(b.OptionalNestedBlocks, nestedBlocks)
		empty = false
	}

	if singleLineBlock && !empty {
		blockToFix.appendNewline()
	}
}

func (b *NestedBlocks) add(arg *NestedBlock) {
	b.Blocks = append(b.Blocks, arg)
}

func (b *NestedBlock) nestedBlocks() []*NestedBlock {
	var nbs []*NestedBlock
	for _, subNbs := range []*NestedBlocks{b.RequiredNestedBlocks, b.OptionalNestedBlocks} {
		if subNbs != nil {
			nbs = append(nbs, subNbs.Blocks...)
		}
	}
	return nbs
}

func (b *NestedBlock) isHeadMeta(argName string) bool {
	return b.BlockType() == "dynamic" && argName == "for_each"
}

func (b *NestedBlock) isTailMeta(argName string) bool {
	return false
}
