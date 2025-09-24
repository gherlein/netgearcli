// poe_status.go - Example program showing how to use the go-netgear library
// to login to a Netgear switch and display POE status for all ports.
//
// Usage: go run poe_status.go [--debug|-d] <switch-hostname>
//
// This example demonstrates:
// - Creating commands with the library
// - Authenticating with password prompt
// - Fetching POE status
// - Displaying the results

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
	"github.com/gherlein/go-netgear"
)

func main() {
	// Parse command line flags
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.BoolVar(&debug, "d", false, "Enable debug output (shorthand)")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--debug|-d] <switch-hostname>\n", os.Args[0])
		os.Exit(1)
	}

	switchAddress := args[0]

	fmt.Printf("Connecting to switch at %s...\n", switchAddress)

	if debug {
		fmt.Println("Debug mode enabled")
		fmt.Printf("Switch address: %s\n", switchAddress)
	}

	// Set up global options
	globalOpts := &go_netgear.GlobalOptions{
		Verbose:      debug,
		OutputFormat: go_netgear.JsonFormat,
	}

	// Try to login first
	// The LoginCommand will check if there's an existing token and use it if valid
	// Otherwise it will try environment variables or prompt for password
	loginCmd := &go_netgear.LoginCommand{
		Address: switchAddress,
		// Password will be empty - command will check env vars or prompt
	}

	err := loginCmd.Run(globalOpts)
	if err != nil {
		// Login might fail if already logged in or if env vars are set
		// Try to continue anyway, the POE command will fail if not authenticated
		if debug {
			fmt.Printf("Login attempt: %v\n", err)
		}

		// If login failed and no environment variables, prompt for password
		if strings.Contains(err.Error(), "no password") || strings.Contains(err.Error(), "authentication") {
			fmt.Print("Enter admin password: ")
			password, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %v", err)
			}
			fmt.Println() // New line after password input

			// Try login again with password
			loginCmd.Password = password
			err = loginCmd.Run(globalOpts)
			if err != nil {
				log.Fatalf("Login failed: %v", err)
			}
		}
	}

	fmt.Printf("Successfully connected to %s\n", switchAddress)

	// Get POE status for all ports
	fmt.Println("\nFetching POE status...")
	if debug {
		fmt.Println("Making request to POE status endpoint...")
	}

	cmd := &go_netgear.PoeStatusCommand{
		Address: switchAddress,
	}

	err = cmd.Run(globalOpts)
	if err != nil {
		log.Fatalf("Failed to get POE status: %v", err)
	}

	if debug {
		fmt.Println("POE status retrieval completed")
	}
}

// readPassword reads a password from stdin without echoing
func readPassword() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePassword)), nil
}