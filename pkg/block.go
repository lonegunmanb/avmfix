package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"strings"
)

// Block is an interface offering general APIs on resource/nested Block
type Block interface {
	// CheckBlock checks the resourceBlock/nestedBlock recursively to find the Block not in order,
	// and invoke the emit function on that Block
	CheckBlock() error

	// ToString prints the sorted Block
	ToString() string

	// DefRange gets the definition range of the Block
	DefRange() hcl.Range

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

func buildNestedBlock(parent Block, nestedBlock *hclsyntax.Block) *NestedBlock {
	nestedBlockName := nestedBlock.Type
	sortField := nestedBlock.Type
	if nestedBlock.Type == "dynamic" {
		nestedBlockName = nestedBlock.Labels[0]
		sortField = strings.Join(nestedBlock.Labels, "")
	}
	path := append(parent.path(), nestedBlockName)
	nb := &NestedBlock{
		block:     newBlock(nestedBlockName, nestedBlock, parent.file(), path, parent.emitter()),
		SortField: sortField,
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

func buildArgs(b Block, attributes hclsyntax.Attributes) {
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

func buildNestedBlocks(b Block, nestedBlocks hclsyntax.Blocks) {
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

type block struct {
	Block                *hclsyntax.Block
	Name                 string
	HeadMetaArgs         *HeadMetaArgs
	RequiredArgs         *Args
	OptionalArgs         *Args
	RequiredNestedBlocks *NestedBlocks
	OptionalNestedBlocks *NestedBlocks
	File                 *hcl.File
	Range                hcl.Range
	Path                 []string
	emit                 func(block Block) error
}

func newBlock(name string, b *hclsyntax.Block, f *hcl.File, path []string, emitter func(block Block) error) *block {
	return &block{
		Name:  name,
		Block: b,
		File:  f,
		Path:  path,
		emit:  emitter,
		Range: b.Range(),
	}
}

func (b *block) addHeadMetaArg(arg *Arg) {
	if b.HeadMetaArgs == nil {
		b.HeadMetaArgs = &HeadMetaArgs{}
	}
	b.HeadMetaArgs.add(arg)
}

func (b *block) addRequiredAttr(arg *Arg) {
	if b.RequiredArgs == nil {
		b.RequiredArgs = &Args{}
	}
	b.RequiredArgs.add(arg)
}

func (b *block) addOptionalAttr(arg *Arg) {
	if b.OptionalArgs == nil {
		b.OptionalArgs = &Args{}
	}
	b.OptionalArgs.add(arg)
}

func (b *block) addRequiredNestedBlock(nb *NestedBlock) {
	if b.RequiredNestedBlocks == nil {
		b.RequiredNestedBlocks = &NestedBlocks{}
	}
	b.RequiredNestedBlocks.add(nb)
}

func (b *block) addOptionalNestedBlock(nb *NestedBlock) {
	if b.OptionalNestedBlocks == nil {
		b.OptionalNestedBlocks = &NestedBlocks{}
	}
	b.OptionalNestedBlocks.add(nb)
}

func (b *block) file() *hcl.File {
	return b.File
}

func (b *block) path() []string {
	return b.Path
}

func (b *block) emitter() func(block Block) error {
	return b.emit
}
