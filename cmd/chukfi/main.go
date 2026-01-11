package main

import (
	"fmt"
	"os"
	"strings"

	cli_frontend_downloader "github.com/chukfi/backend/internal/cli/frontend-downloader"
	cli_generate_types "github.com/chukfi/backend/internal/cli/generate-types"
	cli_init "github.com/chukfi/backend/internal/cli/init"
	"github.com/joho/godotenv"
)

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
	fmt.Printf("Usage: %s <command> [options]\n", cmd)
	fmt.Println("\nCommands:")
	fmt.Println("  generate-types       Generate Go types from the database schema")
	fmt.Println("  setup-frontend       Clone and build the frontend application")
	fmt.Println("  init                 Initialize the project by cloning frontend and backend repositories")
	fmt.Println("\nUse '<command> --help' for more information about a command.")
}

func main() {
	godotenv.Load()

	args := os.Args[1:]

	if len(args) == 0 {
		printHelp()
		os.Exit(0)
		return
	}

	command := args[0]
	otherArgs := args[1:]

	switch command {
	case "help", "--help", "-h":
		printHelp()
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
		cli_init.CLI(otherArgs)
	}
}
