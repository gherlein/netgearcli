// poe_status_simple.go - Simple example using the go-netgear library
// This version uses environment variables for automatic authentication.
//
// Usage:
//   export NETGEAR_SWITCHES="switch1=password123"
//   # OR export NETGEAR_PASSWORD_<HOST>=password123
//   go run poe_status_simple.go [--debug|-d] <switch-hostname>

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
		fmt.Fprintf(os.Stderr, "Set environment variables:\n")
		fmt.Fprintf(os.Stderr, "  NETGEAR_SWITCHES=\"host=password;...\"\n")
		fmt.Fprintf(os.Stderr, "  OR NETGEAR_PASSWORD_<HOST>=password\n")
		os.Exit(1)
	}

	switchAddress := args[0]

	if debug {
		fmt.Printf("Debug mode enabled\n")
		fmt.Printf("Connecting to: %s\n", switchAddress)
	}

	// Set up global options
	globalOpts := &go_netgear.GlobalOptions{
		Verbose:      debug,
		OutputFormat: go_netgear.JsonFormat,
	}

	// Try to login - will use cached token, environment variables, or fail
	loginCmd := &go_netgear.LoginCommand{
		Address: switchAddress,
		// Password will be empty - command will check token cache and env vars
	}

	err := loginCmd.Run(globalOpts)
	if err != nil {
		// If login fails, check if it's an authentication issue
		if debug {
			fmt.Printf("Login attempt result: %v\n", err)
		}

		if strings.Contains(err.Error(), "no session") || strings.Contains(err.Error(), "login") || strings.Contains(err.Error(), "password") {
			log.Fatalf("Authentication failed: %v\nEnsure environment variables are set:\n  NETGEAR_SWITCHES=\"host=password;...\"\n  OR NETGEAR_PASSWORD_<HOST>=password", err)
		}
	}

	fmt.Printf("✓ Authenticated with %s\n\n", switchAddress)

	// Get POE status using the real go-netgear command
	cmd := &go_netgear.PoeStatusCommand{
		Address: switchAddress,
	}

	err = cmd.Run(globalOpts)
	if err != nil {
		log.Fatalf("Failed to get POE status: %v", err)
	}
}