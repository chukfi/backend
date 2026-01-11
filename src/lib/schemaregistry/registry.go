package schemaregistry

import (
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm/schema"
)

type FieldMetadata struct {
	Name       string
	Type       string
	GormTag    string
	JSONTag    string
	Required   bool
	PrimaryKey bool
}

type SchemaMetadata struct {
	TableName string
	AdminOnly bool
	Fields    []FieldMetadata
}

type simpleMetadata struct {
	AdminOnly bool
}

var (
	registry = make(map[string]SchemaMetadata)
	aliases  = make(map[string]string)
	mu       sync.RWMutex
)

func getTableName(model interface{}) string {
	if tabler, ok := model.(interface{ TableName() string }); ok {
		return tabler.TableName()
	}

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return schema.NamingStrategy{}.TableName(t.Name())
}

func extractFields(model interface{}) []FieldMetadata {
	var fields []FieldMetadata

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	extractFieldsRecursive(t, &fields)

	return fields
}

func extractFieldsRecursive(t reflect.Type, fields *[]FieldMetadata) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Anonymous {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct {
				extractFieldsRecursive(embeddedType, fields)
			}
			continue
		}

		if strings.ToLower(field.Name) == "adminonly" {
			continue
		}

		gormTag := field.Tag.Get("gorm")
		if gormTag == "-" || gormTag == "-:all" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		jsonName := field.Name
		if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				jsonName = parts[0]
			}
		}

		required := strings.Contains(gormTag, "not null")
		primaryKey := strings.Contains(gormTag, "primaryKey") || strings.Contains(gormTag, "primarykey")

		fieldMeta := FieldMetadata{
			Name:       jsonName,
			Type:       field.Type.String(),
			GormTag:    gormTag,
			JSONTag:    jsonTag,
			Required:   required,
			PrimaryKey: primaryKey,
		}

		*fields = append(*fields, fieldMeta)
	}
}

func hasHiddenField(model interface{}) bool {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.ToLower(field.Name) == "hidden" {
			return true
		}
	}

	return false
}

func hasAdminOnlyField(model interface{}) bool {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.ToLower(field.Name) == "adminonly" {
			return true
		}
	}

	return false
}

func singularize(name string) string {
	if len(name) == 0 {
		return name
	}

	if strings.HasSuffix(name, "ies") {
		return name[:len(name)-3] + "y"
	}
	if strings.HasSuffix(name, "ses") || strings.HasSuffix(name, "xes") || strings.HasSuffix(name, "zes") || strings.HasSuffix(name, "ches") || strings.HasSuffix(name, "shes") {
		return name[:len(name)-2]
	}
	if strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") {
		return name[:len(name)-1]
	}

	return name
}

func RegisterSchema(model interface{}) {
	tableName := getTableName(model)
	adminOnly := hasAdminOnlyField(model)
	hasHiddenField := hasHiddenField(model)
	fields := extractFields(model)

	// if hidden, do NOT register
	if hasHiddenField {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	registry[tableName] = SchemaMetadata{
		TableName: tableName,
		AdminOnly: adminOnly,
		Fields:    fields,
	}

	singular := singularize(tableName)
	if singular != tableName {
		aliases[singular] = tableName
	}
}

func RegisterSchemas(models []interface{}) {
	for _, model := range models {
		RegisterSchema(model)
	}
}

func IsAdminOnly(tableName string) bool {
	mu.RLock()
	defer mu.RUnlock()

	if meta, exists := registry[tableName]; exists {
		return meta.AdminOnly
	}

	return false
}

func GetMetadata(tableName string) (SchemaMetadata, bool) {
	mu.RLock()
	defer mu.RUnlock()

	meta, exists := registry[tableName]
	return meta, exists
}

func GetFields(tableName string) ([]FieldMetadata, bool) {
	mu.RLock()
	defer mu.RUnlock()

	if meta, exists := registry[tableName]; exists {
		return meta.Fields, true
	}

	return nil, false
}

func GetRequiredFields(tableName string) []string {
	mu.RLock()
	defer mu.RUnlock()

	var required []string
	if meta, exists := registry[tableName]; exists {
		for _, field := range meta.Fields {
			if field.Required && !field.PrimaryKey {
				required = append(required, field.Name)
			}
		}
	}

	return required
}

func GetFieldNames(tableName string) []string {
	mu.RLock()
	defer mu.RUnlock()

	var names []string
	if meta, exists := registry[tableName]; exists {
		for _, field := range meta.Fields {
			names = append(names, field.Name)
		}
	}

	return names
}

// ValidateBody checks for missing required fields and unknown fields in the provided body map.
// Uses a schema registry to validate the fields.
// e.g Post -> posts
func ValidateBody(tableName string, body map[string]interface{}) (missingFields []string, unknownFields []string) {
	mu.RLock()
	defer mu.RUnlock()

	meta, exists := registry[tableName]
	if !exists {
		return nil, nil
	}

	fieldMap := make(map[string]FieldMetadata)
	for _, field := range meta.Fields {
		fieldMap[field.Name] = field
	}

	for _, field := range meta.Fields {
		if field.Required && !field.PrimaryKey {
			if _, exists := body[field.Name]; !exists {
				missingFields = append(missingFields, field.Name)
			}
		}
	}

	for key := range body {
		if _, exists := fieldMap[key]; !exists {
			unknownFields = append(unknownFields, key)
		}
	}

	return missingFields, unknownFields
}

// ResolveTableName resolves the actual table name from the provided name,
// considering aliases and naming strategies.
func ResolveTableName(name string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	if _, exists := registry[name]; exists {
		return name, true
	}

	if actual, exists := aliases[name]; exists {
		return actual, true
	}

	snakeName := schema.NamingStrategy{}.TableName(name)
	if _, exists := registry[snakeName]; exists {
		return snakeName, true
	}

	if actual, exists := aliases[snakeName]; exists {
		return actual, true
	}

	return "", false
}

func GenerateAllTypescriptInterfaces() map[string]string {
	mu.RLock()
	defer mu.RUnlock()

	interfaces := make(map[string]string)
	for tableName := range registry {
		if tsInterface, ok := GenerateTypescriptInterface(tableName); ok {
			interfaces[tableName] = tsInterface
		}
	}

	return interfaces
}

func GenerateTypescriptInterface(tableName string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	meta, exists := registry[tableName]
	if !exists {
		return "", false
	}

	var sb strings.Builder
	sb.WriteString("export interface " + strings.Title(singularize(tableName)) + " {\n")
	for _, field := range meta.Fields {
		tsType := "any"
		switch {
		case strings.Contains(field.Type, "string"), strings.Contains(field.Type, "Text"), strings.Contains(field.Type, "UUID"):
			tsType = "string"
		case strings.Contains(field.Type, "int"), strings.Contains(field.Type, "uint"), strings.Contains(field.Type, "float"), strings.Contains(field.Type, "double"):
			tsType = "number"
		case strings.Contains(field.Type, "bool"):
			tsType = "boolean"
		case strings.Contains(field.Type, "Time"):
			tsType = "Date"
		}

		optional := ""
		if !field.Required && !field.PrimaryKey {
			optional = "?"
		}

		sb.WriteString("  " + field.Name + optional + ": " + tsType + ";\n")
	}
	sb.WriteString("}\n")

	return sb.String(), true
}

// Returns all registered schemas with info such as table name & admin only
func GetAllRegisteredSchemas() map[string]simpleMetadata {
	mu.RLock()
	defer mu.RUnlock()

	schemas := make(map[string]simpleMetadata)
	for tableName, meta := range registry {
		schemas[tableName] = simpleMetadata{
			AdminOnly: meta.AdminOnly,
		}
	}

	return schemas
}

func IsRegistered(name string) bool {
	_, exists := ResolveTableName(name)
	return exists
}
