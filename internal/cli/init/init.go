package cli_init

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var requiredCommands = []string{"git", "go"}
var isVerbose bool = false

var green = "\033[32m"
var red = "\033[31m"
var reset = "\033[0m"

func printInColour(color string, message string) {
	fmt.Printf("%s%s%s\n", color, message, reset)
}

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

	fmt.Printf("Usage: %s init [--frontend-url=<repo-url>] [--backend-url=<repo_url>] [--no-frontend] [--verbose|-v]\n", cmd)
	fmt.Println("\nOptions:")
	fmt.Println("  --backend-url=<repo_url>          The URL of the backend repository to clone (default: https://github.com/chukfi/cms.git)")
	fmt.Println("  --frontend-url=<repo-url>          The URL of the frontend repository to clone (default: https://github.com/chukfi/frontend.git)")
	fmt.Println("  --verbose, -v             Enable verbose output")
	fmt.Println("  --help, -h               Show this help message")

}

func CheckIfCommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return !os.IsNotExist(err)
}

// this is the main CLI function for generating types, do not call directly, use CLI by running the command
func CLI(args []string) {
	// check args
	frontend_url := "https://github.com/chukfi/frontend.git"
	backend_url := "https://github.com/chukfi/cms.git"
	use_frontend := true

	for _, arg := range args {
		if strings.HasPrefix(arg, "--frontend-url=") {
			frontend_url = strings.TrimPrefix(arg, "--frontend-url=")
		} else if strings.HasPrefix(arg, "--backend-url") {
			backend_url = strings.TrimPrefix(arg, "--backend-url=")
		} else if arg == "--no-frontend" {
			use_frontend = false
		} else if arg == "--verbose" || arg == "-v" {
			isVerbose = true
		}
	}

	// check if commands exist

	for _, cmd := range requiredCommands {
		if !CheckIfCommandExists(cmd) {
			fmt.Printf(red+"Error: Required command '%s' not found in PATH. Please install it before running this tool.\n"+reset, cmd)
			return
		}
	}

	printInColour(green, "Cloning backend repository:" + backend_url)

	cmd := exec.Command("git", "clone", backend_url)

	if isVerbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		printInColour(red, "Error cloning backend repository: " + err.Error())
		return
	}

}
