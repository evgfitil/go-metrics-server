package agentcore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/go-resty/resty/v2"
	"time"
)

const (
	retryCount       = 3
	retryWait        = 2 * time.Second
	retryMaxWaitTime = 5 * time.Second
)

func SendMetrics(key string, metrics []MetricInterface, serverURL string) {
	for _, metric := range metrics {
		sendingMetric, err := json.Marshal(metric)
		if err != nil {
			logger.Sugar.Errorln("error marshaling json: %v", err)
		}
		urlFormat := "/update/"
		url := serverURL + urlFormat

		client := resty.New()
		client.
			SetRetryCount(retryCount).
			SetRetryWaitTime(retryWait).
			SetRetryMaxWaitTime(retryMaxWaitTime)
		resp := client.R().
			SetHeader("Content-type", "application/json").
			SetBody(sendingMetric)
		if key != "" {
			hash := computeHash(key, sendingMetric)
			client.R().SetHeader("HashSHA256", hash)
		}
		_, err = resp.Post(url)
		if err != nil {
			logger.Sugar.Errorln("error sending metric: %v", err)
			continue
		}
	}
}

func SendBatchMetrics(key string, metrics []MetricInterface, serverURL string) {
	sendingMetrics, err := json.Marshal(metrics)
	if err != nil {
		logger.Sugar.Errorln("error marshaling json: %v", err)
	}
	urlFormat := "/updates/"
	url := serverURL + urlFormat

	client := resty.New()
	client.
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryWait).
		SetRetryMaxWaitTime(retryMaxWaitTime)
	resp := client.R().
		SetHeader("Content-type", "application/json").
		SetBody(sendingMetrics)
	if key != "" {
		hash := computeHash(key, sendingMetrics)
		client.R().SetHeader("HashSHA256", hash)
	}
	_, err = resp.Post(url)

	if err != nil {
		logger.Sugar.Errorf("error sending metrics: %v", err)
	}
}

func computeHash(key string, data []byte) string {
	secretKey := []byte(key)
	h := hmac.New(sha256.New, secretKey)
	h.Write(data)
	dst := h.Sum(nil)
	return fmt.Sprintf("%x", dst)
}
