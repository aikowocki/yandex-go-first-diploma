package httpclient

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type RetryClient struct {
	client     *http.Client
	maxRetries int
	baseDelay  time.Duration
}

func NewRetryClient(maxRetries int, timeout time.Duration) *RetryClient {
	return &RetryClient{
		client:     &http.Client{Timeout: timeout},
		maxRetries: maxRetries,
		baseDelay:  1 * time.Second,
	}
}

func (c *RetryClient) Do(r *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err = c.client.Do(r)

		if err == nil && !isRetryable(resp.StatusCode) {
			return resp, nil
		}

		if attempt < c.maxRetries {
			delay := c.baseDelay * time.Duration(1<<attempt)
			zap.S().Warnw("retrying request",
				"attempt", attempt+1,
				"delay", delay,
				"error", err,
			)
			// ожидаем но уважаем отмену по контексту
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				return nil, r.Context().Err()
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return resp, err
}

func isRetryable(statusCode int) bool {
	return statusCode >= 500
}
