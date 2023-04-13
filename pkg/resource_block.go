package pkg

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	tfjson "github.com/hashicorp/terraform-json"
	"strings"
)

// Block is an interface offering general APIs on resource/nested block
type Block interface {
	// CheckBlock checks the resourceBlock/nestedBlock recursively to find the block not in order,
	// and invoke the emit function on that block
	CheckBlock() error

	// ToString prints the sorted block
	ToString() string

	// DefRange gets the definition range of the block
	DefRange() hcl.Range

	getSyntaxAttribute(name string) *hclsyntax.Attribute
	getWriteAttribute(name string) *hclwrite.Attribute
}

var _ block = &ResourceBlock{}
var _ rootBlock = &ResourceBlock{}

// ResourceBlock is the wrapper of a resource block
type ResourceBlock struct {
	*AbstractBlock
	Name                 string
	Type                 string
	File                 *hcl.File
	writeFile            *hclwrite.File
	Block                *hclsyntax.Block
	HeadMetaArgs         *HeadMetaArgs
	RequiredArgs         *Args
	OptionalArgs         *Args
	RequiredNestedBlocks *NestedBlocks
	OptionalNestedBlocks *NestedBlocks
	TailMetaArgs         *Args
	TailMetaNestedBlocks *NestedBlocks
	Path                 []string
	emit                 func(block Block) error
	writeBlock           *hclwrite.Block
}

func (b *ResourceBlock) headMetaArgs() *HeadMetaArgs {
	return b.HeadMetaArgs
}

// CheckBlock checks the resource block and nested block recursively to find the block not in order,
// and invoke the emit function on that block
func (b *ResourceBlock) CheckBlock() error {
	if !b.CheckOrder() {
		return b.emit(b)
	}
	var err error
	for _, nb := range b.nestedBlocks() {
		if subErr := nb.CheckBlock(); subErr != nil {
			err = multierror.Append(err, subErr)
		}
	}
	return err
}

// DefRange gets the definition range of the resource block
func (b *ResourceBlock) DefRange() hcl.Range {
	return b.Block.DefRange()
}

// BuildResourceBlock Build the root block wrapper using hclsyntax.Block
func BuildResourceBlock(block *hclsyntax.Block, file *hcl.File,
	emitter func(block Block) error) *ResourceBlock {
	wFile, _ := hclwrite.ParseConfig(file.Bytes, "", hcl.InitialPos)
	wBlock := wFile.Body().FirstMatchingBlock(block.Type, block.Labels)
	b := &ResourceBlock{
		AbstractBlock: &AbstractBlock{},
		Type:          block.Labels[0],
		Name:          block.Labels[1],
		File:          file,
		writeFile:     wFile,
		writeBlock:    wBlock,
		Block:         block,
		Path:          []string{block.Type, block.Labels[0]},
		emit:          emitter,
	}
	buildArgs(b, block.Body.Attributes)
	buildNestedBlocks(b, block.Body.Blocks)
	return b
}

// CheckOrder checks whether the resourceBlock is sorted
func (b *ResourceBlock) CheckOrder() bool {
	return b.sorted() && b.gaped()
}

// ToString prints the sorted resource block
func (b *ResourceBlock) ToString() string {
	headMetaTxt := toString(b.HeadMetaArgs)
	argTxt := toString(b.RequiredArgs, b.OptionalArgs)
	nbTxt := toString(b.RequiredNestedBlocks, b.OptionalNestedBlocks)
	tailMetaArgTxt := toString(b.TailMetaArgs)
	tailMetaNbTxt := toString(b.TailMetaNestedBlocks)
	var txts []string
	for _, subTxt := range []string{
		headMetaTxt,
		argTxt,
		nbTxt,
		tailMetaArgTxt,
		tailMetaNbTxt} {
		if subTxt != "" {
			txts = append(txts, subTxt)
		}
	}
	txt := strings.Join(txts, "\n\n")
	blockHead := string(b.Block.DefRange().SliceBytes(b.File.Bytes))
	if strings.TrimSpace(txt) == "" {
		txt = fmt.Sprintf("%s {}", blockHead)
	} else {
		txt = fmt.Sprintf("%s {\n%s\n}", blockHead, txt)
	}
	return string(hclwrite.Format([]byte(txt)))
}

func (b *ResourceBlock) getSyntaxAttribute(name string) *hclsyntax.Attribute {
	return b.Block.Body.Attributes[name]
}

func (b *ResourceBlock) getWriteAttribute(name string) *hclwrite.Attribute {
	return b.writeBlock.Body().GetAttribute(name)
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

func (b *ResourceBlock) sorted() bool {
	sections := []Section{
		b.HeadMetaArgs,
		b.RequiredArgs,
		b.OptionalArgs,
		b.RequiredNestedBlocks,
		b.OptionalNestedBlocks,
		b.TailMetaArgs,
		b.TailMetaNestedBlocks,
	}
	lastEndLine := -1
	for _, s := range sections {
		if !s.CheckOrder() {
			return false
		}
		r := s.GetRange()
		if r == nil {
			continue
		}
		if r.Start.Line <= lastEndLine {
			return false
		}
		lastEndLine = r.End.Line
	}
	return true
}

func (b *ResourceBlock) gaped() bool {
	ranges := []*hcl.Range{
		b.HeadMetaArgs.GetRange(),
		mergeRange(b.RequiredArgs, b.OptionalArgs),
		mergeRange(b.RequiredNestedBlocks, b.OptionalNestedBlocks),
		b.TailMetaArgs.GetRange(),
		b.TailMetaNestedBlocks.GetRange(),
	}
	lastEndLine := -2
	for _, r := range ranges {
		if r == nil {
			continue
		}
		if r.Start.Line-lastEndLine < 2 {
			return false
		}
		lastEndLine = r.End.Line
	}
	return true
}

func (b *ResourceBlock) addTailMetaArg(arg *Arg) {
	if b.TailMetaArgs == nil {
		b.TailMetaArgs = &Args{}
	}
	b.TailMetaArgs.add(arg)
}

func (b *ResourceBlock) addRequiredAttr(arg *Arg) {
	if b.RequiredArgs == nil {
		b.RequiredArgs = &Args{}
	}
	b.RequiredArgs.add(arg)
}

func (b *ResourceBlock) addOptionalAttr(arg *Arg) {
	if b.OptionalArgs == nil {
		b.OptionalArgs = &Args{}
	}
	b.OptionalArgs.add(arg)
}

func (b *ResourceBlock) addTailMetaNestedBlock(nb *NestedBlock) {
	if b.TailMetaNestedBlocks == nil {
		b.TailMetaNestedBlocks = &NestedBlocks{}
	}
	b.TailMetaNestedBlocks.add(nb)
}

func (b *ResourceBlock) addRequiredNestedBlock(nb *NestedBlock) {
	if b.RequiredNestedBlocks == nil {
		b.RequiredNestedBlocks = &NestedBlocks{}
	}
	b.RequiredNestedBlocks.add(nb)
}

func (b *ResourceBlock) addOptionalNestedBlock(nb *NestedBlock) {
	if b.OptionalNestedBlocks == nil {
		b.OptionalNestedBlocks = &NestedBlocks{}
	}
	b.OptionalNestedBlocks.add(nb)
}

func (b *ResourceBlock) file() *hcl.File {
	return b.File
}

func (b *ResourceBlock) path() []string {
	return b.Path
}

func (b *ResourceBlock) emitter() func(block Block) error {
	return b.emit
}
