# Login Process Documentation

## Overview

The go-netgear library implements authentication for Netgear switches using a session-based token system. The login process varies slightly based on the switch model series (30x vs 316).

## Login Flow

### Step 1: Model Detection
The library first detects the switch model to determine which authentication method to use (see `models.DetectNetgearModel()`).

### Step 2: Obtain Seed Value
Before authentication, the library fetches a random seed value from the switch:
- **GS30x models**: GET request to `http://{host}/login.cgi`
- **GS316 models**: GET request to `http://{host}/wmi/login`

The seed value is extracted from an HTML input field with `id="rand"`.

### Step 3: Encrypt Password
The password is encrypted using a custom algorithm that:
1. Merges the password and seed value character-by-character
2. Computes an MD5 hash of the merged string

See `encryptPassword()` in `/home/developer/go/pkg/mod/github.com/gherlein/go-netgear@v0.0.1/internal/client/login.go:202`

### Step 4: Authenticate
The encrypted password is submitted via POST request:
- **GS30x models**: POST to `http://{host}/login.cgi` with form data `password={encrypted_pwd}`
- **GS316 models**: POST to `http://{host}/redirect.html` with form data `LoginPassword={encrypted_pwd}`

### Step 5: Extract Token
Upon successful authentication, a session token is returned:
- **GS30x models**: Token is returned in `Set-Cookie` header as `SID` cookie
- **GS316 models**: Token is returned as a hidden form field named `Gambit` in the response HTML

## Token Types

### GS30x Series (GS305EP, GS305EPP, GS308EP, GS308EPP)
- **Token Type**: SID (Session ID) cookie
- **Format**: Alphanumeric string
- **Usage**: Sent as HTTP Cookie header: `Cookie: SID={token}`

### GS316 Series (GS316EP, GS316EPP)
- **Token Type**: Gambit token
- **Format**: Alphanumeric string
- **Usage**:
  - Sent as HTTP Cookie header: `Cookie: gambitCookie={token}`
  - Appended to URL query parameters: `?Gambit={token}`

## Token Lifetime

**Important**: The library does **not** enforce any expiration on cached tokens. Tokens are stored persistently and used until they fail authentication.

### Token Validation
- No explicit expiration timestamp is stored with the token
- Token validity is determined by the switch itself
- When a cached token is no longer valid, the switch responds with content indicating login is required (see `CheckIsLoginRequired()` in `/home/developer/go/pkg/mod/github.com/gherlein/go-netgear@v0.0.1/internal/common/http.go:88`)
- Detection criteria: response contains `/login.cgi`, `/wmi/login`, or `/redirect.html`

### Expected Behavior
Based on typical session management:
- Tokens likely expire after a period of inactivity (exact duration is switch-dependent)
- Tokens may become invalid if the switch is rebooted
- Multiple concurrent sessions may or may not be supported (switch-dependent)

## Token Caching

### Storage Location
Tokens are cached in the filesystem at:
```
{TokenDir}/.config/ntgrrc/token-{hash}
```

Where:
- `{TokenDir}` defaults to `os.TempDir()` if not specified
- `{hash}` is the Adler32 hash of the switch hostname (8 hex characters)

### File Contents
Token files contain:
```
{model}:{token}
```

Example:
```
GS305EP:abc123def456
```

### Permissions
- Token directory: `0700` (owner read/write/execute only)
- Token files: `0644` (owner read/write, group/others read)

## Caching Options

### Option 1: Default Temporary Storage
**Implementation**: Do not set `GlobalOptions.TokenDir`

```go
globalOpts := &go_netgear.GlobalOptions{
    // TokenDir not set - defaults to os.TempDir()
}
```

**Characteristics**:
- Tokens stored in system temp directory
- May be cleaned up on system reboot
- Suitable for short-lived scripts or testing

**Location examples**:
- Linux: `/tmp/.config/ntgrrc/`
- macOS: `/var/folders/.../.config/ntgrrc/`
- Windows: `%TEMP%\.config\ntgrrc\`

### Option 2: Persistent Storage
**Implementation**: Set `GlobalOptions.TokenDir` to a persistent location

```go
globalOpts := &go_netgear.GlobalOptions{
    TokenDir: "/home/user/.local/share/netgear",
}
```

**Characteristics**:
- Tokens survive system reboots
- Suitable for daemons, long-running services, or frequent CLI usage
- User controls the directory location

**Recommended locations**:
- Linux: `$HOME/.config/netgear` or `$HOME/.local/share/netgear`
- macOS: `$HOME/Library/Application Support/netgear`
- Windows: `%APPDATA%\netgear`

### Option 3: In-Memory (No Caching)
**Implementation**: Set `GlobalOptions.Token` and `GlobalOptions.Model` directly

```go
globalOpts := &go_netgear.GlobalOptions{
    Token: "abc123def456",
    Model: go_netgear.GS305EP,
}
```

**Characteristics**:
- No disk storage
- Token must be provided for every invocation
- Suitable for stateless services or when token is managed externally
- Application must handle token acquisition and storage

### Option 4: Custom Cache Management
**Implementation**: Use custom TokenDir and implement your own cache invalidation

```go
globalOpts := &go_netgear.GlobalOptions{
    TokenDir: "/custom/path",
}

// Manually delete token when needed
tokenHash := fmt.Sprintf("%x", adler32.Checksum([]byte(hostname)))
tokenPath := filepath.Join(customDir, ".config", "ntgrrc", "token-"+tokenHash)
os.Remove(tokenPath)
```

**Characteristics**:
- Full control over token lifecycle
- Can implement custom expiration policies
- Can integrate with external secret management systems

## Security Considerations

1. **Token File Permissions**: While token files have restrictive permissions (0644), the directory is only protected at 0700. Ensure the parent `TokenDir` is in a secure location.

2. **No Encryption**: Tokens are stored in plain text on disk. Consider:
   - Using encrypted filesystems for sensitive environments
   - Implementing your own encryption layer if storing in shared locations
   - Using Option 3 (in-memory) for highest security

3. **Token Rotation**: The library does not automatically rotate tokens. Old tokens remain cached until authentication fails.

4. **Multi-User Systems**: On shared systems, use user-specific directories (e.g., within `$HOME`) to prevent token access by other users.

## Code References

- Login implementation: `/home/developer/go/pkg/mod/github.com/gherlein/go-netgear@v0.0.1/internal/client/login.go`
- Token storage: `/home/developer/go/pkg/mod/github.com/gherlein/go-netgear@v0.0.1/internal/client/token.go:15`
- Token retrieval: `/home/developer/go/pkg/mod/github.com/gherlein/go-netgear@v0.0.1/internal/common/token.go:17`
- HTTP usage: `/home/developer/go/pkg/mod/github.com/gherlein/go-netgear@v0.0.1/internal/common/http.go`
- Type definitions: `/home/developer/go/pkg/mod/github.com/gherlein/go-netgear@v0.0.1/internal/types/types.go`

## Example Workflows

### Example 1: Simple CLI with Default Caching
```go
func main() {
    opts := &go_netgear.GlobalOptions{
        OutputFormat: go_netgear.JsonFormat,
    }

    // Login once - token cached in temp directory
    login := &go_netgear.LoginCommand{
        Address:  "192.168.1.10",
        Password: "admin123",
    }
    login.Run(opts)

    // Subsequent commands use cached token automatically
    status := &go_netgear.PoeStatusCommand{Address: "192.168.1.10"}
    status.Run(opts)
}
```

### Example 2: Service with Persistent Tokens
```go
func main() {
    homeDir, _ := os.UserHomeDir()

    opts := &go_netgear.GlobalOptions{
        TokenDir:     filepath.Join(homeDir, ".config", "netgear"),
        OutputFormat: go_netgear.JsonFormat,
    }

    // Token persists across service restarts
    login := &go_netgear.LoginCommand{
        Address:  "192.168.1.10",
        Password: os.Getenv("NETGEAR_PASSWORD"),
    }

    if err := login.Run(opts); err != nil {
        log.Fatal(err)
    }
}
```

### Example 3: Stateless with External Token Management
```go
func main() {
    // Assume token retrieved from secrets manager
    token := getTokenFromSecretManager("netgear-switch-01")

    opts := &go_netgear.GlobalOptions{
        Model:        go_netgear.GS305EP,
        Token:        token,
        OutputFormat: go_netgear.JsonFormat,
    }

    // No caching - uses provided token directly
    status := &go_netgear.PoeStatusCommand{Address: "192.168.1.10"}
    status.Run(opts)
}
```
