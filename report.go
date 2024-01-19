package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
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
	records, err := Cache.AllRecords(ctx)
	if err != nil {
		logs.From(ctx).Errorf("get all records: %v", err)
	}

	for _, v := range records {
		if time.Since(v.ReportedAt) < 20*time.Minute {
			continue
		}
		if time.Since(v.StoppedAt) > time.Hour {
			continue
		}
		if v.Count < 5 {
			continue
		}
		if err := reportAbuseIpDb(ctx, v); err != nil {
			logs.From(ctx).Errorf("report abuse ip db: %v", err)
			continue
		}
		v.ReportedAt = time.Now()
		if err := Cache.SetRecord(ctx, v); err != nil {
			logs.From(ctx).Errorf("set record: %v", err)
		}
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

func reportAbuseIpDb(ctx context.Context, record *model.Record) error {
	data := url.Values{}
	data.Set("ip", record.Ip)
	data.Add("categories", "18,22")
	data.Add("comment", fmt.Sprintf("Caught by Funeypot, tried to crack SSH password %d times within %s.", record.Count, record.Duration().Truncate(time.Second).String()))
	data.Add("timestamp", time.Now().Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.abuseipdb.com/api/v2/report", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Key", abuseIpDbKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	result := &abuseIpDbResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		logs.From(ctx).Errorf("report abuse ip db: %v", result.Errors)
		return fmt.Errorf("report abuse ip db: %v", result.Errors)
	}

	logs.From(ctx).With("ip", result.Data.IpAddress, "score", result.Data.AbuseConfidenceScore).Debugf("report abuse ip db")
	return nil
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
