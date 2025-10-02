// poe_status_simple.go - Simple example using the go-netgear library
// This version uses environment variables for automatic authentication.
//
// Usage:
//   export NETGEAR_PASSWORD_TSWITCH1="password123"
//   # OR export NETGEAR_SWITCHES="switch1=password123;switch2=password456"
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
		fmt.Fprintf(os.Stderr, "  NETGEAR_PASSWORD_<hostname>=password\n")
		fmt.Fprintf(os.Stderr, "  OR NETGEAR_SWITCHES=\"host:password;...\"\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  export NETGEAR_PASSWORD_tswitch1=\"None1234@\"\n")
		fmt.Fprintf(os.Stderr, "  export NETGEAR_SWITCHES=\"tswitch1:None1234@;tswitch2:None1234@\"\n")
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

	// Try to get password from environment variables
	password := getPasswordFromEnv(switchAddress, debug)

	// Try to login with password from environment or empty (will use cached token)
	loginCmd := &go_netgear.LoginCommand{
		Address:  switchAddress,
		Password: password, // If empty, LoginCommand will prompt
	}

	err := loginCmd.Run(globalOpts)
	if err != nil {
		// If login fails, check if it's an authentication issue
		if debug {
			fmt.Printf("Login attempt result: %v\n", err)
		}

		if strings.Contains(err.Error(), "no session") || strings.Contains(err.Error(), "login") || strings.Contains(err.Error(), "password") {
			log.Fatalf("Authentication failed: %v\nEnsure environment variables are set correctly", err)
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

// getPasswordFromEnv checks for password in environment variables
func getPasswordFromEnv(address string, debug bool) string {
	// Check NETGEAR_PASSWORD_<hostname> format
	envVar := "NETGEAR_PASSWORD_" + address
	if password := os.Getenv(envVar); password != "" {
		if debug {
			fmt.Printf("Found password in environment variable %s\n", envVar)
		}
		return password
	}

	// Check NETGEAR_SWITCHES format: "host1:password1;host2:password2"
	if switches := os.Getenv("NETGEAR_SWITCHES"); switches != "" {
		for _, entry := range strings.Split(switches, ";") {
			parts := strings.SplitN(entry, ":", 2)
			if len(parts) == 2 {
				host := strings.TrimSpace(parts[0])
				pass := strings.TrimSpace(parts[1])
				if host == address {
					if debug {
						fmt.Printf("Found password for %s in NETGEAR_SWITCHES\n", address)
					}
					return pass
				}
			}
		}
	}

	if debug {
		fmt.Printf("No password found in environment variables for %s\n", address)
		fmt.Printf("Checked: %s and NETGEAR_SWITCHES\n", envVar)
	}

	return ""
}