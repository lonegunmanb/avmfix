package pkg

import (
	"github.com/hashicorp/hcl/v2"
	tfjson "github.com/hashicorp/terraform-json"
)

// Block is an interface offering general APIs on resource/nested Block
type blockWithSchema interface {
	file() *hcl.File
	
	path() []string
	schemaBlock() (*tfjson.SchemaBlock, error)
	isHeadMeta(argNameOrNestedBlockType string) bool
	isTailMeta(argNameOrNestedBlockType string) bool
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

func buildArgs(b blockWithSchema, attributes map[string]*HclAttribute) error {
	argSchemas, err := b.schemaBlock()
	if err != nil {
		return err
	}
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
	return nil
}

func buildNestedBlocks(b blockWithSchema, nestedBlocks []*HclBlock) error {
	blockSchema, err := b.schemaBlock()
	if err != nil {
		return err
	}
	for i, nestedBlock := range nestedBlocks {
		nb, err := buildNestedBlock(b, i, nestedBlock)
		if err != nil {
			return err
		}
		rb, rootBlock := b.(rootBlock)
		if rootBlock && b.isTailMeta(nb.Name) {
			rb.addTailMetaNestedBlock(nb)
			continue
		}
		if metaArgOrUnknownBlock(blockSchema) {
			b.addOptionalNestedBlock(nb)
			continue
		}
		nbSchema, knownBlock := blockSchema.NestedBlocks[nb.Name]
		if knownBlock && nbSchema.MinItems > 0 {
			b.addRequiredNestedBlock(nb)
		} else {
			b.addOptionalNestedBlock(nb)
		}
	}
	return nil
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
	Path                 []string
}

func newBlock(name string, b *HclBlock, f *hcl.File, path []string) *resourceBlock {
	return &resourceBlock{
		Name:     name,
		HclBlock: b,
		File:     f,
		Path:     path,
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
