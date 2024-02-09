package agentcore

import (
	"encoding/json"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/go-resty/resty/v2"
)

func SendMetrics(metrics []MetricInterface, serverURL string) {
	for _, metric := range metrics {
		sendingMetric, err := json.Marshal(metric)
		if err != nil {
			logger.Sugar.Errorln("error marshaling json: %v", err)
		}
		urlFormat := "/update/"
		url := serverURL + urlFormat

		client := resty.New()
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

func SendBatchMetrics(metrics []MetricInterface, serverURL string) {
	sendingMetrics, err := json.Marshal(metrics)
	if err != nil {
		logger.Sugar.Errorln("error marshaling json: %v", err)
	}
	urlFormat := "/updates/"
	url := serverURL + urlFormat

	client := resty.New()
	_, err = client.R().
		SetHeader("Content-type", "application/json").
		SetBody(sendingMetrics).
		Post(url)

	if err != nil {
		logger.Sugar.Errorln("error sending metrics: %v", err)
	}
}
