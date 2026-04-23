package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type ErrTooManyRequests struct {
	RetryAfter time.Duration
}

func (e *ErrTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests, retry after %s", e.RetryAfter)
}

func (c *Client) GetOrderAccrual(ctx context.Context, number string) (*AccrualResponse, error) {
	url := c.baseURL + "/api/orders/" + number
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	switch resp.StatusCode {
	case http.StatusOK:
		var res AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return nil, err
		}
		return &res, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return nil, &ErrTooManyRequests{
			RetryAfter: time.Duration(retryAfter) * time.Second,
		}
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
