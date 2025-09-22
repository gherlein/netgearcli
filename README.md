# Netgear CLI Examples

This directory contains standalone example programs that demonstrate how to use the `ntgrrc` library to interact with Netgear switches.

## Prerequisites

1. Install Go 1.23 or later
2. Set up the library dependency by either:
   - Using the local development version: `go mod edit -replace ntgrrc=../../ntgrrc`
   - Or use a published version: `go get github.com/your-org/ntgrrc`

## Examples

### poe_status.go
Shows comprehensive POE status for all ports with formatted table output.

**Features:**
- Interactive password prompt if environment variables not set
- Detailed table showing port status, power usage, temperature, etc.
- Summary statistics
- Debug mode support

**Usage:**
```bash
go run poe_status.go [--debug|-d] <switch-hostname>
```

### poe_management.go
Demonstrates POE management operations including enabling/disabling ports and power cycling.

**Features:**
- Token-based authentication persistence
- Multiple commands: status, settings, enable, disable, cycle
- Batch port operations
- Environment variable authentication

**Usage:**
```bash
go run poe_management.go [--debug|-d] <switch-hostname> <command> [port-numbers...]

# Examples:
go run poe_management.go 192.168.1.10 status
go run poe_management.go --debug 192.168.1.10 enable 1 2 3
go run poe_management.go 192.168.1.10 cycle 5
```

### poe_status_simple.go
Minimal example focused on environment variable authentication.

**Features:**
- Simple POE status display
- Environment variable authentication only
- Minimal dependencies

**Usage:**
```bash
# Set environment variables first:
export NETGEAR_SWITCHES="192.168.1.10=password123"
# OR
export NETGEAR_PASSWORD_192_168_1_10=password123

go run poe_status_simple.go [--debug|-d] <switch-hostname>
```

## Authentication

All examples support multiple authentication methods:

### Environment Variables

**Method 1: Multi-switch configuration**
```bash
export NETGEAR_SWITCHES="switch1=password1;switch2=password2"
```

**Method 2: Host-specific password**
```bash
export NETGEAR_PASSWORD_<HOST>=password
# Example:
export NETGEAR_PASSWORD_192_168_1_10=mypassword
```

### Interactive Password Prompt
If no environment variables are set, `poe_status.go` will prompt for a password.

### Token Persistence
The `poe_management.go` example uses file-based token storage to avoid repeated authentication.

## Running the Examples

1. Navigate to the examples directory:
   ```bash
   cd netgearcli/examples
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run an example:
   ```bash
   go run poe_status.go 192.168.1.10
   ```

## Debug Mode

All examples support debug mode with `--debug` or `-d` flags to see:
- HTTP requests and responses
- Authentication details
- Internal library operations
- Detailed error information

Example:
```bash
go run poe_status.go --debug 192.168.1.10
```