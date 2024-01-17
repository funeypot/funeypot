package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func IpGeo(ctx context.Context, ip string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://ip-api.com/json/%s?fields=status,message,continent,country,regionName,city,district,isp&lang=zh-CN", ip), nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	result := &GeoResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if result.Status != "success" {
		return "", fmt.Errorf("status: %s", result.Status)
	}

	var strs []string
	if result.Continent != "" {
		strs = append(strs, result.Continent)
	}
	if result.Country != "" {
		strs = append(strs, result.Country)
	}
	if result.RegionName != "" {
		strs = append(strs, result.RegionName)
	}
	if result.City != "" {
		strs = append(strs, result.City)
	}
	if result.District != "" {
		strs = append(strs, result.District)
	}
	if result.Isp != "" {
		strs = append(strs, result.Isp)
	}
	return strings.Join(strs, ","), nil
}

type GeoResponse struct {
	Status     string `json:"status"`
	Continent  string `json:"continent"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	District   string `json:"district"`
	Isp        string `json:"isp"`
}
