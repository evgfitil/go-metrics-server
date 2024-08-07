// Package agentcore provides core functionalities for the agent to collect metrics
// from the system and send them to the server.
package agentcore

import (
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/evgfitil/go-metrics-server.git/internal/logger"
)

const (
	retryCount       = 3
	retryWait        = 2 * time.Second
	retryMaxWaitTime = 5 * time.Second
)

// SendMetrics sends individual metrics to the server.
func SendMetrics(metrics []MetricInterface, serverURL string) {
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
		_, err = client.R().
			SetHeader("Content-type", "application/json").
			SetBody(sendingMetric).
			Post(url)

		if err != nil {
			logger.Sugar.Errorln("error sending metric: %v", err)
			continue
		}
	}
}

// SendBatchMetrics sends a batch of metrics to the server.
func SendBatchMetrics(metrics []MetricInterface, serverURL string) {
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
	_, err = client.R().
		SetHeader("Content-type", "application/json").
		SetBody(sendingMetrics).
		Post(url)

	if err != nil {
		logger.Sugar.Errorf("error sending metrics: %v", err)
	}
}
