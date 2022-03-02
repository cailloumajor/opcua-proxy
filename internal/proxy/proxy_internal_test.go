package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

func TestMethodHandler(t *testing.T) {
	cases := []struct {
		name              string
		gotMethod         string
		expectStatusCode  int
		expectAllowHeader string
	}{
		{
			name:              "MethodNotAllowed",
			gotMethod:         http.MethodHead,
			expectStatusCode:  http.StatusMethodNotAllowed,
			expectAllowHeader: http.MethodGet,
		},
		{
			name:              "AllowedMethod",
			gotMethod:         http.MethodGet,
			expectStatusCode:  http.StatusOK,
			expectAllowHeader: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := methodHandler(http.MethodGet, http.HandlerFunc(okHandler))
			req := httptest.NewRequest(tc.gotMethod, "/", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			resp := rec.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("error reading body: %v", err)
			}

			if got, want := resp.StatusCode, tc.expectStatusCode; got != want {
				t.Errorf("status code: want %d, got %d", want, got)
			}
			if got, want := resp.Header.Get("Allow"), tc.expectAllowHeader; got != want {
				t.Errorf("\"Allow\" header: want %q, got %q", want, got)
			}
			if got, want := string(body), http.StatusText(tc.expectStatusCode)+"\n"; got != want {
				t.Errorf("body: want %q, got %q", want, got)
			}
		})
	}
}
