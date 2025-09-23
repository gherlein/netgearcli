// poe_status_simple.go - Simple example using the ntgrrc library
// This version uses environment variables for automatic authentication.
//
// Usage:
//   export NETGEAR_SWITCHES="switch1=password123"
//   # OR export NETGEAR_PASSWORD_<HOST>=password123
//   go run poe_status_simple.go [--debug|-d] <switch-hostname>

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gherlein/ntgrrc/pkg/netgearcli/pkg/netgear"
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

	// Create client with optional debug logging - will auto-authenticate if environment variables are set
	var client *netgear.Client
	var err error

	if debug {
		client, err = netgear.NewClient(switchAddress, netgear.WithVerbose(true))
		fmt.Printf("Debug mode enabled\n")
		fmt.Printf("Connecting to: %s\n", switchAddress)
	} else {
		client, err = netgear.NewClient(switchAddress)
	}
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Check if we're already authenticated (auto-login happened)
	ctx := context.Background()
	if !client.IsAuthenticated() {
		// No auto-authentication occurred, try explicit login
		err = client.LoginAuto(ctx)
		if err != nil {
			log.Fatalf("Authentication failed: %v\nEnsure environment variables are set:\n  NETGEAR_PASSWORD_<HOST> or NETGEAR_SWITCHES", err)
		}
	}

	fmt.Printf("✓ Authenticated with %s (Model: %s)\n\n", switchAddress, client.GetModel())

	// Get POE status
	if debug {
		fmt.Println("Fetching POE status...")
	}
	statuses, err := client.POE().GetStatus(ctx)
	if err != nil {
		log.Fatalf("Failed to get POE status: %v", err)
	}

	if debug {
		fmt.Printf("Retrieved data for %d ports\n", len(statuses))
	}

	// Print results
	fmt.Println("POE Port Status:")
	for _, status := range statuses {
		fmt.Printf("Port %d: %s", status.PortID, status.Status)
		if status.PowerW > 0 {
			fmt.Printf(" (%.1fW)", status.PowerW)
		}
		if status.PortName != "" {
			fmt.Printf(" - %s", status.PortName)
		}
		fmt.Println()
	}
}