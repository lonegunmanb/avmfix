package pkg

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	tfjson "github.com/hashicorp/terraform-json"
)

var _ Block = &ResourceBlock{}
var _ rootBlock = &ResourceBlock{}

// ResourceBlock is the wrapper of a resource Block
type ResourceBlock struct {
	*block
	Type                 string
	TailMetaArgs         *Args
	TailMetaNestedBlocks *NestedBlocks
}

func (b *ResourceBlock) headMetaArgs() *HeadMetaArgs {
	return b.HeadMetaArgs
}

// CheckBlock checks the resource Block and nested Block recursively to find the Block not in order,
// and invoke the emit function on that Block
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

// DefRange gets the definition range of the resource Block
func (b *ResourceBlock) DefRange() hcl.Range {
	return b.HclBlock.DefRange()
}

// BuildResourceBlock Build the root Block wrapper using hclsyntax.Block
func BuildResourceBlock(block *HclBlock, file *hcl.File,
	emitter func(block Block) error) *ResourceBlock {
	resourceType, resourceName := block.Labels[0], block.Labels[1]
	b := &ResourceBlock{
		block: newBlock(resourceName, block, file, []string{block.Type, resourceType}, emitter),
		Type:  resourceType,
	}
	buildArgs(b, block.Attributes())
	buildNestedBlocks(b, block.NestedBlocks())
	return b
}

// CheckOrder checks whether the resourceBlock is sorted
func (b *ResourceBlock) CheckOrder() bool {
	return b.sorted() && b.gaped()
}

func (b *ResourceBlock) AutoFix() {
	for _, nestedBlock := range b.nestedBlocks() {
		nestedBlock.AutoFix()
	}
	blockToFix := b.HclBlock
	attributes := blockToFix.WriteBlock.Body().Attributes()
	nestedBlocks := blockToFix.WriteBlock.Body().Blocks()
	blockToFix.Clear()
	if b.RequiredArgs != nil || b.OptionalArgs != nil {
		blockToFix.writeNewLine()
		blockToFix.writeArgs(b.RequiredArgs, attributes)
		blockToFix.writeArgs(b.OptionalArgs, attributes)
	}
	if len(b.nestedBlocks()) > 0 {
		blockToFix.writeNewLine()
		blockToFix.writeNestedBlocks(b.RequiredNestedBlocks, nestedBlocks)
		blockToFix.writeNestedBlocks(b.OptionalNestedBlocks, nestedBlocks)
	}
}

// ToString prints the sorted resource Block
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
	blockHead := string(b.HclBlock.DefRange().SliceBytes(b.File.Bytes))
	if strings.TrimSpace(txt) == "" {
		txt = fmt.Sprintf("%s {}", blockHead)
	} else {
		txt = fmt.Sprintf("%s {\n%s\n}", blockHead, txt)
	}
	return string(hclwrite.Format([]byte(txt)))
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

func (b *ResourceBlock) addTailMetaNestedBlock(nb *NestedBlock) {
	if b.TailMetaNestedBlocks == nil {
		b.TailMetaNestedBlocks = &NestedBlocks{}
	}
	b.TailMetaNestedBlocks.add(nb)
}
