package schemaregistry

import (
	"strings"
	"testing"
)

func TestProtoCompileASTComparison(t *testing.T) {
	originalProtoSchema := `syntax = "proto3";package com.fetchrewards.locationservice.proto;option java_outer_classname = "FidoLocationTrackerProto";
	
	
	message FidoLocationTracker {
	   string location_id = 1; 
	   
	 
	   string fido = 2;}
	
	`

	expectedProtoSchema := "syntax = \"proto3\";\n" +
		"package com.fetchrewards.locationservice.proto;\n\n" +
		"option java_outer_classname = \"FidoLocationTrackerProto\";\n\n" +
		"message FidoLocationTracker {\n" +
		"  string location_id = 1;\n" +
		"  string fido = 2;\n" +
		"}\n"

	schemaEquals, err := CompareASTs(originalProtoSchema, expectedProtoSchema)

	if err != nil {
		t.Error("Proto string formatter should not error")
	}

	if !schemaEquals {
		t.Errorf(
			"Original Proto Schema does not match expected Schema\nexpected:\n`%s`\n\nactual:\n`%s`",
			originalProtoSchema,
			expectedProtoSchema,
		)
	}
}

func TestProtoCompileASTComparisonFidoTracking(t *testing.T) {
	newSchema := `syntax = "proto3";package com.fetchrewards.locationservice.proto;


	option java_outer_classname = "FidoLocationTrackerProto";
	
	message FidoLocationTracker {
	  string location_id = 1;
	  string fido = 2;}
	
	`

	oldSchema := `syntax = "proto3";
	package com.fetchrewards.locationservice.proto;
	
	option java_outer_classname = "FidoLocationTrackerProto";
	
	message FidoLocationTracker {
	  string location_id = 1;
	  string fido = 2;
	}`

	schemaEquals, err := CompareASTs(newSchema, oldSchema)

	if err != nil {
		t.Error("Proto string formatter should not error")
	}

	if !schemaEquals {
		t.Errorf(
			"Original Proto Schema does not match expected Schema\nexpected:\n`%s`\n\nactual:\n`%s`",
			newSchema,
			oldSchema,
		)
	}
}

func TestMalformedProtoFileFormatting(t *testing.T) {
	// Expect a file that is not actually a proto schema to have a nil pointer dereference exception error
	originalProtoSchema := "Not a Protobuf schema"
	expectedProtoSchema := "Not a Protobuf schema"

	schemaEquals, err := CompareASTs(originalProtoSchema, expectedProtoSchema)

	if schemaEquals {
		t.Error("Expected schemaEquals to be false")
	}

	if !strings.Contains(err.Error(), "error parsing .proto file") {
		t.Errorf("expected error to contain 'failed to format proto file', but got: %v", err)
	}

}
