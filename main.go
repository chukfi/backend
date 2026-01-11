package main

import (
	"os"
	"strings"

	cli_generate_types "github.com/chukfi/backend/internal/cli/generate-types"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	args := os.Args[1:]

	if len(args) == 0 {
		println("No command provided")
		os.Exit(0)
		return
	}

	command := args[0]
	otherArgs := args[1:]
	println(strings.Join(args, " "))

	switch command {
	case "generate-types":
		dsn := os.Getenv("DATABASE_DSN")

		if dsn == "" {
			// check if theres a otherArgs containing --dsn=""
			for _, arg := range otherArgs {
				if strings.HasPrefix(arg, "--dsn=") {
					dsn = strings.TrimPrefix(arg, "--dsn=")
				}
			}
		}

		cli_generate_types.CLI(dsn, []interface{}{}, otherArgs)
	}
}
