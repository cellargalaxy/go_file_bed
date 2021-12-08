package sdk

import (
	"context"
	"crypto/tls"
	"fmt"
	common_model "github.com/cellargalaxy/go_common/model"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type FileBedHandler struct {
	Address string `json:"address"`
	Secret  string `json:"-"`
}

func (this FileBedHandler) String() string {
	return util.ToJsonString(this)
}

func (this FileBedHandler) GetAddress(ctx context.Context) string {
	return this.Address
}
func (this FileBedHandler) GetSecret(ctx context.Context) string {
	return this.Secret
}

type FileBedClient struct {
	retry      int
	handler    model.FileBedHandlerInter
	httpClient *resty.Client
}

func NewDefaultFileBedClient(address, secret string) (*FileBedClient, error) {
	return NewFileBedClient(3*time.Second, 3*time.Second, 3, &FileBedHandler{Address: address, Secret: secret})
}

func NewFileBedClient(timeout, sleep time.Duration, retry int, handler model.FileBedHandlerInter) (*FileBedClient, error) {
	if handler == nil {
		return nil, fmt.Errorf("MsgHandlerInter为空")
	}
	httpClient := createHttpClient(timeout, sleep, retry)
	return &FileBedClient{retry: retry, handler: handler, httpClient: httpClient}, nil
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

//发送微信通用模板信息
func (this FileBedClient) SendTemplateToCommonTag(ctx context.Context, text string) (bool, error) {
	var jsonString string
	var object bool
	var err error
	for i := 0; i < this.retry; i++ {
		jwtToken, err := this.genJWT(ctx)
		if err != nil {
			return false, err
		}
		jsonString, err = this.requestSendTemplateToCommonTag(ctx, jwtToken, text)
		if err == nil {
			object, err = this.analysisSendTemplateToCommonTag(ctx, jsonString)
			if err == nil {
				return object, err
			}
		}
	}
	return object, err
}

//发送微信通用模板信息
func (this FileBedClient) analysisSendTemplateToCommonTag(ctx context.Context, jsonString string) (bool, error) {
	return this.analysisSendWxTemplateToTag(ctx, jsonString)
}

//发送微信通用模板信息
func (this FileBedClient) requestSendTemplateToCommonTag(ctx context.Context, jwtToken string, text string) (string, error) {
	response, err := this.httpClient.R().SetContext(ctx).
		SetHeader("Content-Type", "application/json;CHARSET=utf-8").
		SetHeader("Authorization", "Bearer "+jwtToken).
		SetHeader(util.LogIdKey, fmt.Sprint(util.GetLogId(ctx))).
		SetBody(map[string]interface{}{
			"text": text,
		}).
		Post(this.handler.GetAddress(ctx) + "/api/sendTemplateToCommonTag")

	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("发送微信通用模板信息，请求异常")
		return "", fmt.Errorf("发送微信通用模板信息，请求异常")
	}
	if response == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("发送微信通用模板信息，响应为空")
		return "", fmt.Errorf("发送微信通用模板信息，响应为空")
	}
	statusCode := response.StatusCode()
	body := response.String()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode, "body": body}).Info("发送微信通用模板信息，响应")
	if statusCode != http.StatusOK {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"StatusCode": statusCode}).Error("发送微信通用模板信息，响应码失败")
		return "", fmt.Errorf("发送微信通用模板信息，响应码失败: %+v", statusCode)
	}
	return body, nil
}

func (this FileBedClient) genJWT(ctx context.Context) (string, error) {
	now := time.Now()
	var claims common_model.Claims
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = now.Unix() + int64(this.retry*3)
	claims.RequestId = fmt.Sprint(util.GenId())
	jwtToken, err := util.GenJWT(ctx, this.handler.GetSecret(ctx), claims)
	return jwtToken, err
}
