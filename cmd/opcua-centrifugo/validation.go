package main

import (
	"fmt"
	"net/url"
)

// ValidateCentrifugoAddress validates an URL to use as Centrifugo API address.
func ValidateCentrifugoAddress(addr string) error {
	url, err := url.ParseRequestURI(addr)
	if err != nil {
		return err
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return fmt.Errorf("bad protocol scheme: %s", url.Scheme)
	}

	return nil
}
