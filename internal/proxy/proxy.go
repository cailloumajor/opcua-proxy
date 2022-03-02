package proxy

import (
	"net/http"

	"github.com/gopcua/opcua"
)

//go:generate moq -out proxy_mocks_test.go . OPCUAClient

// OPCUAClient is a consumer contract modelling an OPC-UA client.
type OPCUAClient interface {
	State() opcua.ConnState
}

func methodHandler(m string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != m {
			w.Header().Set("Allow", m)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Proxy handles requests for the service.
type Proxy struct {
	opc OPCUAClient

	http.Handler
}

// NewProxy creates and returns a ready to use proxy.
func NewProxy(opc OPCUAClient) *Proxy {
	p := &Proxy{
		opc: opc,
	}

	mux := http.NewServeMux()
	mux.Handle("/health", methodHandler(http.MethodGet, http.HandlerFunc(p.handleHealth)))
	p.Handler = mux

	return p
}

func (p *Proxy) handleHealth(w http.ResponseWriter, r *http.Request) {
	nok := func(msg string) {
		http.Error(w, msg, http.StatusInternalServerError)
	}

	switch {
	case p.opc.State() != opcua.Connected:
		nok("OPC-UA client not connected")
	}
}
