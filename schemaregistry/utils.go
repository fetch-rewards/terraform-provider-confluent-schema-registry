package schemaregistry

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

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

// CompareASTs compares two AST nodes for equality
func CompareASTs(protoSchemaString1 string, protoSchemaString2 string) (bool, error) {

	// Parse the .proto file
	node1, err := protoStringToAST(protoSchemaString1)
	if err != nil {
		return false, fmt.Errorf("error parsing .proto file: %v", err)
	}

	// Parse the .proto file
	node2, err := protoStringToAST(protoSchemaString2)
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

func protoStringToAST(protoSchemaString string) (parser.Result, error) {
	errHandler := reporter.NewHandler(nil)
	var reader io.Reader = strings.NewReader(protoSchemaString)
	res, err := parser.Parse(" ", reader, errHandler)
	if err != nil {
		return nil, err
	}
	return parser.ResultFromAST(res, true, errHandler)
}
