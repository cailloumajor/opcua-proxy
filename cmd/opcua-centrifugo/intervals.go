package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cailloumajor/opcua-centrifugo/internal/centrifugo"
	"github.com/centrifugal/gocent/v3"
)

//go:generate moq -out intervals_mocks_test.go . CentrifugoChannels

// CentrifugoChannels is a consumer contract modelling a Centrifugo channels query.
type CentrifugoChannels interface {
	Channels(ctx context.Context, opts ...gocent.ChannelsOption) (gocent.ChannelsResult, error)
}

// ChannelIntervals returns the intervals of the Centrifugo channels in the given namespace.
func ChannelIntervals(ctx context.Context, ch CentrifugoChannels, ns string) ([]time.Duration, error) {
	pat := fmt.Sprint(ns, centrifugo.NsSeparator, "*")
	res, err := ch.Channels(ctx, gocent.WithPattern(pat))
	if err != nil {
		return nil, fmt.Errorf("error querying channels: %w", err)
	}

	var durations []time.Duration

	for chName := range res.Channels {
		c, err := centrifugo.ParseChannel(chName, ns)
		if err != nil {
			continue
		}
		durations = append(durations, c.Interval())
	}

	return durations, nil
}
