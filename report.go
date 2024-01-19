package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"sshless/internal/pkg/logs"
)

func StartReport(ctx context.Context) {
	if abuseIpDbKey == "" {
		logs.From(ctx).Infof("skip report abuse ip db")
		return
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
			case <-ticker.C:
				report(ctx)
			}
		}
	}()
}

func report(ctx context.Context) {
	var toReport []Record

	recordM.Lock()
	for _, v := range records {
		if time.Since(v.ReportedAt) < 15*time.Minute {
			continue
		}
		if time.Since(v.StoppedAt) > time.Hour {
			continue
		}
		if v.Count < 5 {
			continue
		}
		v.ReportedAt = time.Now()
		toReport = append(toReport, *v)
	}
	recordM.Unlock()

	for _, v := range toReport {
		reportAbuseIpDb(ctx, v)
	}
}

/*
curl https://api.abuseipdb.com/api/v2/report \
  --data-urlencode "ip=127.0.0.1" \
  -d categories=18,22 \
  --data-urlencode "comment=SSH login attempts with user root." \
  --data-urlencode "timestamp=2023-10-18T11:25:11-04:00" \
  -H "Key: YOUR_OWN_API_KEY" \
  -H "Accept: application/json"
*/

func reportAbuseIpDb(ctx context.Context, record Record) {
	data := url.Values{}
	data.Set("ip", record.Ip)
	data.Add("categories", "18,22")
	data.Add("comment", fmt.Sprintf("Caught by honeypots, tried to crack SSH password %d times within %s.", record.Count, record.Duration().Truncate(time.Second).String()))
	data.Add("timestamp", time.Now().Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.abuseipdb.com/api/v2/report", bytes.NewBufferString(data.Encode()))
	if err != nil {
		logs.From(ctx).Errorf("new request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Key", abuseIpDbKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logs.From(ctx).Errorf("do request: %v", err)
		return
	}
	defer resp.Body.Close()

	result := &abuseIpDbResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		logs.From(ctx).Errorf("decode response: %v", err)
		return
	}

	if len(result.Errors) > 0 {
		logs.From(ctx).Errorf("report abuse ip db: %v", result.Errors)
		return
	}

	logs.From(ctx).With("ip", result.Data.IpAddress, "score", result.Data.AbuseConfidenceScore).Infof("report abuse ip db")
}

type abuseIpDbResponse struct {
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
