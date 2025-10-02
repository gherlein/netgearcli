// poe_status.go - Example program showing how to use the go-netgear library
// to login to a Netgear switch and display POE status for all ports.
//
// Usage: go run poe_status.go [--debug|-d] <switch-hostname>
//
// This example demonstrates:
// - Creating commands with the library
// - Checking environment variables before prompting for password
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

	// Check for password in environment variables first
	password := getPasswordFromEnv(switchAddress, debug)

	// Try to login - use env password if found, otherwise LoginCommand will prompt
	loginCmd := &go_netgear.LoginCommand{
		Address:  switchAddress,
		Password: password,
	}

	err := loginCmd.Run(globalOpts)
	if err != nil {
		// If login failed and we didn't have an env password, try prompting
		if password == "" && (strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "authentication")) {
			fmt.Print("Enter admin password: ")
			promptPassword, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %v", err)
			}
			fmt.Println() // New line after password input

			// Try login again with prompted password
			loginCmd.Password = promptPassword
			err = loginCmd.Run(globalOpts)
			if err != nil {
				log.Fatalf("Login failed: %v", err)
			}
		} else {
			log.Fatalf("Login failed: %v", err)
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

// readPassword reads a password from stdin without echoing
func readPassword() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePassword)), nil
}