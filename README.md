# Netgear CLI Examples

This repository contains example command-line programs demonstrating how to use the [go-netgear](https://github.com/gherlein/go-netgear) library to interact with Netgear managed switches via their web API.

## Credits

This project builds upon the excellent work of [ntgrrc](https://github.com/nitram509/ntgrrc) by Martin W. Kirst, which provided the foundation for understanding and implementing the Netgear switch API protocol. The go-netgear library used by these examples is a Go implementation inspired by that project.

## Quick Start

### Building

Build all examples:
```bash
make build
```

This creates binaries in the `bin/` directory:
- `bin/poe-status` - Comprehensive POE status display
- `bin/poe-status-simple` - Simple POE status with environment auth
- `bin/poe-management` - Full POE management with multiple commands

### Authentication

All programs support multiple authentication methods. See [docs/login.md](docs/login.md) for detailed information about the authentication process, token caching, and security considerations.

#### Quick Authentication Setup

**Method 1: Environment Variable (Multi-switch)**
```bash
export NETGEAR_SWITCHES="switch1:password1;switch2:password2"
# Example:
export NETGEAR_SWITCHES="192.168.1.10:admin123;tswitch16:mypass"
```

**Method 2: Environment Variable (Single switch)**
```bash
export NETGEAR_PASSWORD_<hostname>=password
# Example:
export NETGEAR_PASSWORD_192_168_1_10=admin123
```

**Method 3: Interactive Prompt**
```bash
# poe-status and poe-status-simple will prompt for password if not set
./bin/poe-status 192.168.1.10
```

**Method 4: Command Line Flag**
```bash
# poe-management accepts password via flag
./bin/poe-management --password mypass 192.168.1.10 status
```

**Token Caching**: After successful authentication, a session token is cached in `/tmp/.config/ntgrrc/` to avoid re-authentication on subsequent commands. See [docs/login.md](docs/login.md) for details on token management and persistence options.

## Programs

### poe-status
Shows comprehensive POE status for all ports with formatted table output.

**Features:**
- Interactive password prompt if environment variables not set
- Detailed table showing port status, power usage, temperature, etc.
- Summary statistics
- Debug mode support

**Usage:**
```bash
./bin/poe-status [--debug|-d] <switch-hostname>

# Examples:
./bin/poe-status 192.168.1.10
./bin/poe-status --debug tswitch16
```

### poe-status-simple
Minimal example focused on environment variable authentication.

**Features:**
- Simple POE status display in JSON format
- Environment variable authentication only
- Minimal dependencies

**Usage:**
```bash
# Set authentication first:
export NETGEAR_SWITCHES="192.168.1.10:password123"

./bin/poe-status-simple [--debug|-d] <switch-hostname>

# Example:
./bin/poe-status-simple tswitch16
```

### poe-management
Comprehensive POE management tool with multiple commands.

**Features:**
- Token-based authentication with automatic caching
- Multiple commands: status, settings, enable, disable, cycle
- Batch port operations with range support (e.g., `1-8`)
- Environment variable and command-line authentication
- Automatic retry with re-authentication on token expiration

**Commands:**
- `status` - Show POE status for all ports (JSON format)
- `settings` - Show POE settings/configuration for all ports
- `enable` - Enable POE on specified ports
- `disable` - Disable POE on specified ports
- `cycle` - Power cycle specified ports

**Usage:**
```bash
./bin/poe-management [options] <switch-hostname> <command> [port-numbers...]

Options:
  --debug, -d       - Enable debug output
  --password, -p    - Admin password for authentication

# Examples:
./bin/poe-management 192.168.1.10 status
./bin/poe-management --password mypass 192.168.1.10 enable 1 2 3
./bin/poe-management 192.168.1.10 enable 1-8           # Enable ports 1 through 8
./bin/poe-management 192.168.1.10 disable 1-8 14-16    # Disable multiple ranges
./bin/poe-management --debug 192.168.1.10 cycle 5
```

**Port Ranges:**
You can specify individual ports, ranges, or combinations:
```bash
./bin/poe-management switch1 enable 1          # Single port
./bin/poe-management switch1 enable 1-8        # Range of ports 1 through 8
./bin/poe-management switch1 enable 1-8 14-16  # Multiple ranges
./bin/poe-management switch1 enable 1 3 5-8    # Mix of single and range
```

## Testing

An integration test script is provided to verify POE management functionality:

```bash
./test-poe-toggle.sh <switch-hostname>

# Example:
export NETGEAR_SWITCHES="tswitch16:password"
./test-poe-toggle.sh tswitch16
```

The test will:
1. Capture current POE port states
2. Toggle all ports (enabled→disabled, disabled→enabled)
3. Verify the changes took effect
4. Restore original port states
5. Verify restoration succeeded

## Debug Mode

All programs support debug mode with `--debug` or `-d` flags to see:
- HTTP requests and responses
- Authentication details
- Token validation and caching
- Internal library operations
- Detailed error information

Example:
```bash
./bin/poe-management --debug 192.168.1.10 status
```

## Development

### Prerequisites

- Go 1.23 or later
- Access to a Netgear managed switch (GS30x or GS316 series)

### Building from Source

```bash
# Install dependencies
go mod download

# Build all binaries
make build

# Or build individually
go build -o bin/poe-status cmd/poe-status/main.go
go build -o bin/poe-status-simple cmd/poe-status-simple/main.go
go build -o bin/poe-management cmd/poe-management/main.go
```

### Clean Build

```bash
make clean   # Remove binaries and clean Go cache
make build   # Rebuild everything
```

## Supported Switches

This tool works with Netgear managed switches that support the web API, including:
- GS305EP, GS305EPP
- GS308EP, GS308EPP
- GS316EP, GS316EPP

For more details on switch models and authentication, see [docs/login.md](docs/login.md).

## Library

These examples use the [go-netgear](https://github.com/gherlein/go-netgear) library, which provides a Go interface to Netgear switch management features.

## License

See LICENSE file for details.

## Acknowledgments

Special thanks to [Martin W. Kirst](https://github.com/nitram509) for creating [ntgrrc](https://github.com/nitram509/ntgrrc), which served as the foundation and inspiration for this Go implementation.
