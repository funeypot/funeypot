package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

func StartReport() {
	if abuseIpDbKey == "" {
		slog.Info("skip report abuse ip db")
		return
	}

	report()
}

func report() {
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
		reportAbuseIpDb(v)
	}

	time.AfterFunc(time.Minute, report)
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

func reportAbuseIpDb(record Record) {
	data := url.Values{}
	data.Set("ip", record.Ip)
	data.Add("categories", "18,22")
	data.Add("comment", fmt.Sprintf("Caught by honeypots, this IP has attempted to crack the SSH login password %d times within the past %s.", record.Count, record.Duration().String()))
	data.Add("timestamp", time.Now().Format(time.RFC3339))

	req, err := http.NewRequest(http.MethodPost, "https://api.abuseipdb.com/api/v2/report", bytes.NewBufferString(data.Encode()))
	if err != nil {
		slog.Error("new request",
			"error", err,
		)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Key", abuseIpDbKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("do request",
			"error", err,
		)
		return
	}
	defer resp.Body.Close()

	result := &abuseIpDbResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		slog.Error("decode response",
			"error", err,
		)
		return
	}

	if len(result.Errors) > 0 {
		slog.Error("report abuse ip db",
			"errors", result.Errors,
		)
		return
	}

	slog.Debug("report abuse ip db",
		"ip", result.Data.IpAddress,
		"score", result.Data.AbuseConfidenceScore,
	)
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
