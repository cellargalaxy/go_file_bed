package sdk

import (
	"crypto/tls"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type MsgClient struct {
	retry      int
	handler    model.MsgHandlerInter
	httpClient *resty.Client
}

func NewDefaultMsgClient() (*MsgClient, error) {
	return NewMsgClient(3*time.Second, 3*time.Second, 3, &MsgHandler{})
}

func NewMsgClient(timeout, sleep time.Duration, retry int, handler model.MsgHandlerInter) (*MsgClient, error) {
	if handler == nil {
		return nil, fmt.Errorf("MsgHandlerInter为空")
	}
	httpClient := createHttpClient(timeout, sleep, retry)
	return &MsgClient{retry: retry, handler: handler, httpClient: httpClient}, nil
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
			retry := statusCode != http.StatusOK || err != nil
			if retry {
				logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode, "err": err}).Warn("HTTP请求异常，进行重试")
			}
			return retry
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
