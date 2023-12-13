terraform {
  required_providers {
    schemaregistry = {
      source = "local/fetch-rewards/confluent-schema-registry"
      version = "1.1.0"
    }

    kafka = {
      source = "mongey/kafka",
      version = "0.5.3"
    }
    aws = {
      source = "hashicorp/aws",
      version = "5.21.0"
    }
  }
}


provider "schemaregistry" {
  alias =  "schema-registry-dev"
  schema_registry_url =  "https://dev-event-tracking-schema-registry.fetchrewards.com"
}

provider "kafka" {
  alias = "event-tracking-dev"
  bootstrap_servers = [
    "dev-event-tracking-kafka.fetchrewards.com:9094"
  ]
  skip_tls_verify = true
  tls_enabled = false
}