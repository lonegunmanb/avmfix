package pkg

import (
	"strings"

	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	tfjson "github.com/hashicorp/terraform-json"
)

func buildNestedBlock(parent blockWithSchema, index int, nestedBlock *HclBlock) *NestedBlock {
	nestedBlockName := nestedBlock.Type
	sortField := nestedBlock.Type
	if nestedBlock.Type == "dynamic" {
		nestedBlockName = nestedBlock.Labels[0]
		sortField = strings.Join(nestedBlock.Labels, "")
	}
	path := append(parent.path(), nestedBlockName)
	nb := &NestedBlock{
		resourceBlock: newBlock(nestedBlockName, nestedBlock, parent.file(), path),
		SortField:     sortField,
		Index:         index,
	}
	attributes := nestedBlock.Attributes()
	blocks := nestedBlock.NestedBlocks()
	if nb.BlockType() == "dynamic" {
		linq.From(attributes).Concat(linq.From(nestedBlock.NestedBlocks()[0].Attributes())).ToMap(&attributes)
		blocks = blocks[0].NestedBlocks()
	}
	buildArgs(nb, attributes)
	buildNestedBlocks(nb, blocks)
	return nb
}

var _ blockWithSchema = &NestedBlock{}

// NestedBlock is a wrapper of the nested Block
type NestedBlock struct {
	*resourceBlock
	SortField string
	Index     int
}

// DefRange gets the definition range of the nested Block
func (b *NestedBlock) DefRange() hcl.Range {
	return b.HclBlock.DefRange()
}

func (b *NestedBlock) schemaBlock() *tfjson.SchemaBlock {
	return queryBlockSchema(b.Path)
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
		contentBlock := blockToFix.NestedBlocks()[0]
		// Enforce dynamic blockWithSchema's meta arguments' order
		forEach := blockToFix.Attributes()["for_each"]
		iterator := blockToFix.Attributes()["iterator"]
		// Fix dynamic blockWithSchema then proceed into the content blockWithSchema
		blockToFix.Clear().
			appendNewline().
			appendAttribute(forEach).
			appendAttribute(iterator).
			appendNewline().
			appendBlock(contentBlock.WriteBlock)
		blockToFix = contentBlock
	}
	singleLineBlock := blockToFix.isSingleLineBlock()
	empty := true
	attributes := blockToFix.WriteBlock.Body().Attributes()
	nestedBlocks := blockToFix.WriteBlock.Body().Blocks()
	blockToFix.Clear()
	if b.RequiredArgs != nil || b.OptionalArgs != nil {
		blockToFix.appendNewline()
		empty = false
	}
	blockToFix.writeArgs(b.RequiredArgs.SortByName(), attributes).
		writeArgs(b.OptionalArgs.SortByName(), attributes)
	if len(b.nestedBlocks()) > 0 {
		blockToFix.appendNewline()
		empty = false
	}
	blockToFix.appendNestedBlocks(b.RequiredNestedBlocks, nestedBlocks).
		appendNestedBlocks(b.OptionalNestedBlocks, nestedBlocks)

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

func (b *NestedBlock) isHeadMeta(argNameOrNestedBlockType string) bool {
	if b.BlockType() != "dynamic" {
		return false
	}
	return argNameOrNestedBlockType == "iterator" || argNameOrNestedBlockType == "for_each"
}

func (b *NestedBlock) isTailMeta(argNameOrNestedBlockType string) bool {
	return false
}
