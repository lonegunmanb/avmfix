package pkg

import (
	"fmt"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	aws "github.com/lonegunmanb/terraform-aws-schema/v6/generated"
	azapi "github.com/lonegunmanb/terraform-azapi-schema/v2/generated"
	azurerm_v3 "github.com/lonegunmanb/terraform-azurerm-schema/v3/generated"
	azurerm "github.com/lonegunmanb/terraform-azurerm-schema/v4/generated"
	random "github.com/lonegunmanb/terraform-random-schema/v3/generated"
	"github.com/prashantv/gostub"
)

var _ SchemaGetter = dummySchemaGetter{}

type dummySchemaGetter struct{}

var preDefinedResourceSchemas = []map[string]*tfjson.Schema{
	azurerm_v3.Resources,
	aws.Resources,
	azapi.Resources,
	random.Resources,
}
var preDefinedDataSources = []map[string]*tfjson.Schema{
	azurerm.DataSources,
	aws.DataSources,
	azapi.DataSources,
	random.DataSources,
}

var preDefinedEphemeralResources = []map[string]*tfjson.Schema{
	azurerm.EphemeralResources,
	aws.EphemeralResources,
	azapi.EphemeralResources,
	random.EphemeralResources,
}

func returnSchema(schemas []map[string]*tfjson.Schema, key string) (*tfjson.Schema, error) {
	for _, schemaMap := range schemas {
		if schema, ok := schemaMap[key]; ok {
			return schema, nil
		}
	}
	return nil, fmt.Errorf("schema for %s not found", key)
}

func (d dummySchemaGetter) GetResourceSchema(request Request, resource string) (*tfjson.Schema, error) {
	return returnSchema(preDefinedResourceSchemas, resource)
}

func (d dummySchemaGetter) GetDataSourceSchema(request Request, dataSource string) (*tfjson.Schema, error) {
	return returnSchema(preDefinedDataSources, dataSource)
}

func (d dummySchemaGetter) GetEphemeralResourceSchema(request Request, ephemeralResource string) (*tfjson.Schema, error) {
	return returnSchema(preDefinedEphemeralResources, ephemeralResource)
}

func TestMain(m *testing.M) {
	stub := gostub.Stub(&resolveNamespace, func(string, *HclFile) (string, error) {
		return "registry.terraform.io/hashicorp", nil
	}).Stub(&resolveProviderVersion, func(string, string, *HclFile) (string, error) {
		return "4.37.0", nil
	}).Stub(&tfPluginServer, dummySchemaGetter{})
	defer stub.Reset()
	// Run the tests
	m.Run()

	// Any cleanup after tests can be done here if necessary.
}
