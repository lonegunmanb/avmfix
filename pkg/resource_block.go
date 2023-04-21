package pkg

import (
	"github.com/hashicorp/hcl/v2"
	tfjson "github.com/hashicorp/terraform-json"
)

var _ block = &ResourceBlock{}
var _ rootBlock = &ResourceBlock{}

// ResourceBlock is the wrapper of a resource Block
type ResourceBlock struct {
	*resourceBlock
	Type                 string
	TailMetaArgs         Args
	TailMetaNestedBlocks *NestedBlocks
}

func (b *ResourceBlock) headMetaArgs() Args {
	return b.HeadMetaArgs
}

// BuildResourceBlock Build the root Block wrapper using hclsyntax.Block
func BuildResourceBlock(block *HclBlock, file *hcl.File) *ResourceBlock {
	resourceType, resourceName := block.Labels[0], block.Labels[1]
	b := &ResourceBlock{
		resourceBlock: newBlock(resourceName, block, file, []string{block.Type, resourceType}),
		Type:          resourceType,
	}
	buildArgs(b, block.Attributes())
	buildNestedBlocks(b, block.NestedBlocks())
	return b
}

func (b *ResourceBlock) AutoFix() {
	schemas := resourceSchemas
	if b.Path[0] == "data" {
		schemas = dataSourceSchemas
	}
	_, ok := schemas[b.Type]
	if !ok {
		return
	}
	for _, nestedBlock := range b.nestedBlocks() {
		nestedBlock.AutoFix()
	}
	blockToFix := b.HclBlock
	singleLineBlock := blockToFix.isSingleLineBlock()
	empty := true
	attributes := blockToFix.WriteBlock.Body().Attributes()
	nestedBlocks := blockToFix.WriteBlock.Body().Blocks()
	blockToFix.Clear()
	if b.HeadMetaArgs != nil {
		blockToFix.writeNewLine()
		blockToFix.writeArgs(b.HeadMetaArgs.SortByName(), attributes)
		empty = false
	}
	if b.RequiredArgs != nil || b.OptionalArgs != nil {
		blockToFix.writeNewLine()
		blockToFix.writeArgs(b.RequiredArgs.SortByName(), attributes)
		blockToFix.writeArgs(b.OptionalArgs.SortByName(), attributes)
		empty = false
	}
	if b.RequiredNestedBlocks != nil || b.OptionalNestedBlocks != nil {
		blockToFix.writeNewLine()
		blockToFix.writeNestedBlocks(b.RequiredNestedBlocks, nestedBlocks)
		blockToFix.writeNestedBlocks(b.OptionalNestedBlocks, nestedBlocks)
		empty = false
	}
	if b.TailMetaArgs != nil {
		blockToFix.writeNewLine()
		blockToFix.writeArgs(b.TailMetaArgs.SortByName(), attributes)
		empty = false
	}
	if b.TailMetaNestedBlocks != nil {
		blockToFix.writeNewLine()
		blockToFix.writeNestedBlocks(b.TailMetaNestedBlocks, nestedBlocks)
		empty = false
	}

	if singleLineBlock && !empty {
		blockToFix.writeNewLine()
	}
}

func (b *ResourceBlock) nestedBlocks() []*NestedBlock {
	var nbs []*NestedBlock
	for _, nb := range []*NestedBlocks{
		b.RequiredNestedBlocks,
		b.OptionalNestedBlocks,
		b.TailMetaNestedBlocks} {
		if nb != nil {
			nbs = append(nbs, nb.Blocks...)
		}
	}
	return nbs
}

func metaArgOrUnknownBlock(blockSchema *tfjson.SchemaBlock) bool {
	return blockSchema == nil || blockSchema.NestedBlocks == nil
}

func (b *ResourceBlock) addTailMetaArg(arg *Arg) {
	b.TailMetaArgs = append(b.TailMetaArgs, arg)
}

func (b *ResourceBlock) addTailMetaNestedBlock(nb *NestedBlock) {
	if b.TailMetaNestedBlocks == nil {
		b.TailMetaNestedBlocks = &NestedBlocks{}
	}
	b.TailMetaNestedBlocks.add(nb)
}
