package pkg

import (
	"math"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Section is an interface offering general APIs of argument collections
type Section interface {
	// CheckOrder checks whether the arguments in the collection is sorted
	CheckOrder() bool

	// ToString prints arguments in the collection in order
	ToString() string

	// GetRange returns the entire range of the argument collection
	GetRange() *hcl.Range
}

func toString(sections ...Section) string {
	var lines []string
	for _, section := range sections {
		line := section.ToString()
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func mergeRange(sections ...Section) *hcl.Range {
	start := hcl.Pos{Line: math.MaxInt}
	end := hcl.Pos{Line: -1}
	filename := ""
	isNil := true
	for _, section := range sections {
		r := section.GetRange()
		if r == nil {
			continue
		}
		isNil = false
		if filename == "" {
			filename = r.Filename
		}
		if r.Start.Line < start.Line {
			start = r.Start
		}
		if r.End.Line > end.Line {
			end = r.End
		}
	}
	if isNil {
		return nil
	}
	return &hcl.Range{
		Filename: filename,
		Start:    start,
		End:      end,
	}
}

// Arg is a wrapper of the attribute
type Arg struct {
	Name  string
	Range hcl.Range
	File  *hcl.File
	Attr  *HclAttribute
}

// ToString prints the arg content
func (a *Arg) ToString() string {
	return string(hclwrite.Format(a.Range.SliceBytes(a.File.Bytes)))
}

// Args is the collection of args with the same type
type Args struct {
	Args  []*Arg
	Range *hcl.Range
}

// CheckOrder checks whether this type of args are sorted
func (a *Args) CheckOrder() bool {
	if a == nil {
		return true
	}
	var name *string
	for _, arg := range a.Args {
		if name != nil && *name > arg.Name {
			return false
		}
		name = &arg.Name
	}
	return true
}

// ToString prints this type of args in order
func (a *Args) ToString() string {
	if a == nil {
		return ""
	}
	sortedArgs := make([]*Arg, len(a.Args))
	copy(sortedArgs, a.Args)
	sort.Slice(sortedArgs, func(i, j int) bool {
		return sortedArgs[i].Name < sortedArgs[j].Name
	})
	var lines []string
	for _, arg := range sortedArgs {
		lines = append(lines, arg.ToString())
	}
	return string(hclwrite.Format([]byte(strings.Join(lines, "\n"))))
}

// GetRange returns the entire range of this type of args
func (a *Args) GetRange() *hcl.Range {
	if a == nil {
		return nil
	}
	return a.Range
}

// HeadMetaArgs is the collection of head meta args
type HeadMetaArgs struct {
	Args  []*Arg
	Range *hcl.Range
}

// CheckOrder checks whether the head meta args are sorted
func (a *HeadMetaArgs) CheckOrder() bool {
	if a == nil {
		return true
	}
	score := math.MaxInt
	for _, arg := range a.Args {
		if score < headMetaArgPriority[arg.Name] {
			return false
		}
		score = headMetaArgPriority[arg.Name]
	}
	return true
}

// ToString prints the head meta args in order
func (a *HeadMetaArgs) ToString() string {
	if a == nil {
		return ""
	}
	sortedArgs := make([]*Arg, len(a.Args))
	copy(sortedArgs, a.Args)
	sort.Slice(sortedArgs, func(i, j int) bool {
		return headMetaArgPriority[sortedArgs[i].Name] > headMetaArgPriority[sortedArgs[j].Name]
	})
	var lines []string
	for _, arg := range sortedArgs {
		lines = append(lines, arg.ToString())
	}
	return string(hclwrite.Format([]byte(strings.Join(lines, "\n"))))
}

// GetRange returns the entire range of head meta args
func (a *HeadMetaArgs) GetRange() *hcl.Range {
	if a == nil {
		return nil
	}
	return a.Range
}

func (a *Args) add(arg *Arg) {
	a.Args = append(a.Args, arg)
	a.updateRange(arg)
}

func (a *Args) updateRange(arg *Arg) {
	if a.Range == nil {
		a.Range = &hcl.Range{
			Filename: arg.Range.Filename,
			Start:    hcl.Pos{Line: math.MaxInt},
			End:      hcl.Pos{Line: -1},
		}
	}
	if a.Range.Start.Line > arg.Range.Start.Line {
		a.Range.Start = arg.Range.Start
	}
	if a.Range.End.Line < arg.Range.End.Line {
		a.Range.End = arg.Range.End
	}
}

func (a *HeadMetaArgs) add(arg *Arg) {
	a.Args = append(a.Args, arg)
	a.updateRange(arg)
}

func (a *HeadMetaArgs) updateRange(arg *Arg) {
	if a.Range == nil {
		a.Range = &hcl.Range{
			Filename: arg.Range.Filename,
			Start:    hcl.Pos{Line: math.MaxInt},
			End:      hcl.Pos{Line: -1},
		}
	}
	if a.Range.Start.Line > arg.Range.Start.Line {
		a.Range.Start = arg.Range.Start
	}
	if a.Range.End.Line < arg.Range.End.Line {
		a.Range.End = arg.Range.End
	}
}

func buildAttrArg(sAttr *HclAttribute, file *hcl.File) *Arg {
	return &Arg{
		Name:  sAttr.Name,
		Range: sAttr.SrcRange,
		File:  file,
		Attr:  sAttr,
	}
}
