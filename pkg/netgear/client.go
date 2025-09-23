// Package netgear provides a client wrapper for interacting with Netgear switches
package netgear

import (
	"context"
	"fmt"

	gonetgear "github.com/gherlein/go-netgear"
)

// Client represents a Netgear switch client
type Client struct {
	host        string
	model       string
	authenticated bool
	options     *ClientOptions
	globalOpts  *gonetgear.GlobalOptions
}

// ClientOptions contains configuration options for the client
type ClientOptions struct {
	Verbose      bool
	TokenManager TokenManager
}

// ClientOption is a function type for configuring the client
type ClientOption func(*ClientOptions)

// WithVerbose enables verbose logging
func WithVerbose(verbose bool) ClientOption {
	return func(opts *ClientOptions) {
		opts.Verbose = verbose
	}
}

// WithTokenManager sets the token manager
func WithTokenManager(tm TokenManager) ClientOption {
	return func(opts *ClientOptions) {
		opts.TokenManager = tm
	}
}

// NewClient creates a new Netgear client
func NewClient(host string, options ...ClientOption) (*Client, error) {
	opts := &ClientOptions{}
	for _, opt := range options {
		opt(opts)
	}

	globalOpts := &gonetgear.GlobalOptions{
		Verbose:      opts.Verbose,
		OutputFormat: gonetgear.JsonFormat, // Use JSON format for structured output
	}

	client := &Client{
		host:       host,
		model:      "Unknown",
		options:    opts,
		globalOpts: globalOpts,
	}

	return client, nil
}

// IsAuthenticated returns whether the client is authenticated
func (c *Client) IsAuthenticated() bool {
	return c.authenticated
}

// GetModel returns the switch model
func (c *Client) GetModel() string {
	if c.model == "Unknown" && c.host != "" {
		// Try to detect model
		model, err := gonetgear.DetectNetgearModel(c.globalOpts, c.host)
		if err == nil {
			c.model = string(model)
		}
	}
	return c.model
}

// Login authenticates with the switch using a password
func (c *Client) Login(ctx context.Context, password string) error {
	// Set up login command
	loginCmd := &gonetgear.LoginCommand{
		Address:  c.host,
		Password: password,
	}

	err := loginCmd.Run(c.globalOpts)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	c.authenticated = true
	c.GetModel() // Update model after successful login
	return nil
}

// LoginAuto attempts automatic authentication using environment variables
func (c *Client) LoginAuto(ctx context.Context) error {
	// Set up login command - for auto login, password is empty and will be read from environment
	loginCmd := &gonetgear.LoginCommand{
		Address: c.host,
		// Password is empty - the command will try to read from environment variables
	}

	err := loginCmd.Run(c.globalOpts)
	if err != nil {
		return fmt.Errorf("auto-login failed: %w", err)
	}

	c.authenticated = true
	c.GetModel() // Update model after successful login
	return nil
}

// POE returns the POE interface for this client
func (c *Client) POE() *POEInterface {
	return &POEInterface{client: c}
}

// POEInterface provides POE-related operations
type POEInterface struct {
	client *Client
}

// POEPortStatus represents the status of a POE port
type POEPortStatus struct {
	PortID       int     `json:"port_id"`
	PortName     string  `json:"port_name"`
	Status       string  `json:"status"`
	PowerClass   string  `json:"power_class"`
	VoltageV     float64 `json:"voltage_v"`
	CurrentMA    float64 `json:"current_ma"`
	PowerW       float64 `json:"power_w"`
	TemperatureC float64 `json:"temperature_c"`
	ErrorStatus  string  `json:"error_status"`
}

// POEPortSetting represents POE port configuration
type POEPortSetting struct {
	PortID       int     `json:"port_id"`
	Enabled      bool    `json:"enabled"`
	Mode         string  `json:"mode"`
	Priority     string  `json:"priority"`
	PowerLimitW  float64 `json:"power_limit_w"`
}

// GetStatus retrieves POE status for all ports
func (p *POEInterface) GetStatus(ctx context.Context) ([]POEPortStatus, error) {
	cmd := &gonetgear.PoeStatusCommand{
		Address: p.client.host,
	}

	err := cmd.Run(p.client.globalOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get POE status: %w", err)
	}

	// NOTE: The go-netgear package outputs directly to stdout with properly
	// formatted data. The real POE status for all available ports has already
	// been displayed. We return an empty slice since the caller should rely
	// on the stdout output rather than the return value.
	return []POEPortStatus{}, nil
}

// GetSettings retrieves POE settings for all ports
func (p *POEInterface) GetSettings(ctx context.Context) ([]POEPortSetting, error) {
	cmd := &gonetgear.PoeShowSettingsCommand{
		Address: p.client.host,
	}

	// We need to call the command to get the real data, but the go-netgear
	// library only prints to stdout and doesn't return structured data.
	// The internal functions that parse the HTML are not exported.
	// For now, we call the real command (which will output to stdout)
	// and acknowledge this limitation.
	err := cmd.Run(p.client.globalOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get POE settings: %w", err)
	}

	// NOTE: This is a limitation of the go-netgear package design.
	// The actual POE settings data is parsed internally but only printed to stdout.
	// To get the real data, we would need access to the internal functions
	// like findPoePortConfInHtml() which are not exported.
	// The real command has already executed above and output the actual data.

	// Since the real command outputs to stdout but we need to return structured data
	// for the example applications to work, we return empty slice to indicate
	// that the data should be read from stdout instead.
	return []POEPortSetting{}, nil
}

// EnablePort enables POE on the specified port
func (p *POEInterface) EnablePort(ctx context.Context, port int) error {
	cmd := &gonetgear.PoeSetConfigCommand{
		Address: p.client.host,
		Ports:   []int{port},
		PortPwr: "enable",
	}

	err := cmd.Run(p.client.globalOpts)
	if err != nil {
		return fmt.Errorf("failed to enable port %d: %w", port, err)
	}

	return nil
}

// DisablePort disables POE on the specified port
func (p *POEInterface) DisablePort(ctx context.Context, port int) error {
	cmd := &gonetgear.PoeSetConfigCommand{
		Address: p.client.host,
		Ports:   []int{port},
		PortPwr: "disable",
	}

	err := cmd.Run(p.client.globalOpts)
	if err != nil {
		return fmt.Errorf("failed to disable port %d: %w", port, err)
	}

	return nil
}

// CyclePower cycles power on the specified ports
func (p *POEInterface) CyclePower(ctx context.Context, ports ...int) error {
	cmd := &gonetgear.PoeCyclePowerCommand{
		Address: p.client.host,
		Ports:   ports,
	}

	err := cmd.Run(p.client.globalOpts)
	if err != nil {
		return fmt.Errorf("failed to cycle power on ports %v: %w", ports, err)
	}

	return nil
}