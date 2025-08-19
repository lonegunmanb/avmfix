package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

type SchemaGetter interface {
	GetResourceSchema(request Request, resource string) (*tfjson.Schema, error)
	GetDataSourceSchema(request Request, dataSource string) (*tfjson.Schema, error)
	GetEphemeralResourceSchema(request Request, ephemeralResource string) (*tfjson.Schema, error)
}

var tfPluginServer SchemaGetter = NewServer(nil)

func queryBlockSchema(path []string, namespace string, version string) (*tfjson.SchemaBlock, error) {
	if len(path) < 2 {
		return nil, fmt.Errorf("invalid path:%v", path)
	}
	blockCategory := path[0]
	blockType := path[1]
	providerType := strings.Split(blockType, "_")[0]
	var getter func(Request, string) (*tfjson.Schema, error)
	switch blockCategory {
	case "resource":
		getter = tfPluginServer.GetResourceSchema
	case "data":
		getter = tfPluginServer.GetDataSourceSchema
	case "ephemeral":
		getter = tfPluginServer.GetEphemeralResourceSchema
	default:
		return nil, fmt.Errorf("unsupport block category: %s", blockCategory)
	}
	namespace = nameSpaceOrDefault(namespace, providerType)
	version, err := versionOrLatest(namespace, providerType, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get version for %s: %w", providerType, err)
	}
	schema, err := getter(Request{
		Namespace: namespace,
		Name:      providerType,
		Version:   version,
	}, blockType)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for %s: %w", blockType, err)
	}
	r := schema.Block
	if postProcessor, ok := schemaPostProcessors[blockType]; ok {
		postProcessor(r)
	}
	for i := 2; i < len(path); i++ {
		nb, ok := r.NestedBlocks[path[i]]
		if !ok {
			return nil, nil
		}
		r = nb.Block
	}
	return r, nil
}

func nameSpaceOrDefault(namespace string, providerType string) string {
	if namespace == "" {
		switch providerType {
		case "azapi":
			return "Azure"
		case "msgraph":
			return "microsoft"
		default:
			return "hashicorp"
		}
	}
	return namespace
}

func getLatestVersion(namespace string, providerType string) (string, error) {
	url := fmt.Sprintf("https://registry.terraform.io/v1/providers/%s/%s", namespace, providerType)

	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return "", fmt.Errorf("failed to fetch provider info from registry: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry API returned status %d for provider %s/%s", resp.StatusCode, namespace, providerType)
	}

	var providerInfo struct {
		Tag string `json:"tag"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&providerInfo); err != nil {
		return "", fmt.Errorf("failed to decode provider info response: %w", err)
	}

	if providerInfo.Tag == "" {
		return "", fmt.Errorf("no tag found in provider info for %s/%s", namespace, providerType)
	}

	return providerInfo.Tag, nil
}

func versionOrLatest(namespace, providerType, version string) (string, error) {
	if version == "" {
		v, err := getLatestVersion(namespace, providerType)
		if err != nil {
			return "", err
		}
		version = v
	}
	return strings.TrimPrefix(version, "v"), nil
}

var schemaPostProcessors = map[string]func(*tfjson.SchemaBlock){
	"azapi_resource":        azapiResourceSchemaPostProcessor,
	"azapi_update_resource": azapiResourceSchemaPostProcessor,
	"azapi_resource_action": azapiResourceSchemaPostProcessor,
}

func azapiResourceSchemaPostProcessor(b *tfjson.SchemaBlock) {
	// `name` and `parent_id` and `location` are optional, but obviously they're more important than body and other attributes.
	for _, key := range []string{
		"name",
		"parent_id",
		"location",
		"resource_id",
		"action",
		"method",
		"query_parameters",
	} {
		if attr, ok := b.Attributes[key]; ok {
			attr.Optional = false
			attr.Required = true
		}
	}
}
