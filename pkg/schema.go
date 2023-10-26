package pkg

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	tfjson "github.com/hashicorp/terraform-json"
	azapi "github.com/lonegunmanb/terraform-azapi-schema/generated"
	azuread "github.com/lonegunmanb/terraform-azuread-schema/v2/generated"
	azurerm "github.com/lonegunmanb/terraform-azurerm-schema/v3/generated"
	local "github.com/lonegunmanb/terraform-local-schema/v2/generated"
	modtm "github.com/lonegunmanb/terraform-modtm-schema/generated"
	null "github.com/lonegunmanb/terraform-null-schema/v3/generated"
	random "github.com/lonegunmanb/terraform-random-schema/v3/generated"
	template "github.com/lonegunmanb/terraform-template-schema/v2/generated"
	time "github.com/lonegunmanb/terraform-time-schema/generated"
	tls "github.com/lonegunmanb/terraform-tls-schema/v4/generated"
)

var resourceSchemas = make(map[string]*tfjson.Schema, 0)
var dataSourceSchemas = make(map[string]*tfjson.Schema, 0)

func init() {
	linq.From(azurerm.Resources).
		Concat(linq.From(azuread.Resources)).
		Concat(linq.From(null.Resources)).
		Concat(linq.From(local.Resources)).
		Concat(linq.From(template.Resources)).
		Concat(linq.From(tls.Resources)).
		Concat(linq.From(azapi.Resources)).
		Concat(linq.From(time.Resources)).
		Concat(linq.From(random.Resources)).
		Concat(linq.From(modtm.Resources)).ToMap(&resourceSchemas)
	linq.From(azurerm.DataSources).
		Concat(linq.From(azuread.DataSources)).
		Concat(linq.From(null.DataSources)).
		Concat(linq.From(local.DataSources)).
		Concat(linq.From(template.DataSources)).
		Concat(linq.From(tls.DataSources)).
		Concat(linq.From(azapi.DataSources)).
		Concat(linq.From(time.DataSources)).
		Concat(linq.From(random.DataSources)).
		Concat(linq.From(modtm.DataSources)).ToMap(&dataSourceSchemas)
}

func queryBlockSchema(path []string) *tfjson.SchemaBlock {
	if path[0] != "resource" && path[0] != "data" {
		return nil
	}
	if len(path) < 2 {
		panic(fmt.Sprintf("invalid path:%v", path))
	}
	root := resourceSchemas
	if path[0] == "data" {
		root = dataSourceSchemas
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
