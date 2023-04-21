package pkg

import (
	"github.com/hashicorp/hcl/v2"
	"sort"
)

// Arg is a wrapper of the attribute
type Arg struct {
	Name string
	File *hcl.File
	*HclAttribute
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
		HclAttribute: sAttr,
		Name:         sAttr.Name,
		File:         file,
	}
}
