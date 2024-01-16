# Azure Verified Module Autofix Tool

![](https://img.shields.io/github/actions/workflow/status/lonegunmanb/azure-verified-module-fix/pr_check.yaml?label=Build&style=for-the-badge)

[Azure Verified Modules](https://github.com/Azure/terraform-azure-modules) are a set of well maintained, consistent and trusted Terraform modules that maintained by Microsoft.

The Azure Verified Module Autofix Tool is a utility that can help you ensure your Terraform modules are in compliance with the [Azure Verified Modules Codex](https://github.com/Azure/terraform-azure-modules/blob/main/codex/README.md). By analyzing your code, the tool can identify some issues and automatically fix them to meet the required standards.

However, it's important to note that manual intervention may be required to fix some issues, as not all can be automatically resolved, but you can follow the guidelines provided in the [Azure Verified Modules Codex](https://github.com/Azure/terraform-azure-modules/blob/main/codex/README.md) to ensure your Terraform modules are compliant. This includes following the recommended directory structure, naming conventions, and documentation standards. Regularly reviewing and updating your modules according to these guidelines will help you maintain high-quality Terraform modules.

For now, the autofix tool can fix the following issues:

* [Orders Within resource and data Blocks](https://github.com/Azure/terraform-azure-modules/blob/main/codex/logic_code/resource.md#orders-within-resource-and-data-blocks)
* [Order to define variable](https://github.com/Azure/terraform-azure-modules/blob/main/codex/logic_code/variables.tf.md#order-to-define-variable)
* [Do not declare `nullable = true` for `variable`](https://github.com/Azure/terraform-azure-modules/blob/main/codex/logic_code/variables.tf.md#do-not-declare-nullable--true)
* Do not declare `sensitive = false` for `variable`
* [`output` should be arranged alphabetically](https://github.com/Azure/terraform-azure-modules/blob/main/codex/logic_code/outputs.md#output-should-be-arranged-alphabetically)
* Do not declare `sensitive = false` for `output`
* [`local` should be arranged alphabetically](https://github.com/Azure/terraform-azure-modules/blob/main/codex/logic_code/locals.tf.md#local-should-be-arranged-alphabetically)

We're adding more autofix capabilities to the tool, so stay tuned for updates!

## Installation

```bash
go install github.com/lonegunmanb/avmfix@latest
```

# How to use

To use `avmfix`, open a shell or terminal and run the following command:

```shell
avmfix -folder /path/to/your/terraform/module
```

Replace `/path/to/your/terraform/module` with the path to the directory containing your Terraform module.

The tool will analyze the specified directory and automatically apply fixes for any issues it identifies, according to the Azure Verified Modules Codex. If the process completes successfully, you will see the message "DirectoryAutoFix completed successfully." If an error occurs during the process, the tool will display an error message.

Keep in mind that `avmfix` may not be able to resolve all issues automatically. Manual intervention may be required for some problems. Regularly review and update your Terraform modules according to the Azure Verified Modules Codex to maintain high-quality modules.

# Supported Providers

`avmfix` currently supports variable block description generation for the following providers:
* Alicloud (`alicloud`)
* AWS (`aws`)
* AWS Cloud Control API (`awscc`)
* AzAPI (`azapi`)
* Azure Resource Manager (`azurerm`)
* Azure Active Directory (`azuread`)
* Google Cloud Platform (`google`)
* Helm (`helm`)
* Kubernetes (`kubernetes`)
* Local (`local`)
* Modtm (`modtm`)
* Null (`null`)
* Random (`random`)
* Template (`template`)
* Time (`time`)
* Tls (`tls`)