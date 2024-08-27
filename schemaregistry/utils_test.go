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

func TestBufFormatting(t *testing.T) {
	originalProtoSchema := `syntax = "proto3";package com.fetchrewards.locationservice.proto;option java_outer_classname = "FidoLocationTrackerProto";
	
	
	message FidoLocationTracker {
	   string location_id = 1; 
	   
	 
	   string fido = 2;}
	
	`
	formattedProtoSchema, err := FormatProtobufString(originalProtoSchema)

	if err != nil {
		t.Error("Proto string formatter should not error")
	}

	expectedProtoSchema := "syntax = \"proto3\";\n" +
		"package com.fetchrewards.locationservice.proto;\n\n" +
		"option java_outer_classname = \"FidoLocationTrackerProto\";\n\n" +
		"message FidoLocationTracker {\n" +
		"  string location_id = 1;\n" +
		"  string fido = 2;\n" +
		"}\n"

	if formattedProtoSchema != expectedProtoSchema {
		t.Errorf(
			"Formatted Proto Schema does not match expected Schema\nexpected:\n`%s`\n\nactual:\n`%s`",
			expectedProtoSchema,
			formattedProtoSchema,
		)
	}
}

func TestMalfromedProtoFileFormatting(t *testing.T) {
	// Expect a file that is not actually a proto schema to have a nil pointer dereference exception error
	originalProtoSchema := "Not a Protobuf schema"
	formattedProtoSchema, err := FormatProtobufString(originalProtoSchema)

	if formattedProtoSchema != "" {
		t.Error("Expected formatted proto schema to be an empty string")
	}

	if !strings.Contains(err.Error(), "failed to format proto file") {
		t.Errorf("expected error to contain 'failed to format proto file', but got: %v", err)
	}

}
