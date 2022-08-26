package api_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/cailloumajor/opcua-proxy/internal/api"
	"github.com/cailloumajor/opcua-proxy/internal/lineprotocol"
	"github.com/cailloumajor/opcua-proxy/internal/opcua"
	"github.com/cailloumajor/opcua-proxy/internal/testutils"
	"github.com/gopcua/opcua/ua"
)

func TestInfluxDbMetricsServiceServeHTTP(t *testing.T) {
	cases := []struct {
		name              string
		querystring       string
		readError         bool
		buildError        bool
		expectStatusCode  int
		expectContentType string
		expectBody        string
		ignoreBody        bool
	}{
		{
			name:              "InvalidQueryString",
			querystring:       "measurement=meas&tag1=val1&othertag=otherval;",
			readError:         false,
			buildError:        false,
			expectStatusCode:  http.StatusInternalServerError,
			expectContentType: "application/problem+json",
			expectBody:        "",
			ignoreBody:        true,
		},
		{
			name:              "MissingMeasurement",
			querystring:       "tag1=val1&othertag=otherval",
			readError:         false,
			buildError:        false,
			expectStatusCode:  http.StatusBadRequest,
			expectContentType: "application/problem+json",
			expectBody:        "",
			ignoreBody:        true,
		},
		{
			name:              "ReadError",
			querystring:       "measurement=meas&tag1=val1&othertag=otherval",
			readError:         true,
			buildError:        false,
			expectStatusCode:  http.StatusInternalServerError,
			expectContentType: "application/problem+json",
			expectBody:        "",
			ignoreBody:        true,
		},
		{
			name:              "BuildError",
			querystring:       "measurement=meas&tag1=val1&othertag=otherval",
			readError:         false,
			buildError:        true,
			expectStatusCode:  http.StatusInternalServerError,
			expectContentType: "application/problem+json",
			expectBody:        "",
			ignoreBody:        true,
		},
		{
			name:              "Success",
			querystring:       "measurement=meas&tag1=val1&othertag=otherval",
			readError:         false,
			buildError:        false,
			expectStatusCode:  http.StatusOK,
			expectContentType: "text/plain",
			expectBody:        "meas map[othertag:otherval tag1:val1] map[field1: field2:val] 2006-01-02 15:04:05 -0700 -0700",
			ignoreBody:        false,
		},
	}

	ts, err := time.Parse(time.Layout, time.Layout)
	if err != nil {
		t.Fatalf("error parsing time: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockedOpcUaReader := &OpcUaReaderMock{
				ReadFunc: func(ctx context.Context) (*opcua.ReadValues, error) {
					if tc.readError {
						return nil, testutils.ErrTesting
					}
					return &opcua.ReadValues{
						Timestamp: ts,
						Values: map[string]*ua.Variant{
							"field1": ua.MustVariant(37.2),
							"field2": ua.MustVariant("val"),
						},
					}, nil
				},
			}
			mockedLineProtocolBuilder := &LineProtocolBuilderMock{
				BuildFunc: func(w io.Writer, measurement string, tags map[string]string, fields map[string]lineprotocol.VariantProvider, ts time.Time) error {
					if tc.buildError {
						return testutils.ErrTesting
					}
					if _, err := fmt.Fprintf(w, "%v %v %v %v", measurement, tags, fields, ts); err != nil {
						t.Fatalf("error writing to ResponseWriter: %v", err)
					}
					return nil
				},
			}
			s := NewInfluxDbMetricsService(mockedOpcUaReader, mockedLineProtocolBuilder)

			req := httptest.NewRequest("", "/?"+tc.querystring, nil)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			if got, want := resp.StatusCode, tc.expectStatusCode; got != want {
				t.Errorf("status code: want %d, got %d", want, got)
			}
			if got, want := resp.Header.Get("content-type"), tc.expectContentType; !strings.HasPrefix(got, want) {
				t.Errorf("content type: want %q, got %q", want, got)
			}
			if got, want := string(body), tc.expectBody; !tc.ignoreBody && got != want {
				t.Errorf("body: got %q, want %q", got, want)
			}
		})
	}
}
