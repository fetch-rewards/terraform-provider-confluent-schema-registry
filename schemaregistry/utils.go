package schemaregistry

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
)

const IDSeparator = "___"

func formatSchemaVersionID(subject string) string {
	return subject
}

func extractSchemaVersionID(id string) string {
	return id
}

// Create a temporary .proto file based on a proto schema string and return
// the path of the file which corresponds to the schema
func CreateTemporaryProtoFile(protoSchemaString string) (string, error) {
	var tempFileName string = ""
	customTempDir, err := os.MkdirTemp("", "proto-")
	if err != nil {
		return tempFileName, fmt.Errorf("error creating custom temp directory: %v", err)
	}

	// I dont think we need this since we moved this to utils instead ans separated into 2 funcs.
	// we can no long clean up the file at the end of the func, maybe at the end of the formatting one though
	// defer os.RemoveAll(customTempDir) // Clean up the directory afterwards

	// Create a temporary file
	tmpFile, err := os.CreateTemp(customTempDir, "*.proto")
	if err != nil {
		return tempFileName, fmt.Errorf("error creating temporary file: %v", err)
	}

	// Write the protobuf string to the file
	if _, err := tmpFile.WriteString(protoSchemaString); err != nil {
		return tempFileName, fmt.Errorf("error writing to temporary file: %v", err)
	}

	// Close the file to ensure all data is flushed
	if err := tmpFile.Close(); err != nil {
		return tempFileName, fmt.Errorf("error closing temporary file: %v", err)
	}

	tempFileName = tmpFile.Name()

	return tempFileName, err
}

// CompareASTs compares two AST nodes for equality
func CompareASTs(protoSchemaString1 string, protoSchemaString2 string) (bool, error) {
	filePath1, err := CreateTemporaryProtoFile(protoSchemaString1)

	if err != nil {
		return false, fmt.Errorf("failed to create temporary proto file: %w", err)
	}

	filePath2, err := CreateTemporaryProtoFile(protoSchemaString2)

	if err != nil {
		return false, fmt.Errorf("failed to create temporary proto file: %w", err)
	}

	// Parse the .proto file
	node1, err := parseProtoFile(filePath1)
	if err != nil {
		return false, fmt.Errorf("error parsing .proto file: %v", err)
	}

	// Parse the .proto file
	node2, err := parseProtoFile(filePath2)
	if err != nil {
		return false, fmt.Errorf("error parsing .proto file: %v", err)
	}

	// Compare everything but the Name field since Name is the file path
	newName := " "
	node1.FileDescriptorProto().Name = &newName
	node2.FileDescriptorProto().Name = &newName

	// Log INFO node1 and node 2 FileDescriptorProto().String to see differences
	log.Printf("[INFO] Proto File Descriptor 1: %s", node1.FileDescriptorProto().String())
	log.Printf("[INFO] Proto File Descriptor 2: %s", node2.FileDescriptorProto().String())

	// Use reflect.DeepEqual to compare the nodes
	return reflect.DeepEqual(node1.FileDescriptorProto(), node2.FileDescriptorProto()), nil
}

func parseProtoFile(filename string) (parser.Result, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	errHandler := reporter.NewHandler(nil)
	res, err := parser.Parse(filename, f, errHandler)
	if err != nil {
		return nil, err
	}
	return parser.ResultFromAST(res, true, errHandler)
}
