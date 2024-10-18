make local_install
cd terraform-test-files || exit
rm -rf .terraform/providers
rm .terraform.lock.hcl
terraform init