resource "kafka_topic" "test_provider_topic" {
  name               = "provider-test-topic"
  partitions         = 1
  replication_factor = 1
  config = {
    "min.insync.replicas": "1"
  }
  provider= kafka.event-tracking-dev
}

resource "schemaregistry_schema" "test_provider_topic_schema" {
  provider= schemaregistry.schema-registry-dev
  schema = "{\n\t\"type\": \"record\",\n\t\"name\": \"AwesomeEvent\",\n\t\"namespace\": \"com.fetchrewards.kia\",\n\t\"fields\": [\n\t\t{\n\t\t\t\"name\": \"id\",\n\t\t\t\"type\": \"int\",\n\t\t\t\"default\": \"null\"\n\t\t},\n\t\t{\n\t\t\t\"name\": \"eventType\",\n\t\t\t\"type\": \"string\"\n\t\t},\n\t\t{\n\t\t\t\"name\": \"updateTs\",\n\t\t\t\"type\": \"int\"\n\t\t}\n\t]\n}"
  schema_type = "AVRO"
  subject = "provider-test-schema"
}

resource "kafka_topic" "test_provider_topic_proto" {
  name               = "provider-test-topic-proto"
  partitions         = 1
  replication_factor = 1
  config = {
    "min.insync.replicas": "1"
  }
  provider= kafka.event-tracking-dev
}

resource "schemaregistry_schema" "test_provider_topic_schema_proto" {
  provider= schemaregistry.schema-registry-dev
  schema = "syntax = \"proto3\";package com.fetchrewards.locationservice.proto;\n\n\noption java_outer_classname = \"FidoLocationTrackerProto\";\n\nmessage FidoLocationTracker {\n  string location_id = 1;\n  string fido = 2;}"
  schema_type = "PROTOBUF"
  subject = "provider-test-schema-proto"
}
