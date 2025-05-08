// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package selfip

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

var (
	urls = []string{
		"https://checkip.amazonaws.com",
		"http://ip-api.com/line/?fields=query",
		"https://api.ipify.org",
		"https://icanhazip.com",
		//"https://ifconfig.me", // it checks for user agent and returns HTML
		"https://ipinfo.io/ip",
		"https://ipecho.net/plain",
		"https://myexternalip.com/raw",
	}
)

func Get(ctx context.Context) (string, error) {
	errs := make([]string, 0, len(urls))
	for _, url := range urls {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		ip, err := getFrom(ctx, url)
		if err == nil {
			return ip, nil
		}
		errs = append(errs, fmt.Sprintf("%s: %v", url, err))
	}

	return "", fmt.Errorf("all attempts failed: %s", strings.Join(errs, "; "))
}

func getFrom(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	ip := strings.TrimSpace(string(body))

	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid response: %q", ip)
	}

	return ip, nil
}
