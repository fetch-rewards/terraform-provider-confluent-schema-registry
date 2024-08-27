package schemaregistry

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bufbuild/buf/private/buf/bufformat"
	"github.com/bufbuild/buf/private/bufpkg/bufmodule"
	"github.com/bufbuild/buf/private/pkg/storage/storageos"
	"github.com/bufbuild/buf/private/pkg/tracing"
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

func FormatProtobufString(protoSchemaString string) (string, error) {
	var finalErr error = nil
	var fileContent string = ""

	re := regexp.MustCompile(`\s+`)
	protoSchemaString = re.ReplaceAllString(protoSchemaString, " ")

	// Need to add two new lines before message to take care of the case when the "option" line
	// Does not have two lines after it, otherwise formatting is off
	newProtoSchemaString := strings.ReplaceAll(protoSchemaString, "message", "\n\nmessage")

	fullFilepath, err := CreateTemporaryProtoFile(newProtoSchemaString)

	if err != nil {
		return fileContent, fmt.Errorf("failed to create temporary proto file: %w", err)
	}

	protoSchemaDir := filepath.Dir(fullFilepath)
	fileName := filepath.Base(fullFilepath) //Need to split on '/' and get last element

	// Not sure what context is doing here other than a mandatory parameter
	ctx := context.Background()

	bucket, err := storageos.NewProvider().NewReadWriteBucket(protoSchemaDir)

	if err != nil {
		return fileContent, fmt.Errorf("failed to get proto file bucket: %w", err)
	}

	moduleSetBuilder := bufmodule.NewModuleSetBuilder(ctx, tracing.NopTracer, bufmodule.NopModuleDataProvider, bufmodule.NopCommitProvider)

	moduleSetBuilder.AddLocalModule(bucket, protoSchemaDir, true)

	moduleSet, err := moduleSetBuilder.Build()

	if err != nil {
		return fileContent, fmt.Errorf("failed to build proto file formatter: %w", err)
	}

	// At this point all the files are formatted
	readBucket, err := bufformat.FormatModuleSet(ctx, moduleSet)

	if err != nil {
		return fileContent, fmt.Errorf("failed to format proto file: %w", err)
	}

	obj, err := readBucket.Get(ctx, fileName)
	if err != nil {
		return fileContent, fmt.Errorf("failed to get formatted proto file '%s': %w", fullFilepath, err)
	}
	defer obj.Close()

	// Read the file content
	content, err := io.ReadAll(obj)
	if err != nil {
		return fileContent, fmt.Errorf("failed to read formatted proto file contents: %w", err)
	}

	fileContent = string(content)

	return fileContent, finalErr
}
