// poe_management.go - Example showing POE management operations
// This example demonstrates reading settings, updating configuration,
// and cycling power on POE ports.
//
// Usage: go run poe_management.go [--debug|-d] <switch-hostname> <command> [port-numbers...]
// Commands: status, settings, enable, disable, cycle

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gherlein/ntgrrc/pkg/netgearcli/pkg/netgear"
)

func main() {
	// Parse command line flags
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.BoolVar(&debug, "d", false, "Enable debug output (shorthand)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switchAddress := args[0]
	command := args[1]

	// Create client with file-based token manager for persistence and optional debug
	tokenManager := netgear.NewFileTokenManager("")
	var client *netgear.Client
	var err error

	if debug {
		client, err = netgear.NewClient(
			switchAddress,
			netgear.WithTokenManager(tokenManager),
			netgear.WithVerbose(true),
		)
	} else {
		client, err = netgear.NewClient(
			switchAddress,
			netgear.WithTokenManager(tokenManager),
		)
	}
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Auto-authenticate (will use cached token if available, or environment variables)
	ctx := context.Background()
	if debug {
		fmt.Printf("Debug mode enabled\n")
		fmt.Printf("Switch: %s, Command: %s\n", switchAddress, command)
	}

	if !client.IsAuthenticated() {
		err = client.LoginAuto(ctx)
		if err != nil {
			log.Fatalf("Authentication failed: %v\nEnsure environment variables are set:\n  NETGEAR_SWITCHES=\"host=password;...\"\n  OR NETGEAR_PASSWORD_<HOST>=password", err)
		}
	}

	// Execute command
	switch command {
	case "status":
		showStatus(ctx, client, debug)
	case "settings":
		showSettings(ctx, client, debug)
	case "enable":
		enablePorts(ctx, client, args[2:], debug)
	case "disable":
		disablePorts(ctx, client, args[2:], debug)
	case "cycle":
		cyclePorts(ctx, client, args[2:], debug)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [--debug|-d] <switch-hostname> <command> [port-numbers...]

Commands:
  status   - Show POE status for all ports
  settings - Show POE settings for all ports
  enable   - Enable POE on specified ports
  disable  - Disable POE on specified ports
  cycle    - Power cycle specified ports

Options:
  --debug, -d  - Enable debug output

Examples:
  %s 192.168.1.10 status
  %s --debug 192.168.1.10 enable 1 2 3
  %s -d 192.168.1.10 cycle 5

Environment (choose one):
  NETGEAR_SWITCHES="host=password;..."          - Multi-switch configuration
  NETGEAR_PASSWORD_<HOST>=password              - Host-specific password
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func showStatus(ctx context.Context, client *netgear.Client, debug bool) {
	if debug {
		fmt.Println("Fetching POE status...")
	}

	// Use the real go-netgear command directly since it outputs properly formatted data
	_, err := client.POE().GetStatus(ctx)
	if err != nil {
		log.Fatalf("Failed to get POE status: %v", err)
	}

	// The real POE status has already been output to stdout by the go-netgear command
	// No need to format it again here
}

func showSettings(ctx context.Context, client *netgear.Client, debug bool) {
	if debug {
		fmt.Println("Fetching POE settings...")
	}

	// Use the real go-netgear command directly since it outputs properly formatted data
	_, err := client.POE().GetSettings(ctx)
	if err != nil {
		log.Fatalf("Failed to get POE settings: %v", err)
	}

	// The real POE settings have already been output to stdout by the go-netgear command
	// No need to format them again here
}

func enablePorts(ctx context.Context, client *netgear.Client, portArgs []string, debug bool) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		log.Fatal("No port numbers specified")
	}

	if debug {
		fmt.Printf("Enabling POE on %d ports: %v\n", len(ports), ports)
	}

	for _, port := range ports {
		if debug {
			fmt.Printf("Enabling port %d...\n", port)
		}
		err := client.POE().EnablePort(ctx, port)
		if err != nil {
			log.Printf("Failed to enable port %d: %v", port, err)
		} else {
			fmt.Printf("✓ Enabled POE on port %d\n", port)
		}
	}
}

func disablePorts(ctx context.Context, client *netgear.Client, portArgs []string, debug bool) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		log.Fatal("No port numbers specified")
	}

	if debug {
		fmt.Printf("Disabling POE on %d ports: %v\n", len(ports), ports)
	}

	for _, port := range ports {
		if debug {
			fmt.Printf("Disabling port %d...\n", port)
		}
		err := client.POE().DisablePort(ctx, port)
		if err != nil {
			log.Printf("Failed to disable port %d: %v", port, err)
		} else {
			fmt.Printf("✓ Disabled POE on port %d\n", port)
		}
	}
}

func cyclePorts(ctx context.Context, client *netgear.Client, portArgs []string, debug bool) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		log.Fatal("No port numbers specified")
	}

	if debug {
		fmt.Printf("Power cycling %d ports: %v\n", len(ports), ports)
	}

	fmt.Printf("Power cycling %d ports...\n", len(ports))
	err := client.POE().CyclePower(ctx, ports...)
	if err != nil {
		log.Fatalf("Failed to cycle power: %v", err)
	}

	fmt.Println("✓ Power cycle completed")
}

func parsePorts(args []string) []int {
	var ports []int
	for _, arg := range args {
		// Handle comma-separated lists
		for _, p := range strings.Split(arg, ",") {
			port, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				log.Fatalf("Invalid port number: %s", p)
			}
			ports = append(ports, port)
		}
	}
	return ports
}