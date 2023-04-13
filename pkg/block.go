package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"strings"
)

type block interface {
	file() *hcl.File
	path() []string
	emitter() func(block Block) error
	isHeadMeta(argName string) bool
	isTailMeta(argName string) bool
	addHeadMetaArg(arg *Arg)
	addOptionalAttr(arg *Arg)
	addRequiredAttr(arg *Arg)
	addOptionalNestedBlock(nb *NestedBlock)
	addRequiredNestedBlock(nb *NestedBlock)
}

type rootBlock interface {
	addTailMetaArg(arg *Arg)
	addTailMetaNestedBlock(nb *NestedBlock)
}

func buildNestedBlock(parent block, nestedBlock *hclsyntax.Block) *NestedBlock {
	nestedBlockName := nestedBlock.Type
	sortField := nestedBlock.Type
	if nestedBlock.Type == "dynamic" {
		nestedBlockName = nestedBlock.Labels[0]
		sortField = strings.Join(nestedBlock.Labels, "")
	}
	path := append(parent.path(), nestedBlockName)
	nb := &NestedBlock{
		AbstractBlock: newAbstractBlock(nestedBlockName, nestedBlock, parent.file(), path, parent.emitter()),
		SortField:     sortField,
	}
	attributes := nestedBlock.Body.Attributes
	blocks := nestedBlock.Body.Blocks
	if nb.BlockType() == "dynamic" {
		linq.From(attributes).Concat(linq.From(nestedBlock.Body.Blocks[0].Body.Attributes)).ToMap(&attributes)
		blocks = nestedBlock.Body.Blocks[0].Body.Blocks
	}
	buildArgs(nb, attributes)
	buildNestedBlocks(nb, blocks)
	return nb
}

func buildArgs(b block, attributes hclsyntax.Attributes) {
	argSchemas := queryBlockSchema(b.path())
	for _, attr := range attributesByLines(attributes) {
		attrName := attr.Name
		arg := buildAttrArg(attr, b.file())
		if b.isHeadMeta(attrName) {
			b.addHeadMetaArg(arg)
			continue
		}
		rb, rootBlock := b.(rootBlock)
		if rootBlock && b.isTailMeta(attrName) {
			rb.addTailMetaArg(arg)
			continue
		}
		if argSchemas == nil {
			b.addOptionalAttr(arg)
			continue
		}
		attrSchema, isAzAttr := argSchemas.Attributes[attrName]
		if isAzAttr && attrSchema.Required {
			b.addRequiredAttr(arg)
		} else {
			b.addOptionalAttr(arg)
		}
	}
}

func buildNestedBlocks(b block, nestedBlocks hclsyntax.Blocks) {
	blockSchema := queryBlockSchema(b.path())
	for _, nestedBlock := range nestedBlocks {
		nb := buildNestedBlock(b, nestedBlock)
		rb, rootBlock := b.(rootBlock)
		if rootBlock && b.isTailMeta(nb.Name) {
			rb.addTailMetaNestedBlock(nb)
			continue
		}
		if metaArgOrUnknownBlock(blockSchema) {
			b.addOptionalNestedBlock(nb)
			continue
		}
		nbSchema, isAzNestedBlock := blockSchema.NestedBlocks[nb.Name]
		if isAzNestedBlock && nbSchema.MinItems > 0 {
			b.addRequiredNestedBlock(nb)
		} else {
			b.addOptionalNestedBlock(nb)
		}
	}
}
