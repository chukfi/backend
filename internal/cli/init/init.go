package cli_init

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	cli_frontend_downloader "github.com/chukfi/backend/internal/cli/frontend-downloader"
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

	fmt.Printf("Usage: %s init [--frontend-url=<repo-url>] [--backend-url=<repo_url>] [--directory=<directory>] [--no-frontend] [--verbose|-v]\n", cmd)
	fmt.Println("\nOptions:")
	fmt.Println("  --frontend-url=<repo-url>   The URL of the frontend repository to clone (default: https://github.com/chukfi/frontend.git)")
	fmt.Println("  --backend-url=<repo_url>    The URL of the backend repository to clone (default: https://github.com/chukfi/cms.git)")
	fmt.Println("  --directory=<directory>     The directory to clone the repositories into (default: ./backend)")
	fmt.Println("  --no-frontend               Do not clone the frontend repository")
	fmt.Println("  --verbose, -v               Enable verbose output")
	fmt.Println("  --help, -h                  Show this help message")

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
	directory := "./backend"

	for _, arg := range args {
		if strings.HasPrefix(arg, "--frontend-url=") {
			frontend_url = strings.TrimPrefix(arg, "--frontend-url=")
		} else if strings.HasPrefix(arg, "--backend-url") {
			backend_url = strings.TrimPrefix(arg, "--backend-url=")
		} else if arg == "--no-frontend" {
			use_frontend = false
		} else if strings.HasPrefix(arg, "--directory=") {
			directory = strings.TrimPrefix(arg, "--directory=")
		} else if arg == "--verbose" || arg == "-v" {
			isVerbose = true
		} else if arg == "--help" || arg == "-h" {
			printHelp()
			return
		}
	}

	// check if commands exist

	for _, cmd := range requiredCommands {
		if !CheckIfCommandExists(cmd) {
			fmt.Printf(red+"Error: Required command '%s' not found in PATH. Please install it before running this tool.\n"+reset, cmd)
			return
		}
	}

	// create temp directory
	tempDir, err := os.MkdirTemp("", "cms-init-")

	if err != nil {
		printInColour(red, "Error creating temporary directory: "+err.Error())
		return
	}

	printInColour(green, "Cloning backend repository: "+backend_url)

	cmd := exec.Command("git", "clone", backend_url, tempDir+"/backend")

	if isVerbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	pullErr := cmd.Run()

	if pullErr != nil {
		printInColour(red, "Error cloning backend repository: "+pullErr.Error())
		return
	}

	printInColour(green, "Cloned backend repository successfully.")

	// if the backend url is https://github.com/chukfi/cms.git, then use the folder inside called "backend-test"
	if backend_url == "https://github.com/chukfi/cms.git" {
		printlnVerbose("using crm directory, using ./backend-test")
		// rename backend-test to ../backend-real, then delete backend and rename backend-real to backend
		err = os.Rename(tempDir+"/backend/backend-test", tempDir+"/backend-real")
		if err != nil {
			printInColour(red, "Error renaming backend-test directory: "+err.Error())
			return
		}
		err = os.RemoveAll(tempDir + "/backend")
		if err != nil {
			printInColour(red, "Error removing old backend directory: "+err.Error())
			return
		}
		err = os.Rename(tempDir+"/backend-real", tempDir+"/backend")
		if err != nil {
			printInColour(red, "Error renaming backend-real to backend: "+err.Error())
			return
		}
	}
	printInColour(green, "Backend repository cloned successfully.")

	// move backend to ./backend
	err = os.Rename(tempDir+"/backend", directory)
	if err != nil {
		printInColour(red, "Error moving backend directory: "+err.Error())
		return
	}

	if use_frontend {
		fakeArgs := []string{"--url=" + frontend_url, "--directory=" + directory + "/public"}
		cli_frontend_downloader.CLI(fakeArgs)
		printlnVerbose("Done!")
	}

	printInColour(green, "Initialization completed successfully.")

}
