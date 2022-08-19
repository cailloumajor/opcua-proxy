// This file contains implementations wrapping to top-level functions of gocpua library.

package opcua

import (
	"context"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

// DefaultClientExtDeps represents the default ClientExtDeps implementation.
type DefaultClientExtDeps struct{}

// GetEndpoints implements ClientExtDeps.
func (DefaultClientExtDeps) GetEndpoints(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
	return opcua.GetEndpoints(ctx, endpoint, opts...)
}

// SelectEndpoint implements ClientExtDeps.
func (DefaultClientExtDeps) SelectEndpoint(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
	return opcua.SelectEndpoint(endpoints, policy, mode)
}

// NewClient implements ClientExtDeps.
func (DefaultClientExtDeps) NewClient(endpoint string, opts ...opcua.Option) *opcua.Client {
	return opcua.NewClient(endpoint, opts...)
}

// AuthUsername implements ClientExtDeps.
func (d DefaultClientExtDeps) AuthUsername(user, pass string) opcua.Option {
	return opcua.AuthUsername(user, pass)
}

// CertificateFile implements ClientExtDeps.
func (d DefaultClientExtDeps) CertificateFile(filename string) opcua.Option {
	return opcua.CertificateFile(filename)
}

// PrivateKeyFile implements ClientExtDeps.
func (d DefaultClientExtDeps) PrivateKeyFile(filename string) opcua.Option {
	return opcua.PrivateKeyFile(filename)
}

// SecurityFromEndpoint implements ClientExtDeps.
func (d DefaultClientExtDeps) SecurityFromEndpoint(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option {
	return opcua.SecurityFromEndpoint(ep, authType)
}
