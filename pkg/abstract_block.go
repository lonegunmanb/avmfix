package pkg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type AbstractBlock struct {
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

func newAbstractBlock(name string, b *hclsyntax.Block, f *hcl.File, path []string, emitter func(block Block) error) *AbstractBlock {
	return &AbstractBlock{
		Name:  name,
		Block: b,
		File:  f,
		Path:  path,
		emit:  emitter,
		Range: b.Range(),
	}
}

func (b *AbstractBlock) addHeadMetaArg(arg *Arg) {
	if b.HeadMetaArgs == nil {
		b.HeadMetaArgs = &HeadMetaArgs{}
	}
	b.HeadMetaArgs.add(arg)
}

func (b *AbstractBlock) addRequiredAttr(arg *Arg) {
	if b.RequiredArgs == nil {
		b.RequiredArgs = &Args{}
	}
	b.RequiredArgs.add(arg)
}

func (b *AbstractBlock) addOptionalAttr(arg *Arg) {
	if b.OptionalArgs == nil {
		b.OptionalArgs = &Args{}
	}
	b.OptionalArgs.add(arg)
}

func (b *AbstractBlock) addRequiredNestedBlock(nb *NestedBlock) {
	if b.RequiredNestedBlocks == nil {
		b.RequiredNestedBlocks = &NestedBlocks{}
	}
	b.RequiredNestedBlocks.add(nb)
}

func (b *AbstractBlock) addOptionalNestedBlock(nb *NestedBlock) {
	if b.OptionalNestedBlocks == nil {
		b.OptionalNestedBlocks = &NestedBlocks{}
	}
	b.OptionalNestedBlocks.add(nb)
}

func (b *AbstractBlock) file() *hcl.File {
	return b.File
}

func (b *AbstractBlock) path() []string {
	return b.Path
}

func (b *AbstractBlock) emitter() func(block Block) error {
	return b.emit
}
