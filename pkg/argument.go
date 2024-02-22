package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"sort"
)

var headMetaArgPriorities = map[string]string{
	"provider": "0",
	"count":    "1",
	"for_each": "2",
}

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

func (a Args) SortHeadMetaArgs() Args {
	sorted := Args{}
	linq.From(a).OrderBy(func(i interface{}) interface{} {
		priority := headMetaArgPriorities[i.(*Arg).Name]
		return priority
	}).ToSlice(&sorted)
	return sorted
}

func buildAttrArg(sAttr *HclAttribute, file *hcl.File) *Arg {
	return &Arg{
		HclAttribute: sAttr,
		Name:         sAttr.Name,
		File:         file,
	}
}
