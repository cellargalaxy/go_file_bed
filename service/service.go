package service

import (
	"crypto/tls"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var httpClient *resty.Client

func init() {
	httpClient = createHttpClient(config.Config.Timeout, config.Config.Sleep, config.Config.Retry)
}

func createHttpClient(timeout, sleep time.Duration, retry int) *resty.Client {
	httpClient := resty.New().
		SetTimeout(timeout).
		SetRetryCount(retry).
		SetRetryWaitTime(sleep).
		SetRetryMaxWaitTime(5 * time.Minute).
		AddRetryCondition(func(response *resty.Response, err error) bool {
			ctx := util.CreateLogCtx()
			if response != nil && response.Request != nil {
				ctx = response.Request.Context()
			}
			var statusCode int
			if response != nil {
				statusCode = response.StatusCode()
			}
			isRetry := statusCode != http.StatusOK || err != nil
			if isRetry {
				logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode, "err": err}).Warn("HTTP请求异常，进行重试")
			}
			return isRetry
		}).
		SetRetryAfter(func(client *resty.Client, response *resty.Response) (time.Duration, error) {
			ctx := util.CreateLogCtx()
			if response != nil && response.Request != nil {
				ctx = response.Request.Context()
			}
			var attempt int
			if response != nil && response.Request != nil {
				attempt = response.Request.Attempt
			}
			if attempt > retry {
				logrus.WithContext(ctx).WithFields(logrus.Fields{"attempt": attempt}).Error("HTTP请求异常，超过最大重试次数")
				return 0, fmt.Errorf("HTTP请求异常，超过最大重试次数")
			}
			duration := util.WareDuration(sleep)
			for i := 0; i < attempt-1; i++ {
				duration *= 10
			}
			logrus.WithContext(ctx).WithFields(logrus.Fields{"attempt": attempt, "duration": duration}).Warn("HTTP请求异常，休眠重试")
			return duration, nil
		}).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	return httpClient
}
