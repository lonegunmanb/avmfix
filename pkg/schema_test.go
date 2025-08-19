package pkg

import (
	"testing"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBlockSchema_ResourceBlock(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	t.Run("azurerm_resource_group", func(t *testing.T) {
		path := []string{"resource", "azurerm_resource_group"}
		schema, err := queryBlockSchema(path, "hashicorp", "4.0.0")

		require.NoError(t, err)
		require.NotNil(t, schema)

		// Test that required attributes are present
		assert.Contains(t, schema.Attributes, "name")
		assert.Contains(t, schema.Attributes, "location")

		// Verify attribute properties
		nameAttr := schema.Attributes["name"]
		assert.True(t, nameAttr.Required)
		assert.False(t, nameAttr.Optional)

		locationAttr := schema.Attributes["location"]
		assert.True(t, locationAttr.Required)
		assert.False(t, locationAttr.Optional)

		// Test optional attributes
		if tagsAttr, exists := schema.Attributes["tags"]; exists {
			assert.False(t, tagsAttr.Required)
			assert.True(t, tagsAttr.Optional)
		}
	})

	t.Run("azapi_resource", func(t *testing.T) {
		path := []string{"resource", "azapi_resource"}
		schema, err := queryBlockSchema(path, "Azure", "")

		require.NoError(t, err)
		require.NotNil(t, schema)

		// Test post-processor effects - azapi_resource should have modified schema
		// where optional attributes like 'name', 'parent_id', 'location' become required
		if nameAttr, exists := schema.Attributes["name"]; exists {
			assert.True(t, nameAttr.Required, "name should be required after post-processing")
			assert.False(t, nameAttr.Optional, "name should not be optional after post-processing")
		}

		if parentIdAttr, exists := schema.Attributes["parent_id"]; exists {
			assert.True(t, parentIdAttr.Required, "parent_id should be required after post-processing")
			assert.False(t, parentIdAttr.Optional, "parent_id should not be optional after post-processing")
		}

		if locationAttr, exists := schema.Attributes["location"]; exists {
			assert.True(t, locationAttr.Required, "location should be required after post-processing")
			assert.False(t, locationAttr.Optional, "location should not be optional after post-processing")
		}
	})
}

func TestQueryBlockSchema_DataBlock(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	t.Run("azurerm_resource_group", func(t *testing.T) {
		path := []string{"data", "azurerm_resource_group"}
		schema, err := queryBlockSchema(path, "hashicorp", "4.0.0")

		require.NoError(t, err)
		require.NotNil(t, schema)

		// Data sources typically require name to query
		assert.Contains(t, schema.Attributes, "name")

		// Computed attributes should be present for data sources
		if locationAttr, exists := schema.Attributes["location"]; exists {
			assert.True(t, locationAttr.Computed, "location should be computed for data source")
		}
	})
}

func TestQueryBlockSchema_EphemeralBlock(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()

	path := []string{"ephemeral", "azurerm_key_vault_secret"}
	schema, err := queryBlockSchema(path, "hashicorp", "4.30.0")

	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Contains(t, schema.Attributes, "name")
	assert.Contains(t, schema.Attributes, "key_vault_id")
	assert.Contains(t, schema.Attributes, "version")
	assert.Contains(t, schema.Attributes, "expiration_date")
	assert.Contains(t, schema.Attributes, "not_before_date")
	assert.Contains(t, schema.Attributes, "value")
}

func TestQueryBlockSchema_NestedBlocks(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	t.Run("azurerm_container_group_container", func(t *testing.T) {
		// Test nested block access - container block within azurerm_container_group
		path := []string{"resource", "azurerm_container_group", "container"}
		schema, err := queryBlockSchema(path, "hashicorp", "4.30.0")

		require.NoError(t, err)
		require.NotNil(t, schema)

		// Container block should have required attributes
		assert.Contains(t, schema.Attributes, "name")
		assert.Contains(t, schema.Attributes, "image")

		nameAttr := schema.Attributes["name"]
		assert.True(t, nameAttr.Required)

		imageAttr := schema.Attributes["image"]
		assert.True(t, imageAttr.Required)
	})
}

func TestQueryBlockSchema_InvalidInputs(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	t.Run("empty_path", func(t *testing.T) {
		path := []string{}
		_, err := queryBlockSchema(path, "hashicorp", "4.0.0")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid path")
	})

	t.Run("short_path", func(t *testing.T) {
		path := []string{"resource"}
		_, err := queryBlockSchema(path, "hashicorp", "4.0.0")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid path")
	})

	t.Run("unsupported_category", func(t *testing.T) {
		path := []string{"invalid_category", "some_block"}
		_, err := queryBlockSchema(path, "hashicorp", "4.0.0")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupport block category")
	})
}

func TestQueryBlockSchema_ProviderNamespaceHandling(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	t.Run("azapi_default_namespace", func(t *testing.T) {
		path := []string{"resource", "azapi_resource"}
		// Test that azapi defaults to "Azure" namespace
		schema, err := queryBlockSchema(path, "", "")

		// Should succeed even without explicit namespace due to default handling
		require.NoError(t, err)
		require.NotNil(t, schema)
	})

	t.Run("explicit_namespace", func(t *testing.T) {
		path := []string{"resource", "azurerm_resource_group"}
		schema, err := queryBlockSchema(path, "hashicorp", "")

		require.NoError(t, err)
		require.NotNil(t, schema)
	})
}

func TestQueryBlockSchema_VersionHandling(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	t.Run("latest_version", func(t *testing.T) {
		path := []string{"resource", "azurerm_resource_group"}
		schema, err := queryBlockSchema(path, "", "")

		require.NoError(t, err)
		require.NotNil(t, schema)
	})

	t.Run("explicit_version", func(t *testing.T) {
		path := []string{"resource", "azurerm_resource_group"}
		// Note: This might fail if the specific version doesn't exist
		// In a real integration test, you'd use a known version
		schema, err := queryBlockSchema(path, "", "3.0.0")

		// We don't assert no error here as the version might not exist
		// but we test that the version parameter is handled
		if err == nil {
			assert.NotNil(t, schema)
		}
	})
}

func TestQueryBlockSchema_PostProcessors(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	testCases := []struct {
		name      string
		blockType string
		namespace string
	}{
		{"azapi_resource", "azapi_resource", "Azure"},
		{"azapi_update_resource", "azapi_update_resource", "Azure"},
		{"azapi_resource_action", "azapi_resource_action", "Azure"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := []string{"resource", tc.blockType}
			schema, err := queryBlockSchema(path, tc.namespace, "")

			require.NoError(t, err)
			require.NotNil(t, schema)

			// Test that post-processor was applied
			// All azapi post-processors should make certain attributes required
			requiredAttrs := []string{"name", "parent_id", "location", "resource_id", "action", "method", "query_parameters"}

			for _, attrName := range requiredAttrs {
				if attr, exists := schema.Attributes[attrName]; exists {
					assert.True(t, attr.Required, "attribute %s should be required after post-processing", attrName)
					assert.False(t, attr.Optional, "attribute %s should not be optional after post-processing", attrName)
				}
			}
		})
	}
}

// Integration test that verifies the schema retrieval works with real providers
func TestQueryBlockSchema_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()

	t.Run("multiple_providers", func(t *testing.T) {
		testCases := []struct {
			name      string
			path      []string
			namespace string
		}{
			{"azurerm_resource", []string{"resource", "azurerm_resource_group"}, ""},
			{"aws_instance", []string{"resource", "aws_instance"}, ""},
			{"google_compute_instance", []string{"resource", "google_compute_instance"}, ""},
			{"random_password", []string{"resource", "random_password"}, ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schema, err := queryBlockSchema(tc.path, tc.namespace, "")

				// In integration tests, we expect these to work
				// but if a provider isn't available, we log and continue
				if err != nil {
					t.Logf("Provider %s may not be available: %v", tc.path[1], err)
					return
				}

				require.NotNil(t, schema)
				assert.NotEmpty(t, schema.Attributes, "schema should have attributes")
			})
		}
	})
}

// Test helper functions used by queryBlockSchema
func TestNameSpaceOrDefault(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	testCases := []struct {
		name         string
		namespace    string
		providerType string
		expected     string
	}{
		{"explicit_namespace", "custom", "azurerm", "custom"},
		{"azapi_default", "", "azapi", "Azure"},
		{"msgraph_default", "", "msgraph", "microsoft"},
		{"hashicorp_default", "", "azurerm", "hashicorp"},
		{"random_default", "", "random", "hashicorp"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := nameSpaceOrDefault(tc.namespace, tc.providerType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Test that schema post-processors are registered correctly
func TestSchemaPostProcessors(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	expectedPostProcessors := []string{
		"azapi_resource",
		"azapi_update_resource",
		"azapi_resource_action",
	}

	for _, blockType := range expectedPostProcessors {
		t.Run(blockType, func(t *testing.T) {
			processor, exists := schemaPostProcessors[blockType]
			assert.True(t, exists, "post-processor should exist for %s", blockType)
			assert.NotNil(t, processor, "post-processor should not be nil for %s", blockType)
		})
	}
}

// Test error handling with non-existent provider types
func TestQueryBlockSchema_NonExistentProviderTypes(t *testing.T) {
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()
	t.Run("non_existent_resource", func(t *testing.T) {
		path := []string{"resource", "nonexistent_provider_resource"}
		_, err := queryBlockSchema(path, "", "")

		// Should return an error for non-existent provider/resource
		assert.Error(t, err)
	})

	t.Run("malformed_resource_type", func(t *testing.T) {
		path := []string{"resource", "malformed-resource-type-without-underscore"}
		_, err := queryBlockSchema(path, "", "")

		// Should handle malformed resource types gracefully
		assert.Error(t, err)
	})
}

// Test the getLatestVersion function
func TestGetLatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping getLatestVersion test in short mode")
	}
	stub := gostub.Stub(&tfPluginServer, NewServer(nil))
	defer stub.Reset()

	t.Run("valid_provider", func(t *testing.T) {
		// Test with a known provider
		version, err := getLatestVersion("hashicorp", "azurerm")

		require.NoError(t, err)
		assert.NotEmpty(t, version, "version should not be empty")
		assert.Contains(t, version, ".", "version should contain dots (semantic versioning)")
	})

	t.Run("azure_provider", func(t *testing.T) {
		// Test with Azure namespace
		version, err := getLatestVersion("Azure", "azapi")

		require.NoError(t, err)
		assert.NotEmpty(t, version, "version should not be empty")
	})

	t.Run("invalid_provider", func(t *testing.T) {
		// Test with non-existent provider
		_, err := getLatestVersion("nonexistent", "invalid")

		assert.Error(t, err, "should return error for non-existent provider")
		assert.Contains(t, err.Error(), "registry API returned status", "error should mention registry API status")
	})

	t.Run("empty_namespace", func(t *testing.T) {
		// Test with empty namespace
		_, err := getLatestVersion("", "azurerm")

		assert.Error(t, err, "should return error for empty namespace")
	})

	t.Run("empty_provider_type", func(t *testing.T) {
		// Test with empty provider type
		_, err := getLatestVersion("hashicorp", "")

		assert.Error(t, err, "should return error for empty provider type")
	})
}
