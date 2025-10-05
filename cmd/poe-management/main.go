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
	"hash/adler32"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	go_netgear "github.com/gherlein/go-netgear"
)

// Global variables for authentication
var (
	globalPassword     string
	globalSwitchAddr   string
	globalDebug        bool
	globalOpts         *go_netgear.GlobalOptions
	loginAttempted     bool
	logFile            *os.File
	logger             *log.Logger
)

func main() {
	// Parse command line flags
	var password string
	var logFilePath string
	flag.BoolVar(&globalDebug, "debug", false, "Enable debug output")
	flag.BoolVar(&globalDebug, "d", false, "Enable debug output (shorthand)")
	flag.StringVar(&password, "password", "", "Admin password for authentication")
	flag.StringVar(&password, "p", "", "Admin password for authentication (shorthand)")
	flag.StringVar(&logFilePath, "log", "", "Log file path for activity logging")
	flag.StringVar(&logFilePath, "l", "", "Log file path for activity logging (shorthand)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Set up logging if log file specified
	if logFilePath != "" {
		var err error
		logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file %s: %v", logFilePath, err)
		}
		defer logFile.Close()

		// Create logger with timestamp
		logger = log.New(logFile, "", log.LstdFlags)

		// Log the command line immediately
		cmdLine := strings.Join(os.Args, " ")
		logger.Printf("Command: %s", cmdLine)
	}

	globalSwitchAddr = args[0]
	command := args[1]

	if globalDebug {
		fmt.Printf("Debug mode enabled\n")
		fmt.Printf("Switch: %s, Command: %s\n", globalSwitchAddr, command)
	}
	logMessage("Debug mode: %v, Switch: %s, Command: %s", globalDebug, globalSwitchAddr, command)

	// Set up global options for all commands
	globalOpts = &go_netgear.GlobalOptions{
		Verbose:      globalDebug,
		OutputFormat: go_netgear.JsonFormat,
	}

	// Priority: 1. CLI flag, 2. Environment variable
	if password == "" {
		password = getPasswordFromEnv(globalSwitchAddr, globalDebug)
	}
	globalPassword = password

	// Ensure we're logged in before executing commands
	err := ensureAuthenticated()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Execute command
	switch command {
	case "status":
		showStatus(globalOpts, globalSwitchAddr)
	case "settings":
		showSettings(globalOpts, globalSwitchAddr)
	case "enable":
		enablePorts(globalOpts, globalSwitchAddr, args[2:])
	case "disable":
		disablePorts(globalOpts, globalSwitchAddr, args[2:])
	case "cycle":
		cyclePorts(globalOpts, globalSwitchAddr, args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// logMessage logs a message to the log file if logging is enabled
func logMessage(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf(format, args...)
	}
}

// ensureAuthenticated ensures we have a valid session, logging in if necessary
func ensureAuthenticated() error {
	logMessage("Ensuring authentication for %s", globalSwitchAddr)
	// Check if cached token exists
	if hasValidToken(globalSwitchAddr, globalDebug) {
		// Validate the token with a keep-alive check
		if validateToken() {
			if globalDebug {
				fmt.Printf("Using cached token\n")
			}
			logMessage("Using cached token for %s", globalSwitchAddr)
			return nil
		}

		// Token is invalid, remove it
		if globalDebug {
			fmt.Printf("Cached token is invalid, will re-login\n")
		}
		logMessage("Cached token invalid, re-authenticating to %s", globalSwitchAddr)
		removeToken(globalSwitchAddr)
	}

	// No valid token, need to login
	return performLogin()
}

// performLogin executes the login command with retry logic
func performLogin() error {
	if globalPassword == "" {
		logMessage("Login failed: no password available for %s", globalSwitchAddr)
		return fmt.Errorf("no password available for authentication")
	}

	// Retry delays: 200ms, 500ms, 1000ms
	retryDelays := []time.Duration{200 * time.Millisecond, 500 * time.Millisecond, 1000 * time.Millisecond}
	maxAttempts := len(retryDelays) + 1 // Initial attempt + retries

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if globalDebug {
			if attempt == 1 {
				fmt.Printf("Logging in to %s...\n", globalSwitchAddr)
			} else {
				fmt.Printf("Retry attempt %d of %d for %s...\n", attempt-1, len(retryDelays), globalSwitchAddr)
			}
		}
		if attempt == 1 {
			logMessage("Logging in to %s", globalSwitchAddr)
		} else {
			logMessage("Retry attempt %d of %d for %s", attempt-1, len(retryDelays), globalSwitchAddr)
		}

		loginCmd := &go_netgear.LoginCommand{
			Address:  globalSwitchAddr,
			Password: globalPassword,
		}

		err := loginCmd.Run(globalOpts)
		if err == nil {
			loginAttempted = true
			if globalDebug {
				fmt.Printf("Login successful\n")
			}
			logMessage("Login successful to %s", globalSwitchAddr)
			return nil
		}

		lastErr = err
		if globalDebug {
			fmt.Printf("Login attempt %d failed: %v\n", attempt, err)
		}
		logMessage("Login attempt %d failed for %s: %v", attempt, globalSwitchAddr, err)

		// If we have more attempts remaining, wait before retrying
		if attempt < maxAttempts {
			delay := retryDelays[attempt-1]
			if globalDebug {
				fmt.Printf("Waiting %v before retry...\n", delay)
			}
			logMessage("Waiting %v before retry", delay)
			time.Sleep(delay)
		}
	}

	// All attempts failed
	logMessage("Giving up after %d failed login attempts for %s: %v", maxAttempts, globalSwitchAddr, lastErr)
	if globalDebug {
		fmt.Printf("Giving up after %d failed login attempts\n", maxAttempts)
	}
	return fmt.Errorf("login failed after %d attempts: %w", maxAttempts, lastErr)
}

// validateToken checks if the current token is still valid by making a lightweight request
func validateToken() bool {
	if globalDebug {
		fmt.Printf("Validating cached token...\n")
	}

	// Make a lightweight status request to check if token is valid
	cmd := &go_netgear.PoeStatusCommand{
		Address: globalSwitchAddr,
	}

	// Temporarily disable verbose AND redirect output to suppress JSON
	savedVerbose := globalOpts.Verbose
	savedFormat := globalOpts.OutputFormat
	globalOpts.Verbose = false
	globalOpts.OutputFormat = go_netgear.MarkdownFormat // Quieter than JSON

	// Redirect stdout to discard output during validation
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Consume pipe output in goroutine to prevent blocking
	done := make(chan bool)
	go func() {
		io.Copy(io.Discard, r)
		done <- true
	}()

	err := cmd.Run(globalOpts)
	w.Close()
	<-done
	os.Stdout = oldStdout

	globalOpts.Verbose = savedVerbose
	globalOpts.OutputFormat = savedFormat

	if err != nil {
		// Check if error indicates login required
		if strings.Contains(err.Error(), "no content") ||
		   strings.Contains(err.Error(), "login") ||
		   strings.Contains(err.Error(), "no session") {
			if globalDebug {
				fmt.Printf("Token validation failed: %v\n", err)
			}
			return false
		}
	}

	if globalDebug {
		fmt.Printf("Token is valid\n")
	}
	return true
}

// handleAuthError executes a function and retries with re-login if authentication fails
func handleAuthError(retryFunc func() error) error {
	// Execute the function
	err := retryFunc()
	if err == nil {
		return nil
	}

	// Check if error indicates authentication issue
	errStr := err.Error()
	if strings.Contains(errStr, "no content") ||
	   strings.Contains(errStr, "login") ||
	   strings.Contains(errStr, "no session") {

		if loginAttempted {
			// Already tried to login once, don't retry infinitely
			return fmt.Errorf("authentication failed even after re-login: %w", err)
		}

		if globalDebug {
			fmt.Printf("Token expired or invalid, re-authenticating...\n")
		}

		// Remove invalid token
		removeToken(globalSwitchAddr)

		// Try to login again
		loginErr := performLogin()
		if loginErr != nil {
			return fmt.Errorf("re-authentication failed: %w", loginErr)
		}

		// Retry the original operation
		return retryFunc()
	}

	return err
}

// removeToken deletes the cached token file
func removeToken(address string) {
	tokenPath := getTokenPath(os.TempDir(), address)
	os.Remove(tokenPath)
	if globalDebug {
		fmt.Printf("Removed invalid token at %s\n", tokenPath)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [options] <switch-hostname> <command> [port-numbers...]

Commands:
  status   - Show POE status for all ports
  settings - Show POE settings for all ports
  enable   - Enable POE on specified ports
  disable  - Disable POE on specified ports
  cycle    - Power cycle specified ports

Options:
  --debug, -d       - Enable debug output
  --password, -p    - Admin password for authentication
  --log, -l         - Log file path for activity logging

Examples:
  %s 192.168.1.10 status
  %s --password mypass 192.168.1.10 enable 1 2 3
  %s -p mypass 192.168.1.10 enable 1-8           - Enable ports 1 through 8
  %s 192.168.1.10 disable 1-8 14-16              - Disable ports 1-8 and 14-16
  %s -d 192.168.1.10 cycle 5

Authentication (in priority order):
  1. --password/-p flag                          - Passed on command line
  2. NETGEAR_SWITCHES="host:password;..."        - Multi-switch configuration
  3. NETGEAR_PASSWORD_<HOST>=password            - Host-specific password
  4. Cached token from previous login            - Stored in /tmp/.config/ntgrrc/
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func showStatus(globalOpts *go_netgear.GlobalOptions, switchAddress string) {
	if globalDebug {
		fmt.Printf("Executing status command...\n")
	}
	logMessage("Executing status command on %s", switchAddress)

	err := handleAuthError(func() error {
		if globalDebug {
			fmt.Printf("Creating PoeStatusCommand...\n")
		}
		cmd := &go_netgear.PoeStatusCommand{
			Address: switchAddress,
		}
		if globalDebug {
			fmt.Printf("Running PoeStatusCommand...\n")
		}
		err := cmd.Run(globalOpts)
		if globalDebug {
			fmt.Printf("PoeStatusCommand completed with err=%v\n", err)
		}
		return err
	})

	if err != nil {
		logMessage("Failed to get POE status from %s: %v", switchAddress, err)
		log.Fatalf("Failed to get POE status: %v", err)
	}
	logMessage("Successfully retrieved POE status from %s", switchAddress)
}

func showSettings(globalOpts *go_netgear.GlobalOptions, switchAddress string) {
	logMessage("Executing settings command on %s", switchAddress)
	err := handleAuthError(func() error {
		cmd := &go_netgear.PoeShowSettingsCommand{
			Address: switchAddress,
		}
		return cmd.Run(globalOpts)
	})

	if err != nil {
		logMessage("Failed to get POE settings from %s: %v", switchAddress, err)
		log.Fatalf("Failed to get POE settings: %v", err)
	}
	logMessage("Successfully retrieved POE settings from %s", switchAddress)
}

func enablePorts(globalOpts *go_netgear.GlobalOptions, switchAddress string, portArgs []string) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		logMessage("Enable ports failed: no port numbers specified")
		log.Fatal("No port numbers specified")
	}

	logMessage("Enabling POE on %s ports %v", switchAddress, ports)
	err := handleAuthError(func() error {
		cmd := &go_netgear.PoeSetConfigCommand{
			Address: switchAddress,
			Ports:   ports,
			PortPwr: "enable",
		}
		return cmd.Run(globalOpts)
	})

	if err != nil {
		logMessage("Failed to enable POE on %s ports %v: %v", switchAddress, ports, err)
		log.Fatalf("Failed to enable POE on ports %v: %v", ports, err)
	}

	fmt.Printf("✓ Enabled POE on ports %v\n", ports)
	logMessage("Successfully enabled POE on %s ports %v", switchAddress, ports)
}

func disablePorts(globalOpts *go_netgear.GlobalOptions, switchAddress string, portArgs []string) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		logMessage("Disable ports failed: no port numbers specified")
		log.Fatal("No port numbers specified")
	}

	if globalDebug {
		fmt.Printf("Disabling POE on ports: %v\n", ports)
	}
	logMessage("Disabling POE on %s ports %v", switchAddress, ports)

	err := handleAuthError(func() error {
		cmd := &go_netgear.PoeSetConfigCommand{
			Address: switchAddress,
			Ports:   ports,
			PortPwr: "disable",
		}
		if globalDebug {
			fmt.Printf("Running PoeSetConfigCommand with PortPwr=%q\n", "disable")
		}
		err := cmd.Run(globalOpts)
		if globalDebug {
			if err != nil {
				fmt.Printf("PoeSetConfigCommand returned error: %v\n", err)
			} else {
				fmt.Printf("PoeSetConfigCommand completed successfully\n")
			}
		}
		return err
	})

	if err != nil {
		logMessage("Failed to disable POE on %s ports %v: %v", switchAddress, ports, err)
		log.Fatalf("Failed to disable POE on ports %v: %v", ports, err)
	}

	fmt.Printf("✓ Disabled POE on ports %v\n", ports)
	logMessage("Successfully disabled POE on %s ports %v", switchAddress, ports)
}

func cyclePorts(globalOpts *go_netgear.GlobalOptions, switchAddress string, portArgs []string) {
	ports := parsePorts(portArgs)
	if len(ports) == 0 {
		logMessage("Cycle ports failed: no port numbers specified")
		log.Fatal("No port numbers specified")
	}

	logMessage("Power cycling POE on %s ports %v", switchAddress, ports)
	err := handleAuthError(func() error {
		cmd := &go_netgear.PoeCyclePowerCommand{
			Address: switchAddress,
			Ports:   ports,
		}
		return cmd.Run(globalOpts)
	})

	if err != nil {
		logMessage("Failed to cycle power on %s ports %v: %v", switchAddress, ports, err)
		log.Fatalf("Failed to cycle power on ports %v: %v", ports, err)
	}

	fmt.Printf("✓ Power cycle completed on ports %v\n", ports)
	logMessage("Successfully cycled power on %s ports %v", switchAddress, ports)
}

func parsePorts(args []string) []int {
	var ports []int
	seen := make(map[int]bool) // Track seen ports to avoid duplicates

	for _, arg := range args {
		// Handle comma-separated lists
		for _, p := range strings.Split(arg, ",") {
			p = strings.TrimSpace(p)

			// Check if it's a range (e.g., "1-8")
			if strings.Contains(p, "-") {
				rangeParts := strings.SplitN(p, "-", 2)
				if len(rangeParts) != 2 {
					log.Fatalf("Invalid port range: %s", p)
				}

				start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				if err != nil {
					log.Fatalf("Invalid port range start: %s", rangeParts[0])
				}

				end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err != nil {
					log.Fatalf("Invalid port range end: %s", rangeParts[1])
				}

				if start > end {
					log.Fatalf("Invalid port range %s: start must be <= end", p)
				}

				// Add all ports in the range
				for port := start; port <= end; port++ {
					if !seen[port] {
						ports = append(ports, port)
						seen[port] = true
					}
				}
			} else {
				// Single port number
				port, err := strconv.Atoi(p)
				if err != nil {
					log.Fatalf("Invalid port number: %s", p)
				}

				if !seen[port] {
					ports = append(ports, port)
					seen[port] = true
				}
			}
		}
	}
	return ports
}

// getPasswordFromEnv checks for password in environment variables
// hasValidToken checks if a cached token file exists for the given address
func hasValidToken(address string, debug bool) bool {
	tokenDir := os.TempDir()
	tokenPath := getTokenPath(tokenDir, address)

	_, err := os.Stat(tokenPath)
	exists := err == nil

	if debug {
		if exists {
			fmt.Printf("Found cached token at %s\n", tokenPath)
		} else {
			fmt.Printf("No cached token found at %s\n", tokenPath)
		}
	}

	return exists
}

// getTokenPath returns the expected token file path for a given address
// This mirrors the logic in the go-netgear library
func getTokenPath(configDir string, host string) string {
	// Using adler32 hash to match library behavior
	hash32 := adler32.New()
	io.WriteString(hash32, host)
	hash := fmt.Sprintf("%x", hash32.Sum(nil))

	if configDir == "" {
		configDir = os.TempDir()
	}
	dotConfigDir := configDir + "/.config/ntgrrc"
	return dotConfigDir + "/token-" + hash
}

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
	switches := os.Getenv("NETGEAR_SWITCHES")
	if debug {
		fmt.Printf("NETGEAR_SWITCHES environment variable: %q\n", switches)
	}

	if switches != "" {
		for _, entry := range strings.Split(switches, ";") {
			if debug {
				fmt.Printf("  Checking entry: %q\n", entry)
			}
			parts := strings.SplitN(entry, ":", 2)
			if debug {
				fmt.Printf("  Split into %d parts: %v\n", len(parts), parts)
			}
			if len(parts) == 2 {
				host := strings.TrimSpace(parts[0])
				pass := strings.TrimSpace(parts[1])
				if debug {
					fmt.Printf("  Parsed: host=%q pass=%q (looking for %q)\n", host, pass, address)
				}
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