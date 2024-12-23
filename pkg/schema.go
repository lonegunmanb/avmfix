package pkg

import (
	"fmt"

	"github.com/ahmetb/go-linq/v3"
	tfjson "github.com/hashicorp/terraform-json"
	alicloud "github.com/lonegunmanb/terraform-alicloud-schema/generated"
	aws "github.com/lonegunmanb/terraform-aws-schema/v5/generated"
	awscc "github.com/lonegunmanb/terraform-awscc-schema/generated"
	azapi "github.com/lonegunmanb/terraform-azapi-schema/generated"
	azuread "github.com/lonegunmanb/terraform-azuread-schema/v3/generated"
	azurerm "github.com/lonegunmanb/terraform-azurerm-schema/v4/generated"
	bytebase "github.com/lonegunmanb/terraform-bytebase-schema/generated"
	gcp "github.com/lonegunmanb/terraform-google-schema/v6/generated"
	helm "github.com/lonegunmanb/terraform-helm-schema/v2/generated"
	kubernetes "github.com/lonegunmanb/terraform-kubernetes-schema/v2/generated"
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
var ephemeralResourceSchemas = make(map[string]*tfjson.Schema, 0)
var blockTypesWithSchema = map[string]map[string]*tfjson.Schema{
	"resource":  resourceSchemas,
	"data":      dataSourceSchemas,
	"ephemeral": ephemeralResourceSchemas,
}

func init() {
	linq.From(azurerm.Resources).
		Concat(linq.From(azuread.Resources)).
		Concat(linq.From(alicloud.Resources)).
		Concat(linq.From(azapi.Resources)).
		Concat(linq.From(aws.Resources)).
		Concat(linq.From(awscc.Resources)).
		Concat(linq.From(bytebase.Resources)).
		Concat(linq.From(gcp.Resources)).
		Concat(linq.From(helm.Resources)).
		Concat(linq.From(kubernetes.Resources)).
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
		Concat(linq.From(alicloud.DataSources)).
		Concat(linq.From(azapi.DataSources)).
		Concat(linq.From(bytebase.DataSources)).
		Concat(linq.From(aws.DataSources)).
		Concat(linq.From(awscc.DataSources)).
		Concat(linq.From(gcp.DataSources)).
		Concat(linq.From(helm.DataSources)).
		Concat(linq.From(kubernetes.DataSources)).
		Concat(linq.From(null.DataSources)).
		Concat(linq.From(local.DataSources)).
		Concat(linq.From(template.DataSources)).
		Concat(linq.From(tls.DataSources)).
		Concat(linq.From(azapi.DataSources)).
		Concat(linq.From(time.DataSources)).
		Concat(linq.From(random.DataSources)).
		Concat(linq.From(modtm.DataSources)).ToMap(&dataSourceSchemas)
	linq.From(azurerm.EphemeralResources).
		Concat(linq.From(aws.EphemeralResources)).
		Concat(linq.From(gcp.EphemeralResources)).ToMap(&ephemeralResourceSchemas)
}

func queryBlockSchema(path []string) *tfjson.SchemaBlock {
	schemas, ok := blockTypesWithSchema[path[0]]
	if !ok {
		return nil
	}
	if len(path) < 2 {
		panic(fmt.Sprintf("invalid path:%v", path))
	}
	b, ok := schemas[path[1]]
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
