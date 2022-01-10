package opcua

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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
	Close() error
}

// NodeMonitor models an OPC-UA node monitor
type NodeMonitor interface {
	ChanSubscribe(context.Context, *opcua.SubscriptionParameters, chan<- *monitor.DataChangeMessage, ...string) (*monitor.Subscription, error)
	SetErrorHandler(cb monitor.ErrHandler)
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

// Monitor is an OPC-UA node monitor
type Monitor struct {
	client      Client
	nodeMonitor NodeMonitor
}

// NewMonitor creates an OPC-UA node monitor
func NewMonitor(ctx context.Context, cfg *Config, deps NewMonitorDeps, logger log.Logger) (*Monitor, error) {
	eps, err := deps.GetEndpoints(ctx, cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("error getting endpoints: %w", err)
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
		return nil, fmt.Errorf("failed to select an endpoint")
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
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	nm, err := deps.NewNodeMonitor(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create a node monitor: %w", err)
	}

	nm.SetErrorHandler(func(c *opcua.Client, s *monitor.Subscription, e error) {
		level.Info(logger).Log("from", "subscription", "sub_id", s.SubscriptionID(), "err", e)
	})

	return &Monitor{client: c, nodeMonitor: nm}, nil
}

// CloseClient closes the OPC-UA client
func (m *Monitor) CloseClient() error {
	return m.client.Close()
}
