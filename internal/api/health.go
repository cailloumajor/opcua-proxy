package api

import (
	"context"
	"fmt"
	"net/http"
)

//go:generate moq -out health_mocks_test.go . Healther

// Healther is a consumer contract modelling a health status provider.
type Healther interface {
	Health(ctx context.Context) (bool, string)
}

type monitored struct {
	name   string
	target Healther
}

// HealthService represents a health checker API service.
//
// The zero value is ready to use.
type HealthService struct {
	checked []monitored
}

// Register adds a named target to be monitored.
func (c *HealthService) Register(name string, target Healther) {
	c.checked = append(c.checked, monitored{name, target})
}

// ServeHTTP implements http.Handler.
func (c *HealthService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, m := range c.checked {
		if ok, s := m.target.Health(r.Context()); !ok {
			title := fmt.Sprintf("component `%v` is unhealthy", m.name)
			ReplyProblemDetails(w, "unhealthy", title, s, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
