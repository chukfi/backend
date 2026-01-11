package generate_types

import (
	"fmt"
	"strings"

	"github.com/chukfi/backend/src/lib/astparser"
	"github.com/chukfi/backend/src/lib/schemaregistry"
	"gorm.io/gorm"
)

type GenerateTypesConfig struct {
	Schema   []interface{}
	Database *gorm.DB
}

func NewGenerateTypesConfig(schema []interface{}, database *gorm.DB) *GenerateTypesConfig {
	return &GenerateTypesConfig{
		Schema:   schema,
		Database: database,
	}
}

func GenerateTypescriptTypes(config *GenerateTypesConfig) []byte {
	types := schemaregistry.GenerateAllTypescriptInterfaces()

	var typeStrings []string
	for _, tsType := range types {
		typeStrings = append(typeStrings, tsType)
	}

	return []byte(strings.Join(typeStrings, "\n\n"))
}

func GenerateTypescriptFromSchemaFile(schemaPath string) (string, error) {
	structs, err := astparser.ParseSchemaFile(schemaPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse schema file: %w", err)
	}

	if len(structs) == 0 {
		return "", fmt.Errorf("no structs found in schema file")
	}

	fmt.Printf("Found %d struct(s) in schema file:\n", len(structs))
	for _, s := range structs {
		fmt.Printf("  - %s (%d fields)\n", s.Name, len(s.Fields))
	}

	typescriptCode := astparser.GenerateTypescriptFromParsed(structs) // :)

	return typescriptCode, nil
}
