package pkg

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// NestedBlock is a wrapper of the nested Block
type NestedBlock struct {
	*block
	writeBlock *hclwrite.Block
	SortField  string
}

var _ Block = &NestedBlock{}

// CheckBlock checks the nestedBlock recursively to find the Block not in order,
// and invoke the emit function on that Block
func (b *NestedBlock) CheckBlock() error {
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

// DefRange gets the definition range of the nested Block
func (b *NestedBlock) DefRange() hcl.Range {
	return b.Block.DefRange()
}

// CheckOrder checks whether the nestedBlock is sorted
func (b *NestedBlock) CheckOrder() bool {
	return b.checkSubSectionOrder() && b.checkGap()
}

// ToString prints the sorted Block
func (b *NestedBlock) ToString() string {
	headMeta := toString(b.HeadMetaArgs)
	args := toString(b.RequiredArgs, b.OptionalArgs)
	nb := toString(b.RequiredNestedBlocks, b.OptionalNestedBlocks)
	var codes []string
	for _, c := range []string{headMeta, args, nb} {
		if c != "" {
			codes = append(codes, c)
		}
	}
	code := strings.Join(codes, "\n\n")
	blockHead := string(b.Block.DefRange().SliceBytes(b.File.Bytes))
	if strings.TrimSpace(code) == "" {
		code = fmt.Sprintf("%s {}", blockHead)
	} else {
		code = fmt.Sprintf("%s {\n%s\n}", blockHead, code)
	}
	return string(hclwrite.Format([]byte(code)))
}

// NestedBlocks is the collection of nestedBlocks with the same type
type NestedBlocks struct {
	Blocks []*NestedBlock
	Range  *hcl.Range
}

// CheckOrder checks whether this type of nestedBlocks are sorted
func (b *NestedBlocks) CheckOrder() bool {
	if b == nil {
		return true
	}
	var sortField *string
	for _, nb := range b.Blocks {
		if sortField != nil && *sortField > nb.SortField {
			return false
		}
		sortField = &nb.SortField
		if !nb.CheckOrder() {
			return false
		}
	}
	return true
}

// ToString prints this type of nestedBlocks in order
func (b *NestedBlocks) ToString() string {
	if b == nil {
		return ""
	}
	sortedBlocks := make([]*NestedBlock, len(b.Blocks))
	copy(sortedBlocks, b.Blocks)
	sort.Slice(sortedBlocks, func(i, j int) bool {
		return sortedBlocks[i].SortField < sortedBlocks[j].SortField
	})
	var lines []string
	for _, nb := range sortedBlocks {
		lines = append(lines, nb.ToString())
	}
	return string(hclwrite.Format([]byte(strings.Join(lines, "\n"))))
}

// GetRange returns the entire range of this type of nestedBlocks
func (b *NestedBlocks) GetRange() *hcl.Range {
	if b == nil {
		return nil
	}
	return b.Range
}

func (b *NestedBlock) BlockType() string {
	return b.Block.Type
}

func (b *NestedBlocks) add(arg *NestedBlock) {
	b.Blocks = append(b.Blocks, arg)
	if b.Range == nil {
		b.Range = &hcl.Range{
			Filename: arg.Range.Filename,
			Start:    hcl.Pos{Line: math.MaxInt},
			End:      hcl.Pos{Line: -1},
		}
	}
	if b.Range.Start.Line > arg.Range.Start.Line {
		b.Range.Start = arg.Range.Start
	}
	if b.Range.End.Line < arg.Range.End.Line {
		b.Range.End = arg.Range.End
	}
}

func (b *NestedBlock) getSyntaxAttribute(name string) *hclsyntax.Attribute {
	return b.Block.Body.Attributes[name]
}

func (b *NestedBlock) getWriteAttribute(name string) *hclwrite.Attribute {
	return b.writeBlock.Body().GetAttribute(name)
}

func (b *NestedBlock) nestedBlocks() []*NestedBlock {
	var nbs []*NestedBlock
	for _, subNbs := range []*NestedBlocks{b.RequiredNestedBlocks, b.OptionalNestedBlocks} {
		if subNbs != nil {
			nbs = append(nbs, subNbs.Blocks...)
		}
	}
	return nbs
}

func (b *NestedBlock) checkSubSectionOrder() bool {
	sections := []Section{
		b.HeadMetaArgs,
		b.RequiredArgs,
		b.OptionalArgs,
		b.RequiredNestedBlocks,
		b.OptionalNestedBlocks,
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

func (b *NestedBlock) checkGap() bool {
	headMetaRange := mergeRange(b.HeadMetaArgs)
	argRange := mergeRange(b.RequiredArgs, b.OptionalArgs)
	nbRange := mergeRange(b.RequiredNestedBlocks, b.OptionalNestedBlocks)
	lastEndLine := -2
	for _, r := range []*hcl.Range{headMetaRange, argRange, nbRange} {
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

func (b *NestedBlock) isHeadMeta(argName string) bool {
	return b.BlockType() == "dynamic" && argName == "for_each"
}

func (b *NestedBlock) isTailMeta(argName string) bool {
	return false
}
