package centrifugo

import (
	"context"
	"fmt"

	"github.com/centrifugal/gocent/v3"
)

//go:generate moq -out client_mocks_test.go . ClientProvider

// ClientProvider is a consumer contract modelling a Centrifugo client.
type ClientProvider interface {
	Channels(ctx context.Context, opts ...gocent.ChannelsOption) (gocent.ChannelsResult, error)
	Info(ctx context.Context) (gocent.InfoResult, error)
	Publish(ctx context.Context, channel string, data []byte, opts ...gocent.PublishOption) (gocent.PublishResult, error)
}

// Client wraps a Centrifugo client.
type Client struct {
	ClientProvider
}

// NewClient creates a wrapped Centrifugo client.
func NewClient(addr, key string) *Client {
	cfg := gocent.Config{
		Addr: addr,
		Key:  key,
	}
	return &Client{gocent.New(cfg)}
}

// Channels returns the the Centrifugo channels in the given namespace.
func (c *Client) Channels(ctx context.Context, ns string) ([]string, error) {
	pat := fmt.Sprint(ns, NsSeparator, "*")
	res, err := c.ClientProvider.Channels(ctx, gocent.WithPattern(pat))
	if err != nil {
		return nil, fmt.Errorf("error querying channels: %w", err)
	}

	chans := make([]string, 0, len(res.Channels))

	for chName := range res.Channels {
		chans = append(chans, chName)
	}

	return chans, nil
}

// Health checks Centrifugo health.
func (c *Client) Health(ctx context.Context) (bool, string) {
	if _, err := c.Info(ctx); err != nil {
		return false, err.Error()
	}

	return true, ""
}
