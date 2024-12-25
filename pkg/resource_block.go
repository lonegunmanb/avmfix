package pkg

import (
	"github.com/hashicorp/hcl/v2"
	tfjson "github.com/hashicorp/terraform-json"
)

var _ blockWithSchema = &ResourceBlock{}
var _ rootBlock = &ResourceBlock{}

// ResourceBlock is the wrapper of a resource Block
type ResourceBlock struct {
	*resourceBlock
	Type                 string
	TailMetaArgs         Args
	TailMetaNestedBlocks *NestedBlocks
}

func (b *ResourceBlock) schemaBlock() *tfjson.SchemaBlock {
	return queryBlockSchema(b.path())
}

// BuildBlockWithSchema Build the root Block wrapper using hclsyntax.Block
func BuildBlockWithSchema(block *HclBlock, file *hcl.File) *ResourceBlock {
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
	schemas := blockTypesWithSchema[b.Path[0]]
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
		blockToFix.appendNewline()
		blockToFix.writeArgs(b.HeadMetaArgs.SortHeadMetaArgs(), attributes)
		empty = false
	}
	if b.RequiredArgs != nil || b.OptionalArgs != nil {
		blockToFix.appendNewline()
		blockToFix.writeArgs(b.RequiredArgs.SortByName(), attributes)
		blockToFix.writeArgs(b.OptionalArgs.SortByName(), attributes)
		empty = false
	}
	if b.RequiredNestedBlocks != nil || b.OptionalNestedBlocks != nil {
		blockToFix.appendNewline()
		empty = false
	}
	blockToFix.appendNestedBlocks(b.RequiredNestedBlocks, nestedBlocks)
	blockToFix.appendNestedBlocks(b.OptionalNestedBlocks, nestedBlocks)
	if b.TailMetaArgs != nil {
		blockToFix.appendNewline()
		empty = false
	}
	blockToFix.writeArgs(b.TailMetaArgs.SortByName(), attributes)
	if b.TailMetaNestedBlocks != nil {
		blockToFix.appendNewline()
		empty = false
	}
	blockToFix.appendNestedBlocks(b.TailMetaNestedBlocks, nestedBlocks)

	if singleLineBlock && !empty {
		blockToFix.appendNewline()
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

var headMetaArgPriority = map[string]int{"for_each": 0, "count": 0, "provider": 1}
var tailMetaArgPriority = map[string]int{"lifecycle": 0, "depends_on": 1}

// IsTailMeta checks whether a name represents a type of tail Meta arg
func (b *ResourceBlock) isTailMeta(argNameOrNestedBlockType string) bool {
	_, isTailMeta := tailMetaArgPriority[argNameOrNestedBlockType]
	return isTailMeta
}

// IsHeadMeta checks whether a name represents a type of head Meta arg
func (b *ResourceBlock) isHeadMeta(argNameOrNestedBlockType string) bool {
	_, isHeadMeta := headMetaArgPriority[argNameOrNestedBlockType]
	return isHeadMeta
}
