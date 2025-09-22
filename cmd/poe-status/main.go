// poe_status.go - Example program showing how to use the ntgrrc library
// to login to a Netgear switch and display POE status for all ports.
//
// Usage: go run poe_status.go [--debug|-d] <switch-hostname>
//
// This example demonstrates:
// - Creating a client with the library
// - Logging in with password prompt
// - Fetching POE status
// - Displaying the results in a formatted table

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
	"github.com/gherlein/ntgrrc/pkg/netgear"
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

	// Create a new client with optional debug output
	fmt.Printf("Connecting to switch at %s...\n", switchAddress)
	var client *netgear.Client
	var err error

	if debug {
		client, err = netgear.NewClient(switchAddress,
			netgear.WithVerbose(true))
	} else {
		client, err = netgear.NewClient(switchAddress)
	}
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	if debug {
		fmt.Println("Debug mode enabled")
		fmt.Printf("Switch address: %s\n", switchAddress)
		fmt.Printf("Detected model: %s\n", client.GetModel())
	}

	// Check if we're already authenticated (auto-login happened)
	if !client.IsAuthenticated() {
		// Try environment variable authentication first
		err = client.LoginAuto(ctx)
		if err != nil {
			// No environment variables, prompt for password
			fmt.Print("Enter admin password: ")
			password, err := readPassword()
			if err != nil {
				log.Fatalf("Failed to read password: %v", err)
			}
			fmt.Println() // New line after password input

			fmt.Println("Logging in...")
			err = client.Login(ctx, password)
			if err != nil {
				log.Fatalf("Login failed: %v", err)
			}
		}
	}

	fmt.Printf("Successfully authenticated with %s (Model: %s)\n", switchAddress, client.GetModel())

	// Get POE status for all ports
	fmt.Println("\nFetching POE status...")
	if debug {
		fmt.Println("Making authenticated request to POE status endpoint...")
	}
	statuses, err := client.POE().GetStatus(ctx)
	if err != nil {
		log.Fatalf("Failed to get POE status: %v", err)
	}

	if debug {
		fmt.Printf("Retrieved POE data for %d ports\n", len(statuses))
	}

	// Display the results in a table
	printPOEStatusTable(statuses)

	// Display summary
	var totalPower float64
	var activePorts int
	for _, status := range statuses {
		totalPower += status.PowerW
		if status.Status == "Delivering Power" {
			activePorts++
		}
	}
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Active POE Ports: %d/%d\n", activePorts, len(statuses))
	fmt.Printf("  Total Power Usage: %.2f W\n", totalPower)
}

// readPassword reads a password from stdin without echoing
func readPassword() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePassword)), nil
}

// printPOEStatusTable prints POE status in a formatted table
func printPOEStatusTable(statuses []netgear.POEPortStatus) {
	// Define column headers and widths
	headers := []string{"Port", "Name", "Status", "Class", "Voltage(V)", "Current(mA)", "Power(W)", "Temp(°C)", "Error"}
	widths := []int{4, 16, 16, 5, 10, 11, 8, 8, 12}

	// Print header
	fmt.Println()
	printTableRow(headers, widths)
	printSeparator(widths)

	// Print each port status
	for _, status := range statuses {
		row := []string{
			fmt.Sprintf("%d", status.PortID),
			truncate(status.PortName, 16),
			status.Status,
			status.PowerClass,
			fmt.Sprintf("%.1f", status.VoltageV),
			fmt.Sprintf("%.0f", status.CurrentMA),
			fmt.Sprintf("%.2f", status.PowerW),
			fmt.Sprintf("%.0f", status.TemperatureC),
			status.ErrorStatus,
		}
		printTableRow(row, widths)
	}
}

// printTableRow prints a single row with proper spacing
func printTableRow(columns []string, widths []int) {
	fmt.Print("|")
	for i, col := range columns {
		fmt.Printf(" %-*s |", widths[i], col)
	}
	fmt.Println()
}

// printSeparator prints a separator line
func printSeparator(widths []int) {
	fmt.Print("|")
	for _, w := range widths {
		fmt.Print(strings.Repeat("-", w+2))
		fmt.Print("|")
	}
	fmt.Println()
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}