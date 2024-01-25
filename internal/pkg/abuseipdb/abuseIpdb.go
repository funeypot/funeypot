package abuseipdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	key string
}

func NewClient(key string) *Client {
	return &Client{
		key: key,
	}
}

func (c *Client) ReportSsh(ctx context.Context, ip string, timestamp time.Time, comment string) (int, error) {
	// see https://www.abuseipdb.com/categories
	return c.Report(ctx, ip, []string{"18", "22"}, timestamp, comment)
}

func (c *Client) ReportHttp(ctx context.Context, ip string, timestamp time.Time, comment string) (int, error) {
	// see https://www.abuseipdb.com/categories
	return c.Report(ctx, ip, []string{"18", "21"}, timestamp, comment)
}

func (c *Client) Report(ctx context.Context, ip string, categories []string, timestamp time.Time, comment string) (int, error) {
	result := &response{}

	_, err := resty.New().R().
		SetContext(ctx).
		SetHeader("Key", c.key).
		SetFormData(map[string]string{
			"ip":         ip,
			"categories": strings.Join(categories, ","),
			"timestamp":  timestamp.Format(time.RFC3339),
			"comment":    comment,
		}).
		SetResult(result).
		Post("https://api.abuseipdb.com/api/v2/report")
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}

	if len(result.Errors) > 0 {
		return 0, fmt.Errorf("response: %v", result.Errors)
	}

	return result.Data.AbuseConfidenceScore, nil
}

type response struct {
	Data struct {
		IpAddress            string `json:"ipAddress"`
		AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
	} `json:"data"`
	Errors []struct {
		Detail string `json:"detail"`
		Status int    `json:"status"`
		Source struct {
			Parameter string `json:"parameter"`
		} `json:"source"`
	} `json:"errors"`
}
