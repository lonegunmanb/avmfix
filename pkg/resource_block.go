package pkg

import (
	tfjson "github.com/hashicorp/terraform-json"
)

var _ blockWithSchema = &ResourceBlock{}
var _ rootBlock = &ResourceBlock{}

// ResourceBlock is the wrapper of a resource Block
type ResourceBlock struct {
	*resourceBlock
	namespace            string
	version              string
	Type                 string
	TailMetaArgs         Args
	TailMetaNestedBlocks *NestedBlocks
}

func (b *ResourceBlock) getProviderNamespace() string {
	return b.namespace
}

func (b *ResourceBlock) getProviderVersion() string {
	return b.version
}

func (b *ResourceBlock) schemaBlock() (*tfjson.SchemaBlock, error) {
	return queryBlockSchema(b.path(), b.namespace, b.version)
}

var resolveNamespace = func(resourceType string, file *HclFile) (string, error) {
	return file.dir.resolveNamespace(resourceType)
}
var resolveProviderVersion = func(namespace, resourceType string, file *HclFile) (string, error) {
	return file.dir.resolveProviderVersion(namespace, resourceType)
}

// BuildBlockWithSchema Build the root Block wrapper using hclsyntax.Block
func BuildBlockWithSchema(block *HclBlock, file *HclFile) (*ResourceBlock, error) {
	resourceType, resourceName := block.Labels[0], block.Labels[1]
	namespace, err := resolveNamespace(resourceType, file)
	if err != nil {
		return nil, err
	}
	version, err := resolveProviderVersion(namespace, resourceType, file)
	if err != nil {
		return nil, err
	}
	b := &ResourceBlock{
		resourceBlock: newBlock(resourceName, block, file.File, []string{block.Type, resourceType}),
		namespace:     namespace,
		version:       version,
		Type:          resourceType,
	}
	err = buildArgs(b, block.Attributes())
	if err != nil {
		return nil, err
	}
	err = buildNestedBlocks(b, block.NestedBlocks())
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *ResourceBlock) AutoFix() error {
	for _, nestedBlock := range b.nestedBlocks() {
		if err := nestedBlock.AutoFix(); err != nil {
			return err
		}
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
	return nil
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
