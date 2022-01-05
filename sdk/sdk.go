package sdk

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/cellargalaxy/go_common/consd"
	common_model "github.com/cellargalaxy/go_common/model"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"io"
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
	return NewFileBedClient(time.Minute, 3*time.Second, 3, &FileBedHandler{Address: address, Secret: secret})
}

func NewFileBedClient(timeout, sleep time.Duration, retry int, handler model.FileBedHandlerInter) (*FileBedClient, error) {
	if handler == nil {
		return nil, fmt.Errorf("FileBedHandlerInter为空")
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

func (this FileBedClient) AddFile(ctx context.Context, filePath string, reader io.Reader) (*model.FileAddResponse, error) {
	var jsonString string
	var object *model.FileAddResponse
	var err error
	for i := 0; i < this.retry; i++ {
		jwtToken, err := this.genJWT(ctx)
		if err != nil {
			return nil, err
		}
		jsonString, err = this.requestAddFile(ctx, jwtToken, filePath, reader)
		if err == nil {
			object, err = this.parseAddFile(ctx, jsonString)
			if err == nil {
				return object, err
			}
		}
	}
	return object, err
}

func (this FileBedClient) parseAddFile(ctx context.Context, jsonString string) (*model.FileAddResponse, error) {
	type Response struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data model.FileAddResponse `json:"data"`
	}
	var response Response
	err := util.UnmarshalJsonString(jsonString, &response)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加文件，解析响应异常")
		return nil, fmt.Errorf("添加文件，解析响应异常")
	}
	if response.Code != consd.HttpSuccessCode {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"code": response.Code, "msg": response.Msg}).Error("添加文件，失败")
		return nil, fmt.Errorf("添加文件，失败")
	}
	return &response.Data, nil
}

func (this FileBedClient) requestAddFile(ctx context.Context, jwtToken string, filePath string, reader io.Reader) (string, error) {
	response, err := this.httpClient.R().SetContext(ctx).
		SetHeader("Authorization", "Bearer "+jwtToken).
		SetHeader(util.LogIdKey, fmt.Sprint(util.GetLogId(ctx))).
		SetFileReader("file", filePath, reader).
		SetFormData(map[string]string{
			"path": filePath,
		}).
		Post(this.handler.GetAddress(ctx) + model.AddFileUrl)

	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加文件，请求异常")
		return "", fmt.Errorf("添加文件，请求异常")
	}
	if response == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加文件，响应为空")
		return "", fmt.Errorf("添加文件，响应为空")
	}
	statusCode := response.StatusCode()
	body := response.String()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode, "body": body}).Info("添加文件，响应")
	if statusCode != http.StatusOK {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"StatusCode": statusCode}).Error("添加文件，响应码失败")
		return "", fmt.Errorf("添加文件，响应码失败: %+v", statusCode)
	}
	return body, nil
}

func (this FileBedClient) GetFileCompleteInfo(ctx context.Context, request model.FileCompleteInfoGetRequest) (*model.FileCompleteInfoGetResponse, error) {
	var jsonString string
	var object *model.FileCompleteInfoGetResponse
	var err error
	for i := 0; i < this.retry; i++ {
		jwtToken, err := this.genJWT(ctx)
		if err != nil {
			return nil, err
		}
		jsonString, err = this.requestGetFileCompleteInfo(ctx, jwtToken, request)
		if err == nil {
			object, err = this.parseGetFileCompleteInfo(ctx, jsonString)
			if err == nil {
				return object, err
			}
		}
	}
	return object, err
}

func (this FileBedClient) parseGetFileCompleteInfo(ctx context.Context, jsonString string) (*model.FileCompleteInfoGetResponse, error) {
	type Response struct {
		Code int                               `json:"code"`
		Msg  string                            `json:"msg"`
		Data model.FileCompleteInfoGetResponse `json:"data"`
	}
	var response Response
	err := util.UnmarshalJsonString(jsonString, &response)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询文件完整信息，解析响应异常")
		return nil, fmt.Errorf("查询文件完整信息，解析响应异常")
	}
	if response.Code != consd.HttpSuccessCode {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"code": response.Code, "msg": response.Msg}).Error("查询文件完整信息，失败")
		return nil, fmt.Errorf("查询文件完整信息，失败")
	}
	return &response.Data, nil
}

func (this FileBedClient) requestGetFileCompleteInfo(ctx context.Context, jwtToken string, request model.FileCompleteInfoGetRequest) (string, error) {
	response, err := this.httpClient.R().SetContext(ctx).
		SetHeader("Authorization", "Bearer "+jwtToken).
		SetHeader(util.LogIdKey, fmt.Sprint(util.GetLogId(ctx))).
		SetQueryParam("path", request.Path).
		Get(this.handler.GetAddress(ctx) + model.GetFileCompleteInfoUrl)

	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询文件完整信息，请求异常")
		return "", fmt.Errorf("查询文件完整信息，请求异常")
	}
	if response == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询文件完整信息，响应为空")
		return "", fmt.Errorf("查询文件完整信息，响应为空")
	}
	statusCode := response.StatusCode()
	body := response.String()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode, "body": body}).Info("查询文件完整信息，响应")
	if statusCode != http.StatusOK {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"StatusCode": statusCode}).Error("查询文件完整信息，响应码失败")
		return "", fmt.Errorf("查询文件完整信息，响应码失败: %+v", statusCode)
	}
	return body, nil
}

func (this FileBedClient) ListFileSimpleInfo(ctx context.Context, request model.FileSimpleInfoListRequest) (*model.FileSimpleInfoListResponse, error) {
	var jsonString string
	var object *model.FileSimpleInfoListResponse
	var err error
	for i := 0; i < this.retry; i++ {
		jwtToken, err := this.genJWT(ctx)
		if err != nil {
			return nil, err
		}
		jsonString, err = this.requestListFileSimpleInfo(ctx, jwtToken, request)
		if err == nil {
			object, err = this.parseListFileSimpleInfo(ctx, jsonString)
			if err == nil {
				return object, err
			}
		}
	}
	return object, err
}

func (this FileBedClient) parseListFileSimpleInfo(ctx context.Context, jsonString string) (*model.FileSimpleInfoListResponse, error) {
	type Response struct {
		Code int                              `json:"code"`
		Msg  string                           `json:"msg"`
		Data model.FileSimpleInfoListResponse `json:"data"`
	}
	var response Response
	err := util.UnmarshalJsonString(jsonString, &response)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询文件简单信息，解析响应异常")
		return nil, fmt.Errorf("查询文件简单信息，解析响应异常")
	}
	if response.Code != consd.HttpSuccessCode {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"code": response.Code, "msg": response.Msg}).Error("查询文件简单信息，失败")
		return nil, fmt.Errorf("查询文件简单信息，失败")
	}
	return &response.Data, nil
}

func (this FileBedClient) requestListFileSimpleInfo(ctx context.Context, jwtToken string, request model.FileSimpleInfoListRequest) (string, error) {
	response, err := this.httpClient.R().SetContext(ctx).
		SetHeader("Authorization", "Bearer "+jwtToken).
		SetHeader(util.LogIdKey, fmt.Sprint(util.GetLogId(ctx))).
		SetQueryParam("path", request.Path).
		Get(this.handler.GetAddress(ctx) + model.ListFileSimpleInfoUrl)

	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询文件简单信息，请求异常")
		return "", fmt.Errorf("查询文件简单信息，请求异常")
	}
	if response == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询文件简单信息，响应为空")
		return "", fmt.Errorf("查询文件简单信息，响应为空")
	}
	statusCode := response.StatusCode()
	body := response.String()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode, "body": body}).Info("查询文件简单信息，响应")
	if statusCode != http.StatusOK {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"StatusCode": statusCode}).Error("查询文件简单信息，响应码失败")
		return "", fmt.Errorf("查询文件简单信息，响应码失败: %+v", statusCode)
	}
	return body, nil
}

func (this FileBedClient) genJWT(ctx context.Context) (string, error) {
	now := time.Now()
	var claims common_model.Claims
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = now.Unix() + int64(this.retry*60)
	claims.RequestId = fmt.Sprint(util.GenId())
	jwtToken, err := util.GenJWT(ctx, this.handler.GetSecret(ctx), claims)
	return jwtToken, err
}
