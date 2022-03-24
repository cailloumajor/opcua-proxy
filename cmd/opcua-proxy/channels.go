package main

import (
	"context"
	"fmt"

	"github.com/cailloumajor/opcua-proxy/internal/centrifugo"
	"github.com/centrifugal/gocent/v3"
)

//go:generate moq -out channels_mocks_test.go . CentrifugoChannels

// CentrifugoChannels is a consumer contract modelling a Centrifugo channels query.
type CentrifugoChannels interface {
	Channels(ctx context.Context, opts ...gocent.ChannelsOption) (gocent.ChannelsResult, error)
}

// Channels returns the the Centrifugo channels in the given namespace.
func Channels(ctx context.Context, ch CentrifugoChannels, ns string) ([]string, error) {
	pat := fmt.Sprint(ns, centrifugo.NsSeparator, "*")
	res, err := ch.Channels(ctx, gocent.WithPattern(pat))
	if err != nil {
		return nil, fmt.Errorf("error querying channels: %w", err)
	}

	chans := make([]string, 0, len(res.Channels))

	for chName := range res.Channels {
		chans = append(chans, chName)
	}

	return chans, nil
}
