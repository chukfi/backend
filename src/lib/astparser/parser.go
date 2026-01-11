package astparser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type ParsedField struct {
	Name     string
	Type     string
	GormTag  string
	JSONTag  string
	Required bool
}

type ParsedStruct struct {
	Name      string
	TableName string
	Fields    []ParsedField
	AdminOnly bool
	Hidden    bool
}

func ParseSchemaFile(filePath string) ([]ParsedStruct, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("schema file not found: %s", absPath)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go file: %w", err)
	}

	var structs []ParsedStruct

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		parsedStruct := ParsedStruct{
			Name:      typeSpec.Name.Name,
			TableName: toSnakeCase(typeSpec.Name.Name) + "s",
			Fields:    []ParsedField{},
		}

		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 {
				if ident, ok := field.Type.(*ast.Ident); ok {
					if ident.Name == "BaseModel" {
						parsedStruct.Fields = append(parsedStruct.Fields, getBaseModelFields()...)
					}
				}
				if sel, ok := field.Type.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "BaseModel" {
						parsedStruct.Fields = append(parsedStruct.Fields, getBaseModelFields()...)
					}
				}
				continue
			}

			for _, name := range field.Names {
				if strings.ToLower(name.Name) == "adminonly" {
					parsedStruct.AdminOnly = true
					continue
				}
				if strings.ToLower(name.Name) == "hidden" {
					parsedStruct.Hidden = true
					continue
				}

				gormTag := ""
				jsonTag := ""
				if field.Tag != nil {
					tag := strings.Trim(field.Tag.Value, "`")
					gormTag = extractTag(tag, "gorm")
					jsonTag = extractTag(tag, "json")
				}

				if gormTag == "-" || gormTag == "-:all" {
					continue
				}

				fieldName := name.Name
				if jsonTag != "" && jsonTag != "-" {
					parts := strings.Split(jsonTag, ",")
					if parts[0] != "" {
						fieldName = parts[0]
					}
				}

				parsedField := ParsedField{
					Name:     fieldName,
					Type:     typeToString(field.Type),
					GormTag:  gormTag,
					JSONTag:  jsonTag,
					Required: strings.Contains(gormTag, "not null"),
				}

				parsedStruct.Fields = append(parsedStruct.Fields, parsedField)
			}
		}

		structs = append(structs, parsedStruct)
		return true
	})

	return structs, nil
}

func getBaseModelFields() []ParsedField {
	return []ParsedField{
		{Name: "ID", Type: "uuid.UUID", GormTag: "type:char(36);primaryKey", JSONTag: "", Required: true},
		{Name: "CreatedAt", Type: "time.Time", GormTag: "", JSONTag: "", Required: false},
		{Name: "UpdatedAt", Type: "time.Time", GormTag: "", JSONTag: "", Required: false},
		{Name: "DeletedAt", Type: "gorm.DeletedAt", GormTag: "index", JSONTag: "", Required: false},
	}
}

func extractTag(tag, key string) string {
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, key+":") {
			value := strings.TrimPrefix(part, key+":")
			value = strings.Trim(value, "\"")
			return value
		}
	}
	return ""
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	default:
		return "interface{}"
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func GoTypeToTypescript(goType string) string {
	switch {
	case strings.Contains(goType, "string"), strings.Contains(goType, "UUID"):
		return "string"
	case strings.Contains(goType, "int"), strings.Contains(goType, "uint"), strings.Contains(goType, "float"), strings.Contains(goType, "double"):
		return "number"
	case strings.Contains(goType, "bool"):
		return "boolean"
	case strings.Contains(goType, "Time"), strings.Contains(goType, "DeletedAt"):
		return "string"
	case strings.HasPrefix(goType, "[]"):
		elementType := strings.TrimPrefix(goType, "[]")
		return GoTypeToTypescript(elementType) + "[]"
	case strings.HasPrefix(goType, "map["):
		return "Record<string, any>"
	default:
		return "any"
	}
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

func GenerateTypescriptFromParsed(structs []ParsedStruct) string {
	var sb strings.Builder

	for i, s := range structs {
		if s.Hidden {
			continue
		}

		interfaceName := s.Name
		sb.WriteString("export interface " + interfaceName + " {\n")

		for _, field := range s.Fields {
			tsType := GoTypeToTypescript(field.Type)
			optional := ""
			if !field.Required && !strings.Contains(field.GormTag, "primaryKey") {
				optional = "?"
			}

			sb.WriteString("  " + field.Name + optional + ": " + tsType + ";\n")
		}

		sb.WriteString("}\n")

		if i < len(structs)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}