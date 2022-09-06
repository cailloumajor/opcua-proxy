// This file contains implementations wrapping to top-level functions.

package opcua

import (
	"context"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type defaultClientExtDeps struct{}

func (defaultClientExtDeps) GetEndpoints(ctx context.Context, endpoint string, opts ...opcua.Option) ([]*ua.EndpointDescription, error) {
	return opcua.GetEndpoints(ctx, endpoint, opts...)
}

func (defaultClientExtDeps) SelectEndpoint(endpoints []*ua.EndpointDescription, policy string, mode ua.MessageSecurityMode) *ua.EndpointDescription {
	return opcua.SelectEndpoint(endpoints, policy, mode)
}

func (defaultClientExtDeps) NewClient(endpoint string, opts ...opcua.Option) *opcua.Client {
	return opcua.NewClient(endpoint, opts...)
}

func (d defaultClientExtDeps) AuthUsername(user, pass string) opcua.Option {
	return opcua.AuthUsername(user, pass)
}

func (d defaultClientExtDeps) CertificateFile(filename string) opcua.Option {
	return opcua.CertificateFile(filename)
}

func (d defaultClientExtDeps) PrivateKeyFile(filename string) opcua.Option {
	return opcua.PrivateKeyFile(filename)
}

func (d defaultClientExtDeps) SecurityFromEndpoint(ep *ua.EndpointDescription, authType ua.UserTokenType) opcua.Option {
	return opcua.SecurityFromEndpoint(ep, authType)
}
