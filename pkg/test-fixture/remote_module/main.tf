module "consul" {
  optional_variable = "value"
  required_variable = "value"
  source = "hashicorp/consul/aws"
  version = "0.1.0"
  for_each = var.for_each
  providers = {}
  depends_on = [null_resource.this]
}