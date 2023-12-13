cd .. || exit
make build
cp dist/terraform-provider-schemaregistry ~/.terraform.d/plugins/local/fetch-rewards/confluent-schema-registry/1.1.0/darwin_arm64/terraform-provider-confluent-schema-registry_1.1.0
cd terraform-files-test || exit
rm -rf .terraform/providers
rm .terraform.lock.hcl
terraform init