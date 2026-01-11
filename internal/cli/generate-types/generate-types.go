package cli_generate_types

import (
	"fmt"
	"os"
	"strings"

	generate_types "github.com/chukfi/backend/cmd/generate-types"
	"github.com/chukfi/backend/src/lib/detection"

	mysql "github.com/chukfi/backend/database/mysql"
)

var isVerbose bool = false

func printlnVerbose(message string) {
	if isVerbose {
		fmt.Println(message)
	}
}

func printHelp() {
	// determine how the command is running (e.g go run main.go vs compiled binary)
	cmd := os.Args[0]
	// if it ends with .exe (windows), remove the preceding path
	if strings.HasSuffix(cmd, ".exe") {
		parts := strings.Split(cmd, string(os.PathSeparator))
		cmd = parts[len(parts)-1]
	} else if strings.Contains(cmd, "go-build") {
		cmd = "go run main.go"
	}

	// for linux/mac, if it contains /, remove preceding path
	if strings.Contains(cmd, "/") {
		parts := strings.Split(cmd, string(os.PathSeparator))
		cmd = parts[len(parts)-1]
	}
	fmt.Printf(`
Usage: %s generate-types [options]

Description:
The generate-types command generates TypeScript types from Go schema definitions.
The database needs to be accessible & running to fetch the schema metadata.

Options:
  --schema=<path>*    Path to a Go file containing your schema structs
                      (e.g., --schema=./schema.go)
					 

  --output=<path>    Output path for generated TypeScript file
                     (default: cms.types.ts)

  --database=<type>  Database type (mysql/postgres)
                     Only needed when the tool cannot auto-detect the database type.

  --dsn=<dsn>        Database DSN connection string
                     (e.g., --dsn="user:password@tcp(127.0.0.1:3306)/dbname")
                     Not needed if you have DATABASE_DSN set in your environment variables.

  * = required

Examples:
   %s generate-types --schema=../backend-test/schema.go
   %s generate-types --schema=./myschema.go --output=./types/api.ts
   %s generate-types --schema=./myschema.go --dsn="user:password@tcp(127.0.0.1:3306)/dbname" --output=./types/api.ts --database=mysql
`, cmd, cmd, cmd, cmd)
}

// this is the main CLI function for generating types, do not call directly, use CLI by running the command
func CLI(dsn string, customSchema []interface{}, args []string) {
	var schemaPath string
	var outputPath string
	var showHelp bool
	var databaseProvider detection.DatabaseType = detection.Unknown

	for _, arg := range args {
		if strings.HasPrefix(arg, "--schema=") {
			schemaPath = strings.TrimPrefix(arg, "--schema=")
		}
		if strings.HasPrefix(arg, "--output=") {
			outputPath = strings.TrimPrefix(arg, "--output=")
		}
		if arg == "--help" || arg == "-h" {
			showHelp = true
		}
		if arg == "--verbose" || arg == "-v" {
			isVerbose = true
			printlnVerbose("is verbose!")
		}
		if strings.HasPrefix(arg, "--database=") {
			databaseArg := strings.TrimPrefix(arg, "--database=")
			switch databaseArg {
			case "mysql":
				databaseProvider = detection.MySQL
			case "postgres":
				databaseProvider = detection.PostgreSQL
			default:
				databaseProvider = detection.Unknown
			}
		}
	}

	if showHelp {
		printHelp()
		return
	}

	if dsn == "" {
		fmt.Println("No DATABASE_DSN not set.")
		printHelp()
		os.Exit(1)
	}

	if databaseProvider == detection.Unknown {
		printlnVerbose("no database, detecting")
		databaseProvider = detection.DetectDatabaseType(dsn)
	}

	if schemaPath == "" {
		fmt.Println("No schema file provided, set one using --schema=<path> to generate types from a Go schema file.")
		return
	}

	if databaseProvider == detection.Unknown {
		panic("Failed to detect the database type, please retry the command with --database=mysql/postgres/etc.")
	}

	switch databaseProvider {
	case detection.MySQL:
		mysql.InitDatabase(customSchema)

		GenerateTypesConfig := generate_types.NewGenerateTypesConfig(customSchema, mysql.DB)
		bytes := generate_types.GenerateTypescriptTypes(GenerateTypesConfig)

		typescriptCode, err := generate_types.GenerateTypescriptFromSchemaFile(schemaPath)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// append bytes to typescriptCode with a newline in between
		typescriptCode += "\n\n" + string(bytes)

		if outputPath == "" {
			outputPath = "./cms.types.ts"
		}

		err = os.WriteFile(outputPath, []byte(typescriptCode), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing TypeScript file: %v\n", err)
			os.Exit(1)
		}
		println("Done! Types have been generated to ./cms.types.ts")
	default:
		panic("Database type not supported yet for type generation")
	}
}
