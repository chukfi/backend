package cli_frontend_downloader

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var requiredCommands = []string{"git", "npm"}
var isVerbose bool = false

var green = "\033[32m"
var red = "\033[31m"
var reset = "\033[0m"

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

	fmt.Printf("Usage: %s frontend [--url=<repo-url>] [--directory=<output-directory>] [--verbose|-v]\n", cmd)
	fmt.Println("\nOptions:")
	fmt.Println("  --url=<repo-url>          The URL of the frontend repository to clone (default: https://github.com/chukfi/frontend.git)")
	fmt.Println("  --directory=<output-directory>  The directory to move the built frontend files to (default: ./public)")
	fmt.Println("  --verbose, -v             Enable verbose output")
	fmt.Println("  --help, -h               Show this help message")

}

func CheckIfCommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return !os.IsNotExist(err)
}

// this is the main CLI function for generating types, do not call directly, use CLI by running the command
func CLI(args []string) {
	var url string
	var directory string

	for _, arg := range args {
		if strings.HasPrefix(arg, "--url="){
			url = strings.TrimPrefix(arg, "--url=")
		} else if strings.HasPrefix(arg, "--directory=") {
			directory = strings.TrimPrefix(arg, "--directory=")
		} else if arg == "--verbose" || arg == "-v" {
			isVerbose = true
		} else if arg == "--help" || arg == "-h" {
			printHelp()
			return
		}
	}

	if url == "" {
		url = "https://github.com/chukfi/frontend.git"
	}
	if directory == "" {
		directory = "./public"
	}

	for _, cmd := range requiredCommands {
		if !CheckIfCommandExists(cmd) {
			fmt.Printf(red+"Error: Required command '%s' not found in PATH. Please install it before running this tool.\n"+reset, cmd)
			return
		}
	}

	fmt.Printf(green+"Cloning frontend repository from %s...\n"+reset, url)
	// git clone the repo
	// build it with npm
	// move the build files to ./public

	// create temp directory
	dir, err := os.MkdirTemp("", "frontend-clone-")

	if err != nil {
		fmt.Println("Error creating temp directory:", err)
		return
	}

	defer os.RemoveAll(dir)

	cloneCmd := fmt.Sprintf("git clone %s %s", url, dir)
	printlnVerbose("Running command: " + cloneCmd)
	cmd := exec.Command("git", "clone", url, dir)

	if isVerbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error cloning repository:", err)
		return
	}

	// detect package manager by checking if package-lock.json (npm), yarn.lock (yarn), or pnpm-lock.yaml (pnpm) exists
	printlnVerbose("Detecting package manager...")
	packageManager := "npm"
	if _, err := os.Stat(fmt.Sprintf("%s/package-lock.json", dir)); err == nil {
		packageManager = "npm"
	} else if _, err := os.Stat(fmt.Sprintf("%s/yarn.lock", dir)); err == nil {
		packageManager = "yarn"
	} else if _, err := os.Stat(fmt.Sprintf("%s/pnpm-lock.yaml", dir)); err == nil {
		packageManager = "pnpm"
	}

	// ensure packageManager is installed on the system
	if !CheckIfCommandExists(packageManager) {
		// try install package manager with npm i -g <packageManager>
		fmt.Printf(red+"Package manager '%s' not found. Attempting to install it globally using npm...\n"+reset, packageManager)
		installCmd := exec.Command("npm", "install", "-g", packageManager)
		if isVerbose {
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
		}
		err = installCmd.Run()
		if err != nil {
			fmt.Printf(red+"Error installing package manager '%s': %v\n"+reset, packageManager, err)
			return
		}
		fmt.Printf(green+"Successfully installed package manager '%s'.\n"+reset, packageManager)
	}

	fmt.Printf(green+"Downloading packages with %s...\n"+reset, packageManager)

	installCmd := exec.Command(packageManager, "install")
	installCmd.Dir = dir
	if isVerbose {
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
	}
	err = installCmd.Run()
	if err != nil {
		fmt.Println("Error installing packages:", err)
		return
	}

	fmt.Printf(green+"Building frontend with %s...\n"+reset, packageManager)
	// set working directory to temp dir
	buildCmd := exec.Command(packageManager, "run", "build")
	buildCmd.Dir = dir
	if isVerbose {
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
	}
	err = buildCmd.Run()
	if err != nil {
		fmt.Println("Error building frontend:", err)
		return
	}
	fmt.Printf(green+"Moving build files to %s...\n"+reset, directory)

	// find which directory contains the build files
	buildDir := ""
	possibleDirs := []string{"build", "dist", "public"}

	for _, d := range possibleDirs {
		fullPath := fmt.Sprintf("%s/%s", dir, d)
		info, err := os.Stat(fullPath)
		if err == nil && info.IsDir() {
			buildDir = fullPath
			break
		}
	}

	if buildDir == "" {
		fmt.Printf(red + "Error: Could not find build directory (tried build/, dist/, public/)\n" + reset)
		return
	}

	// move buildDir to directory
	err = os.RemoveAll(directory)
	if err != nil {
		fmt.Printf(red+"Error removing existing directory: %v\n"+reset, err)
		return
	}
	err = os.Rename(buildDir, directory)
	if err != nil {
		fmt.Printf(red+"Error moving build files: %v\n"+reset, err)
		return
	}
	fmt.Printf(green+"Frontend successfully built and moved to %s\n"+reset, directory)
}
