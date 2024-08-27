package schemaregistry

import (
	"os"
	"strings"
	"testing"
)

func TestCreateTemporaryProtoFile(t *testing.T) {
	var protoSchemaString string = "syntax = \"proto3\";"

	protoFilePath, err := CreateTemporaryProtoFile(protoSchemaString)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\nFile Path: `%s`", protoFilePath)

	// Read from the tmp file and verify it matches the correct string
	// Read the entire file content
	content, err := os.ReadFile(protoFilePath)
	if err != nil {
		t.Fatal(err)
	}

	// Convert the byte slice to a string
	fileContent := string(content)

	t.Logf("\nFile Contents: `%s`", fileContent)

	if fileContent != protoSchemaString {
		t.Errorf("File content does not match original str")
	}

}

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
