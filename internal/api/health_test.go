package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/cailloumajor/opcua-proxy/internal/api"
)

func TestHealthServiceHandler(t *testing.T) {
	cases := []struct {
		name             string
		unhealthy        bool
		expectStatusCode int
	}{
		{
			name:             "Unhealthy",
			unhealthy:        true,
			expectStatusCode: http.StatusInternalServerError,
		},
		{
			name:             "OK",
			unhealthy:        false,
			expectStatusCode: http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedFirstHealther := &HealtherMock{
				HealthFunc: func(ctx context.Context) (bool, string) {
					return true, "well"
				},
			}
			mockedSecondHealther := &HealtherMock{
				HealthFunc: func(ctx context.Context) (bool, string) {
					if tc.unhealthy {
						return false, "sick"
					}
					return true, "good"
				},
			}
			s := &HealthService{}
			s.Register("first", mockedFirstHealther)
			s.Register("second", mockedSecondHealther)

			req := httptest.NewRequest("", "/", nil)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)

			resp := w.Result()

			if got, want := resp.StatusCode, tc.expectStatusCode; got != want {
				t.Errorf("status code: want %d, got %d", want, got)
			}
		})
	}
}
