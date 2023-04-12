package pkg

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/lonegunmanb/terraform-azurerm-schema/v3/generated"
)

func queryBlockSchema(path []string) *tfjson.SchemaBlock {
	if path[0] != "resource" && path[0] != "data" {
		return nil
	}
	if len(path) < 2 {
		panic(fmt.Sprintf("invalid path:%v", path))
	}
	root := generated.Resources
	if path[0] == "data" {
		root = generated.DataSources
	}

	b, ok := root[path[1]]
	if !ok {
		return nil
	}
	r := b.Block
	for i := 2; i < len(path); i++ {
		nb, ok := r.NestedBlocks[path[i]]
		if !ok {
			return nil
		}
		r = nb.Block
	}
	return r
}
