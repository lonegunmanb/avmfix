package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"strings"
)

// Block is an interface offering general APIs on resource/nested Block
type block interface {
	file() *hcl.File
	path() []string
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

func buildNestedBlock(parent block, index int, nestedBlock *HclBlock) *NestedBlock {
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

func buildArgs(b block, attributes map[string]*HclAttribute) {
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

func buildNestedBlocks(b block, nestedBlocks []*HclBlock) {
	blockSchema := queryBlockSchema(b.path())
	for i, nestedBlock := range nestedBlocks {
		nb := buildNestedBlock(b, i, nestedBlock)
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

type resourceBlock struct {
	HclBlock             *HclBlock
	Name                 string
	HeadMetaArgs         Args
	RequiredArgs         Args
	OptionalArgs         Args
	RequiredNestedBlocks *NestedBlocks
	OptionalNestedBlocks *NestedBlocks
	File                 *hcl.File
	Range                hcl.Range
	Path                 []string
}

func newBlock(name string, b *HclBlock, f *hcl.File, path []string) *resourceBlock {
	return &resourceBlock{
		Name:     name,
		HclBlock: b,
		File:     f,
		Path:     path,
		Range:    b.Range(),
	}
}

func (b *resourceBlock) addHeadMetaArg(arg *Arg) {
	b.HeadMetaArgs = append(b.HeadMetaArgs, arg)
}

func (b *resourceBlock) addRequiredAttr(arg *Arg) {
	b.RequiredArgs = append(b.RequiredArgs, arg)
}

func (b *resourceBlock) addOptionalAttr(arg *Arg) {
	b.OptionalArgs = append(b.OptionalArgs, arg)
}

func (b *resourceBlock) addRequiredNestedBlock(nb *NestedBlock) {
	if b.RequiredNestedBlocks == nil {
		b.RequiredNestedBlocks = &NestedBlocks{}
	}
	b.RequiredNestedBlocks.add(nb)
}

func (b *resourceBlock) addOptionalNestedBlock(nb *NestedBlock) {
	if b.OptionalNestedBlocks == nil {
		b.OptionalNestedBlocks = &NestedBlocks{}
	}
	b.OptionalNestedBlocks.add(nb)
}

func (b *resourceBlock) file() *hcl.File {
	return b.File
}

func (b *resourceBlock) path() []string {
	return b.Path
}
