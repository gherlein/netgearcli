// poe_management.go - Example showing POE management operations
// This example demonstrates reading settings, updating configuration,
// and cycling power on POE ports.
//
// Usage: go run poe_management.go [--debug|-d] <switch-hostname> <command> [port-numbers...]
// Commands: status, settings, enable, disable, cycle

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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
	if len(args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switchAddress := args[0]
	command := args[1]

	if debug {
		fmt.Printf("Debug mode enabled\n")
		fmt.Printf("Switch: %s, Command: %s\n", switchAddress, command)
	}

	// Set up global options for all commands
	globalOpts := &go_netgear.GlobalOptions{
		Verbose:      debug,
		OutputFormat: go_netgear.JsonFormat,
	}

	// Try to login first - the command will use cached token if available
	// or check environment variables for password
	loginCmd := &go_netgear.LoginCommand{
		Address: switchAddress,
		// Password will be empty - command will check token cache and env vars
	}

	err := loginCmd.Run(globalOpts)
	if err != nil {
		// If login fails, it might be because we're already logged in (have a valid token)
		// or need to provide credentials. The subsequent commands will fail if not authenticated
		if debug {
			fmt.Printf("Login attempt result: %v\n", err)
		}

		// Check if the error is about authentication
		if strings.Contains(err.Error(), "no session") || strings.Contains(err.Error(), "login") {
			log.Fatalf("Authentication failed: %v\nEnsure environment variables are set:\n  NETGEAR_SWITCHES=\"host=password;...\"\n  OR NETGEAR_PASSWORD_<HOST>=password", err)
		}
	}

	// Execute command
	switch command {
	case "status":
		showStatus(globalOpts, switchAddress)
	case "settings":
		showSettings(globalOpts, switchAddress)
	case "enable":
		enablePorts(globalOpts, switchAddress, args[2:])
	case "disable":
		disablePorts(globalOpts, switchAddress, args[2:])
	case "cycle":
		cyclePorts(globalOpts, switchAddress, args[2:])
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

func showStatus(globalOpts *go_netgear.GlobalOptions, switchAddress string) {
	cmd := &go_netgear.PoeStatusCommand{
		Address: switchAddress,
	}

	err := cmd.Run(globalOpts)
	if err != nil {
		log.Fatalf("Failed to get POE status: %v", err)
	}
}

func showSettings(globalOpts *go_netgear.GlobalOptions, switchAddress string) {
	cmd := &go_netgear.PoeShowSettingsCommand{
		Address: switchAddress,
	}

	err := cmd.Run(globalOpts)
	if err != nil {
		log.Fatalf("Failed to get POE settings: %v", err)
	}
}

func enablePorts(globalOpts *go_netgear.GlobalOptions, switchAddress string, portArgs []string) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		log.Fatal("No port numbers specified")
	}

	cmd := &go_netgear.PoeSetConfigCommand{
		Address: switchAddress,
		Ports:   ports,
		PortPwr: "enable",
	}

	err := cmd.Run(globalOpts)
	if err != nil {
		log.Fatalf("Failed to enable POE on ports %v: %v", ports, err)
	}

	fmt.Printf("✓ Enabled POE on ports %v\n", ports)
}

func disablePorts(globalOpts *go_netgear.GlobalOptions, switchAddress string, portArgs []string) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		log.Fatal("No port numbers specified")
	}

	cmd := &go_netgear.PoeSetConfigCommand{
		Address: switchAddress,
		Ports:   ports,
		PortPwr: "disable",
	}

	err := cmd.Run(globalOpts)
	if err != nil {
		log.Fatalf("Failed to disable POE on ports %v: %v", ports, err)
	}

	fmt.Printf("✓ Disabled POE on ports %v\n", ports)
}

func cyclePorts(globalOpts *go_netgear.GlobalOptions, switchAddress string, portArgs []string) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		log.Fatal("No port numbers specified")
	}

	cmd := &go_netgear.PoeCyclePowerCommand{
		Address: switchAddress,
		Ports:   ports,
	}

	err := cmd.Run(globalOpts)
	if err != nil {
		log.Fatalf("Failed to cycle power on ports %v: %v", ports, err)
	}

	fmt.Printf("✓ Power cycle completed on ports %v\n", ports)
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