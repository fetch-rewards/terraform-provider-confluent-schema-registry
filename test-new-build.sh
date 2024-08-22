# make build
# cp dist/terraform-provider-schemaregistry ~/.terraform.d/plugins/local/fetch-rewards/confluent-schema-registry/1.1.0/linux_amd64/terraform-provider-confluent-schema-registry_1.1.0
make local_install
cd terraform-test-files || exit
rm -rf .terraform/providers
rm .terraform.lock.hcl
terraform init