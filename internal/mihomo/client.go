// Package mihomo provides a client for the Mihomo external-controller API.
package mihomo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	errInvalidAPIURL  = errors.New("mihomo API URL must use http or https")
	errInvalidTimeout = errors.New("mihomo API timeout must be greater than zero")
)

// Client communicates with a Mihomo external-controller API.
type Client struct {
	baseURL    *url.URL
	secret     string
	httpClient *http.Client
}

// New creates a Client using apiURL as the external-controller URL.
func New(apiURL, secret string, timeout time.Duration) (*Client, error) {
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parse mihomo API URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" || parsedURL.Host == "" {
		return nil, fmt.Errorf("validate mihomo API URL %q: %w", apiURL, errInvalidAPIURL)
	}
	if timeout <= 0 {
		return nil, errInvalidTimeout
	}

	return &Client{
		baseURL: parsedURL,
		secret:  secret,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) get(ctx context.Context, segments []string, query url.Values, destination any) error {
	requestURL, err := c.endpointURL(segments, query)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("create Mihomo API request: %w", err)
	}
	if c.secret != "" {
		request.Header.Set("Authorization", "Bearer "+c.secret)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send Mihomo API request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, copyErr := io.Copy(io.Discard, response.Body)
		if copyErr != nil {
			return fmt.Errorf("drain Mihomo API response: %w", copyErr)
		}

		return responseStatusError{path: request.URL.EscapedPath(), status: response.Status}
	}
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode Mihomo API response: %w", err)
	}

	return nil
}

func (c *Client) endpointURL(segments []string, query url.Values) (string, error) {
	escapedSegments := make([]string, 0, len(segments)+1)
	escapedSegments = append(escapedSegments, strings.TrimSuffix(c.baseURL.EscapedPath(), "/"))
	for _, segment := range segments {
		escapedSegments = append(escapedSegments, url.PathEscape(segment))
	}
	escapedPath := strings.Join(escapedSegments, "/")
	path, err := url.PathUnescape(escapedPath)
	if err != nil {
		return "", fmt.Errorf("unescape Mihomo API path: %w", err)
	}

	endpoint := *c.baseURL
	endpoint.Path = path
	endpoint.RawPath = escapedPath
	endpoint.RawQuery = query.Encode()

	return endpoint.String(), nil
}

type responseStatusError struct {
	path   string
	status string
}

func (e responseStatusError) Error() string {
	return fmt.Sprintf("mihomo API GET %s: unexpected status %s", e.path, e.status)
}
