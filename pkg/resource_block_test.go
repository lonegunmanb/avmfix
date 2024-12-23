package pkg_test

import (
	"testing"

	"github.com/lonegunmanb/avmfix/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildResourceGroup_ArgumentsOnly(t *testing.T) {
	code := `
resource "azurerm_resource_group" "example" {
  name     = "example"
  location = "West Europe"
  tags     = {
	environment = "Production"
  }
}`
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	assert.Equal(t, "example", resourceBlock.Name)
	assert.Equal(t, "azurerm_resource_group", resourceBlock.Type)
	assert.Equal(t, 2, len(resourceBlock.RequiredArgs))
	assert.Equal(t, "name", resourceBlock.RequiredArgs[0].Name)
	assert.Equal(t, "location", resourceBlock.RequiredArgs[1].Name)
	assert.Equal(t, 1, len(resourceBlock.OptionalArgs))
	assert.Equal(t, "tags", resourceBlock.OptionalArgs[0].Name)
}

func TestBuildResourceGroup_MetaArguments(t *testing.T) {
	code := `
resource "azurerm_resource_group" "example" {
  count    = var.create_resource_group ? 1 : 0
  provider = azurerm.example

  name     = "example"
  location = "West Europe"
  tags     = {
	environment = "Production"
  }

  depends_on = [var.depends_on]

  lifecycle {
    create_before_destroy = false
	prevent_destroy 	  = false
	ignore_changes 		  = [
		tags,
	]
	replace_triggered_by = [
		"null_resource.trigger",
	]
  }
}`
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	assert.Equal(t, 2, len(resourceBlock.HeadMetaArgs))
	assert.Equal(t, "count", resourceBlock.HeadMetaArgs[0].Name)
	assert.Equal(t, "provider", resourceBlock.HeadMetaArgs[1].Name)
	assert.Equal(t, "depends_on", resourceBlock.TailMetaArgs[0].Name)
	assert.Equal(t, 1, len(resourceBlock.TailMetaNestedBlocks.Blocks))
	lifecycleBlock := resourceBlock.TailMetaNestedBlocks.Blocks[0]
	assert.Equal(t, "lifecycle", lifecycleBlock.Name)
	assert.Nil(t, lifecycleBlock.RequiredArgs)
	assert.Equal(t, 4, len(lifecycleBlock.OptionalArgs))
	assert.Equal(t, "create_before_destroy", lifecycleBlock.OptionalArgs[0].Name)
	assert.Equal(t, "prevent_destroy", lifecycleBlock.OptionalArgs[1].Name)
	assert.Equal(t, "ignore_changes", lifecycleBlock.OptionalArgs[2].Name)
	assert.Equal(t, "replace_triggered_by", lifecycleBlock.OptionalArgs[3].Name)
}

func TestResourceBlockAutoFix(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  name                = "example-continst"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  ip_address_type     = "Public"
  dns_name_label      = "aci-label"
  os_type             = "Linux"

  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }

  tags = {
    environment = "testing"
  }
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location
  name                = "example-continst"
  os_type             = "Linux"
  resource_group_name = azurerm_resource_group.example.name
  dns_name_label      = "aci-label"
  ip_address_type     = "Public"
  tags = {
    environment = "testing"
  }

  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlockAutoFix_HeadMetaArgs(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  provider            = azurerm.east
  count               = 1
  name                = "example-continst"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  ip_address_type     = "Public"
  dns_name_label      = "aci-label"
  os_type             = "Linux"

  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }

  tags = {
    environment = "testing"
  }
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_container_group" "example" {
  provider            = azurerm.east
  count               = 1

  location            = azurerm_resource_group.example.location
  name                = "example-continst"
  os_type             = "Linux"
  resource_group_name = azurerm_resource_group.example.name
  dns_name_label      = "aci-label"
  ip_address_type     = "Public"
  tags = {
    environment = "testing"
  }

  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_TailMetaNestedBlockShouldBePutAtResourceBlockTail(t *testing.T) {
	code := `
resource "azurerm_kubernetes_cluster" "example" {
  lifecycle {
    prevent_destroy = true
  }
  service_mesh_profile {
    mode = "Istio"
  }
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_kubernetes_cluster" "example" {
  service_mesh_profile {
    mode = "Istio"
  }

  lifecycle {
    prevent_destroy = true
  }
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_TailMetaArgOrder(t *testing.T) {
	code := `
resource "azurerm_kubernetes_cluster" "example" {
  lifecycle {
    prevent_destroy = true
  }

  depends_on = [azurerm_resource_group.this]
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_kubernetes_cluster" "example" {
  depends_on = [azurerm_resource_group.this]
  
  lifecycle {
    prevent_destroy = true
  }
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_AutoFix_DynamicNestedBlock(t *testing.T) {
	code := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location

  dynamic "dns_config" {
	for_each = var.nameservers == null ? ["dns_config"] : []

	content {
    	nameservers = var.nameservers
	}
  }
  container {
	cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"
	memory_limit = 1.5

	dynamic "gpu_limit" {
      for_each = var.gpu_limit_enabled ? ["1"] : []

		content {
			sku = "K80"
			count = 1
		}
	}
  }
}
`
	file, diag := pkg.ParseConfig([]byte(code), "")
	require.False(t, diag.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_container_group" "example" {
  location            = azurerm_resource_group.example.location

  container {
	cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"
	memory_limit = 1.5

	dynamic "gpu_limit" {
      for_each = var.gpu_limit_enabled ? ["1"] : []

		content {
			count = 1
			sku = "K80"
		}
	}
  }
  dynamic "dns_config" {
	for_each = var.nameservers == null ? ["dns_config"] : []

	content {
    	nameservers = var.nameservers
	}
  }
}
`
	s := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(s))
}

func TestResourceBlockAutoFix_CommentsShouldBePreserved(t *testing.T) {
	code := `
# Multi-Line Comments
#   Here
resource "azurerm_container_group" "example" {
  name                = "example-continst"
  os_type             = "Linux"
  resource_group_name = azurerm_resource_group.example.name
  ip_address_type     = "Public"
  # Nested Block Comment
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
	# Multi-Line
	#   Nested Block Comment
    ports {
      port     = 443
      protocol = "TCP"
    }
    name   = "hello-world"
  }

  # Argument comment dns_name_label
  dns_name_label      = "aci-label"
  #Argument Comment location
  location            = azurerm_resource_group.example.location
  
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }
  # tags
  tags = {
    environment = "testing"
  }
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
# Multi-Line Comments
#   Here
resource "azurerm_container_group" "example" {
  #Argument Comment location
  location            = azurerm_resource_group.example.location
  name                = "example-continst"
  os_type             = "Linux"
  resource_group_name = azurerm_resource_group.example.name
  # Argument comment dns_name_label
  dns_name_label      = "aci-label"
  ip_address_type     = "Public"
  # tags
  tags = {
    environment = "testing"
  }
  
  # Nested Block Comment
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    memory = "1.5"
    name   = "hello-world"

	# Multi-Line
	#   Nested Block Comment
    ports {
      port     = 443
      protocol = "TCP"
    }
  }
  container {
    cpu    = "0.5"
    image  = "mcr.microsoft.com/azuredocs/aci-tutorial-sidecar"
    memory = "1.5"
    name   = "sidecar"
  }
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlockAutoFix_Datasource(t *testing.T) {
	code := `
data "azurerm_virtual_network" "example" {
  resource_group_name = "networking"
  name                = "production"
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
data "azurerm_virtual_network" "example" {
  name                = "production"
  resource_group_name = "networking"
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlockAutoFix_DependsOn(t *testing.T) {
	code := `
resource "azurerm_resource_group" "example" {
  depends_on = [var.depends_on]
  name     = "example"
  location = "West Europe"
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_resource_group" "example" {
  location = "West Europe"
  name     = "example"

  depends_on = [var.depends_on]
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlockAutoFix_Lifecycle(t *testing.T) {
	code := `
resource "azurerm_resource_group" "example" {
  lifecycle {
    precondition {
      condition     = startswith(var.resource_group_name, "dev_")
      error_message = "Resource Group name must starts with dev_"  
    }
    prevent_destroy = true
  }
  depends_on = [var.depends_on]
  name     = var.resource_group_name
  location = "West Europe"
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `
resource "azurerm_resource_group" "example" {
  location = "West Europe"
  name     = var.resource_group_name
  
  depends_on = [var.depends_on]
  
  lifecycle {
    prevent_destroy = true
    
    precondition {
      condition     = startswith(var.resource_group_name, "dev_")
      error_message = "Resource Group name must starts with dev_"  
    }
  }
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlockAutoFix_SingleLineLifecycle(t *testing.T) {
	code := `resource "azurerm_key_vault" "kv" {
  lifecycle { ignore_changes = [tags] }
  name                     = local.keyvault_name
}
`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `resource "azurerm_key_vault" "kv" {
  name                     = local.keyvault_name

  lifecycle { 
	ignore_changes = [tags] 
  }
}
`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_SingleLineResource(t *testing.T) {
	code := `resource "random_pet" "test" { prefix = "abc" }`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `resource "random_pet" "test" {
  prefix = "abc" 
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_EmptyBlock(t *testing.T) {
	code := `resource "random_pet" "test" {}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `resource "random_pet" "test" {}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_ProviderShouldBeTheFirstMetaArgument(t *testing.T) {
	code := `resource "azurerm_resource_group" "test" {
count = 1
provider = azurerm.test

name = "test"
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `resource "azurerm_resource_group" "test" {
provider = azurerm.test
count = 1

name = "test"
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_UnknownNestedBlockShouldBeTreatedAsOptionalBlock(t *testing.T) {
	code := `resource "azurerm_kubernetes_cluster" "test" {
    an_unknown_block {
      unknown_argument = 1
    }
	default_node_pool {
      name       = "default"
      vm_size    = "Standard_D2_v2"
      node_count = 1
    }
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	expected := `resource "azurerm_kubernetes_cluster" "test" {
	default_node_pool {
      name       = "default"
      vm_size    = "Standard_D2_v2"
      node_count = 1
    }
    an_unknown_block {
      unknown_argument = 1
    }
}`
	fixed := string(file.WriteFile.Bytes())
	assert.Equal(t, formatHcl(expected), formatHcl(fixed))
}

func TestResourceBlock_WellFormattedDatasource(t *testing.T) {
	code := `# Enabling vm extensions - Log Analytics for arc and vulnerability assessment
data "azurerm_policy_definition" "vm_policies" {
  for_each = contains(var.mdc_plans_list, "VirtualMachines") ? local.virtual_machine_policies : {}

  display_name = each.value.definition_display_name
}`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	assert.Equal(t, formatHcl(code), formatHcl(string(file.WriteFile.Bytes())))
}

func TestResourceBlock_DatasourceShouldBeFixed(t *testing.T) {
	code := `
data "azurerm_resources" "spokes" {
  type = "Microsoft.Network/virtualNetworks"
  resource_group_name = "spokes_rg"
}
`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	want := `
data "azurerm_resources" "spokes" {
  resource_group_name = "spokes_rg"
  type = "Microsoft.Network/virtualNetworks"
}
`
	assert.Equal(t, formatHcl(want), formatHcl(string(file.WriteFile.Bytes())))
}

func TestEphemeralResource(t *testing.T) {
	code := `
ephemeral "aws_kms_secrets" "example" {
  secret {
    key_id               = "ab123456-c012-4567-890a-deadbeef123"
    encryption_algorithm = "RSAES_OAEP_SHA_256"
    payload = "AQECAHgaPa0J8WadplGCqqVAr4HNvDaFSQ+NaiwIBhmm6qDSFwAAAGIwYAYJKoZIhvcNAQcGoFMwUQIBADBMBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDI+LoLdvYv8l41OhAAIBEIAfx49FFJCLeYrkfMfAw6XlnxP23MmDBdqP8dPp28OoAQ=="
    name    = "app_specific_secret"
  }
}
`
	file, diagnostics := pkg.ParseConfig([]byte(code), "")
	require.False(t, diagnostics.HasErrors())
	resourceBlock := pkg.BuildResourceBlock(file.GetBlock(0), file.File)
	resourceBlock.AutoFix()
	want := `
ephemeral "aws_kms_secrets" "example" {
  secret {
    name    = "app_specific_secret"
    payload = "AQECAHgaPa0J8WadplGCqqVAr4HNvDaFSQ+NaiwIBhmm6qDSFwAAAGIwYAYJKoZIhvcNAQcGoFMwUQIBADBMBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDI+LoLdvYv8l41OhAAIBEIAfx49FFJCLeYrkfMfAw6XlnxP23MmDBdqP8dPp28OoAQ=="
    encryption_algorithm = "RSAES_OAEP_SHA_256"
    key_id               = "ab123456-c012-4567-890a-deadbeef123"
  }
}
`
	assert.Equal(t, formatHcl(want), formatHcl(string(file.WriteFile.Bytes())))
}
