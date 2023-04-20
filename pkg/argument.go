package pkg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"sort"
)

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
type Args []*Arg

func (a Args) SortByName() Args {
	sortedArgs := make([]*Arg, len(a))
	copy(sortedArgs, a)
	sort.Slice(sortedArgs, func(i, j int) bool {
		return sortedArgs[i].Name < sortedArgs[j].Name
	})
	return sortedArgs
}

func buildAttrArg(sAttr *HclAttribute, file *hcl.File) *Arg {
	return &Arg{
		Name:  sAttr.Name,
		Range: sAttr.SrcRange,
		File:  file,
		Attr:  sAttr,
	}
}
