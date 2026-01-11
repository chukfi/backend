package main

import (
	"os"
	"strings"

	cli_frontend_downloader "github.com/chukfi/backend/internal/cli/frontend-downloader"
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
	case "setup-frontend":
		// git clones frontend repo (or a repo specified with --url=...)
		// and builds it with npm build
		// then moves the build files to ./public
		// then allow it to be served by the backend
		cli_frontend_downloader.CLI(otherArgs)

	case "init":
		// setups frontend and backend
		
	}
}
