// This file contains implementations wrapping to top-level functions of gocpua library.

package opcua

import (
	"context"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

// DefaultSecurityExtDeps represents the default SecurityExtDeps implementation.
type DefaultSecurityExtDeps struct{}

// AuthUsername implements SecurityExtDeps.
func (d DefaultSecurityExtDeps) AuthUsername(user, pass string) opcua.Option {
	return opcua.AuthUsername(user, pass)
}

// CertificateFile implements SecurityExtDeps.
func (d DefaultSecurityExtDeps) CertificateFile(filename string) opcua.Option {
	return opcua.CertificateFile(filename)
}

// PrivateKeyFile implements SecurityExtDeps.
func (d DefaultSecurityExtDeps) PrivateKeyFile(filename string) opcua.Option {
	return opcua.PrivateKeyFile(filename)
}

// SecurityFromEndpoint implements SecurityExtDeps.
func (d DefaultSecurityExtDeps) SecurityFromEndpoint(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option {
	return opcua.SecurityFromEndpoint(ep, authType)
}

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
func (DefaultClientExtDeps) NewClient(endpoint string, opts ...opcua.Option) RawClientProvider {
	return opcua.NewClient(endpoint, opts...)
}
