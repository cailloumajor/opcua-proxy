package opcua

import (
	"context"
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
)

//go:generate moq -out opcua_mocks_test.go . Client NodeMonitor NewMonitorDeps

// Config holds the OPC-UA part of the configuration
type Config struct {
	ServerURL string
	User      string `envconfig:"optional"`
	Password  string `envconfig:"optional"`
	CertFile  string `envconfig:"optional"`
	KeyFile   string `envconfig:"optional"`
}

// Client models an OPC-UA client
type Client interface {
	Connect(context.Context) (err error)
}

// NodeMonitor models an OPC-UA node monitor
type NodeMonitor interface {
	ChanSubscribe(context.Context, *opcua.SubscriptionParameters, chan<- *monitor.DataChangeMessage, ...string) (*monitor.Subscription, error)
}

// NewMonitorDeps models the dependencies of NewMonitor
type NewMonitorDeps interface {
	GetEndpoints(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error)
	SelectEndpoint(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription
	AuthUsername(user, pass string) opcua.Option
	CertificateFile(filename string) opcua.Option
	PrivateKeyFile(filename string) opcua.Option
	SecurityFromEndpoint(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option
	NewClient(endpoint string, opts ...opcua.Option) Client
	NewNodeMonitor(client Client) (NodeMonitor, error)
}

// NewMonitor creates an OPC-UA node monitor
func NewMonitor(ctx context.Context, cfg *Config, deps NewMonitorDeps) (NodeMonitor, error) {
	eps, err := deps.GetEndpoints(ctx, cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("NewMonitor: error getting endpoints: %w", err)
	}

	var opts []opcua.Option

	var p string
	var msm ua.MessageSecurityMode
	if cfg.CertFile != "" || cfg.KeyFile != "" {
		p = "Basic256Sha256"
		msm = ua.MessageSecurityModeSignAndEncrypt
		opts = append(opts, deps.CertificateFile(cfg.CertFile), deps.PrivateKeyFile(cfg.KeyFile))
	} else {
		p = "None"
		msm = ua.MessageSecurityModeNone
	}

	ep := deps.SelectEndpoint(eps, p, msm)
	if ep == nil {
		return nil, fmt.Errorf("NewMonitor: failed to select an endpoint")
	}

	var utt ua.UserTokenType
	if cfg.User == "" {
		utt = ua.UserTokenTypeAnonymous
	} else {
		utt = ua.UserTokenTypeUserName
		opts = append(opts, deps.AuthUsername(cfg.User, cfg.Password))
	}
	opts = append(opts, deps.SecurityFromEndpoint(ep, utt))

	c := deps.NewClient(ep.EndpointURL, opts...)

	if err := c.Connect(ctx); err != nil {
		return nil, fmt.Errorf("NewMonitor: failed to connect: %w", err)
	}

	m, err := deps.NewNodeMonitor(c)
	if err != nil {
		return nil, fmt.Errorf("NewMonitor: failed to create a node monitor: %w", err)
	}

	return m, nil
}
