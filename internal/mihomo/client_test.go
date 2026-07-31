package mihomo_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

func TestClientEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got, want := request.Header.Get("Authorization"), "Bearer test-secret"; got != want {
			t.Errorf("Authorization header = %q, want %q", got, want)
		}

		switch request.URL.EscapedPath() {
		case "/api/version":
			_, err := writer.Write([]byte(`{"meta":true,"version":"v1.19.0"}`))
			if err != nil {
				t.Errorf("write version response: %v", err)
			}
		case "/api/proxies":
			writeFixture(t, writer, "proxies.json")
		case "/api/providers/proxies":
			writeFixture(t, writer, "providers.json")
		case "/api/providers/proxies/example-provider/healthcheck":
			_, err := writer.Write([]byte(`{}`))
			if err != nil {
				t.Errorf("write healthcheck response: %v", err)
			}
		case "/api/proxies/proxy%2Fblue/delay":
			got := request.URL.Query()
			want := url.Values{
				"timeout": []string{"5000"},
				"url":     []string{"https://test.example/generate_204"},
			}
			if !valuesEqual(got, want) {
				t.Errorf("delay query = %q, want %q", got.Encode(), want.Encode())
			}
			_, err := writer.Write([]byte(`{"delay":123}`))
			if err != nil {
				t.Errorf("write delay response: %v", err)
			}
		case "/api/connections":
			writeFixture(t, writer, "connections.json")
		default:
			t.Errorf("request path = %q, want a Mihomo endpoint", request.URL.EscapedPath())
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	client, err := mihomo.New(server.URL+"/api", "test-secret", time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	tests := map[string]func(context.Context) error{
		"version": func(ctx context.Context) error {
			version, versionErr := client.Version(ctx)
			if versionErr == nil && (!version.Meta || version.Version != "v1.19.0") {
				t.Errorf("version = %#v, want meta=true and version v1.19.0", version)
			}

			return versionErr
		},
		"proxies": func(ctx context.Context) error {
			proxies, proxiesErr := client.Proxies(ctx)
			if proxiesErr == nil && proxies["example-ss"].ProviderName != "example-provider" {
				t.Errorf("proxy provider = %q, want %q", proxies["example-ss"].ProviderName, "example-provider")
			}

			return proxiesErr
		},
		"providers": func(ctx context.Context) error {
			providers, providersErr := client.Providers(ctx)
			if providersErr == nil && providers["example-provider"].VehicleType != "HTTP" {
				t.Errorf("provider vehicle type = %q, want %q", providers["example-provider"].VehicleType, "HTTP")
			}

			return providersErr
		},
		"provider health check": func(ctx context.Context) error {
			return client.ProviderHealthCheck(ctx, "example-provider")
		},
		"proxy delay": func(ctx context.Context) error {
			delay, delayErr := client.ProxyDelay(ctx, "proxy/blue", "https://test.example/generate_204", 5*time.Second)
			if delayErr == nil && delay != 123 {
				t.Errorf("delay = %d, want %d", delay, 123)
			}

			return delayErr
		},
		"connections": func(ctx context.Context) error {
			connections, connectionsErr := client.Connections(ctx)
			if connectionsErr == nil && connections.DownloadTotal != 200 {
				t.Errorf("download total = %d, want %d", connections.DownloadTotal, 200)
			}

			return connectionsErr
		},
	}
	for name, call := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := call(t.Context()); err != nil {
				t.Errorf("call %s: %v", name, err)
			}
		})
	}
}

func TestClientErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		handler func(http.ResponseWriter, *http.Request)
		timeout time.Duration
		want    error
	}{
		"non successful status": {
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusUnauthorized)
			},
			timeout: time.Second,
		},
		"client timeout": {
			handler: func(_ http.ResponseWriter, request *http.Request) {
				<-request.Context().Done()
			},
			timeout: 10 * time.Millisecond,
			want:    context.DeadlineExceeded,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(testCase.handler))
			t.Cleanup(server.Close)

			client, err := mihomo.New(server.URL, "", testCase.timeout)
			if err != nil {
				t.Fatalf("new client: %v", err)
			}
			_, err = client.Version(t.Context())
			if err == nil {
				t.Fatal("version error = nil, want an error")
			}
			if testCase.want != nil && !errors.Is(err, testCase.want) {
				t.Errorf("version error = %v, want errors.Is(_, %v)", err, testCase.want)
			}
		})
	}
}

func writeFixture(t *testing.T, writer http.ResponseWriter, name string) {
	t.Helper()

	contents, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Errorf("read fixture %q: %v", name, err)
		return
	}
	if _, err := writer.Write(contents); err != nil {
		t.Errorf("write fixture %q: %v", name, err)
	}
}

func valuesEqual(got, want url.Values) bool {
	return got.Encode() == want.Encode()
}
