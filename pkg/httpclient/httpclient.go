package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	maxRetries int
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		maxRetries: 3,
	}
}

func (c *Client) DoRequest(
	ctx context.Context,
	method string,
	url string,
	body []byte,
	headers map[string]string,
) ([]byte, int, error) {

	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, 0, err
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
			continue
		}

		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, err
		}

		// Retry on 5xx errors
		if resp.StatusCode >= 500 && attempt < c.maxRetries {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
			continue
		}

		return respBody, resp.StatusCode, nil
	}

	return nil, 0, lastErr
}
