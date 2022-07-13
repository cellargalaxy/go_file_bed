package sdk

import (
	"context"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type FileBedHandlerInter interface {
	ListAddress(ctx context.Context) []string
	GetSecret(ctx context.Context) string
}

type FileBedHandler struct {
	Address string `json:"address"`
	Secret  string `json:"-"`
}

func (this FileBedHandler) String() string {
	return util.ToJsonString(this)
}

func (this FileBedHandler) ListAddress(ctx context.Context) []string {
	address := this.Address
	if strings.HasSuffix(address, "/") {
		address = address[:len(address)-1]
		this.Address = address
	}
	return []string{address}
}
func (this FileBedHandler) GetSecret(ctx context.Context) string {
	return this.Secret
}

type FileBedClient struct {
	timeout        time.Duration
	retry          int
	httpClient     *resty.Client
	httpClientLong *resty.Client
	handler        FileBedHandlerInter
}

func NewDefaultFileBedClient(ctx context.Context, address, secret string) (*FileBedClient, error) {
	httpClientLong := util.CreateNotRetryHttpClient(time.Hour)
	return NewFileBedClient(ctx, util.TimeoutDefault, util.RetryDefault, util.GetHttpClient(), httpClientLong, &FileBedHandler{Address: address, Secret: secret})
}

func NewFileBedClient(ctx context.Context, timeout time.Duration, retry int, httpClient, httpClientLong *resty.Client, handler FileBedHandlerInter) (*FileBedClient, error) {
	if handler == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Error("创建FileBedClient，FileBedHandlerInter为空")
		return nil, fmt.Errorf("创建FileBedClient，FileBedHandlerInter为空")
	}
	return &FileBedClient{timeout: timeout, retry: retry, handler: handler, httpClient: httpClient, httpClientLong: httpClientLong}, nil
}

func (this *FileBedClient) DownloadFile(ctx context.Context, filePath string, writer io.Writer) error {
	url, err := this.GetFileDownloadUrl(ctx, filePath)
	if err != nil {
		return err
	}

	response, err := this.httpClientLong.R().SetContext(ctx).
		SetDoNotParseResponse(true).
		Get(url)

	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("下载文件，文件下载异常")
		return fmt.Errorf("下载文件，文件下载异常")
	}
	if response == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("下载文件，文件下载响应为空")
		return fmt.Errorf("下载文件，文件下载响应为空")
	}
	statusCode := response.StatusCode()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode}).Info("下载文件，文件下载响应")
	if statusCode != http.StatusOK {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"StatusCode": statusCode}).Error("下载文件，文件下载响应码失败")
		return fmt.Errorf("下载文件，文件下载响应码失败: %+v", statusCode)
	}

	reader := response.RawBody()
	defer reader.Close()
	written, err := io.Copy(writer, reader)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("下载文件，拷贝数据异常")
		return fmt.Errorf("下载文件，拷贝数据异常: %+v", err)
	} else {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"filePath": filePath, "written": written}).Info("下载文件，拷贝数据完成")
	}
	return nil
}

func (this *FileBedClient) GetFileDownloadUrl(ctx context.Context, filePath string) (string, error) {
	url := this.getAddress(ctx)
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	url += path.Join(model.FileUrl, filePath)
	logrus.WithContext(ctx).WithFields(logrus.Fields{"url": url}).Info("获取下载文件链接")
	return url, nil
}

func (this *FileBedClient) AddFile(ctx context.Context, filePath string, reader io.Reader, raw bool) (*model.FileSimpleInfo, error) {
	var jsonString string
	var object *model.FileAddResponse
	var err error
	for i := 0; i < this.retry; i++ {
		jsonString, err = this.requestAddFile(ctx, filePath, reader, raw)
		if err == nil {
			object, err = this.parseAddFile(ctx, jsonString)
			if object != nil && err == nil {
				return object.Info, err
			}
		}
	}
	return nil, err
}
func (this *FileBedClient) parseAddFile(ctx context.Context, jsonString string) (*model.FileAddResponse, error) {
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
	if response.Code != util.HttpSuccessCode {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"code": response.Code, "msg": response.Msg}).Error("添加文件，失败")
		return nil, fmt.Errorf("添加文件，失败")
	}
	return &response.Data, nil
}
func (this *FileBedClient) requestAddFile(ctx context.Context, filePath string, reader io.Reader, raw bool) (string, error) {
	response, err := this.httpClient.R().SetContext(ctx).
		SetHeader(this.genJWT(ctx)).
		SetFileReader("file", filePath, reader).
		SetFormData(map[string]string{
			"path": filePath,
			"raw":  strconv.FormatBool(raw),
		}).
		Post(this.GetUrl(ctx, model.AddFileUrl))

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

func (this *FileBedClient) GetFileCompleteInfo(ctx context.Context, request model.FileCompleteInfoGetRequest) (*model.FileCompleteInfo, error) {
	var jsonString string
	var object *model.FileCompleteInfoGetResponse
	var err error
	for i := 0; i < this.retry; i++ {
		jsonString, err = this.requestGetFileCompleteInfo(ctx, request)
		if err == nil {
			object, err = this.parseGetFileCompleteInfo(ctx, jsonString)
			if object != nil && err == nil {
				return object.Info, err
			}
		}
	}
	return nil, err
}
func (this *FileBedClient) parseGetFileCompleteInfo(ctx context.Context, jsonString string) (*model.FileCompleteInfoGetResponse, error) {
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
	if response.Code != util.HttpSuccessCode {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"code": response.Code, "msg": response.Msg}).Error("查询文件完整信息，失败")
		return nil, fmt.Errorf("查询文件完整信息，失败")
	}
	return &response.Data, nil
}
func (this *FileBedClient) requestGetFileCompleteInfo(ctx context.Context, request model.FileCompleteInfoGetRequest) (string, error) {
	response, err := this.httpClient.R().SetContext(ctx).
		SetHeader(this.genJWT(ctx)).
		SetQueryParam("path", request.Path).
		Get(this.GetUrl(ctx, model.GetFileCompleteInfoUrl))

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

func (this *FileBedClient) ListFileSimpleInfo(ctx context.Context, request model.FileSimpleInfoListRequest) ([]model.FileSimpleInfo, error) {
	var jsonString string
	var object *model.FileSimpleInfoListResponse
	var err error
	for i := 0; i < this.retry; i++ {
		jsonString, err = this.requestListFileSimpleInfo(ctx, request)
		if err == nil {
			object, err = this.parseListFileSimpleInfo(ctx, jsonString)
			if object != nil && err == nil {
				return object.Infos, err
			}
		}
	}
	return nil, err
}
func (this *FileBedClient) parseListFileSimpleInfo(ctx context.Context, jsonString string) (*model.FileSimpleInfoListResponse, error) {
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
	if response.Code != util.HttpSuccessCode {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"code": response.Code, "msg": response.Msg}).Error("查询文件简单信息，失败")
		return nil, fmt.Errorf("查询文件简单信息，失败")
	}
	return &response.Data, nil
}
func (this *FileBedClient) requestListFileSimpleInfo(ctx context.Context, request model.FileSimpleInfoListRequest) (string, error) {
	response, err := this.httpClient.R().SetContext(ctx).
		SetHeader(this.genJWT(ctx)).
		SetQueryParam("path", request.Path).
		Get(this.GetUrl(ctx, model.ListFileSimpleInfoUrl))

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

func (this *FileBedClient) GetUrl(ctx context.Context, path string) string {
	return this.getUrl(ctx, this.getAddress(ctx), path)
}
func (this *FileBedClient) getUrl(ctx context.Context, address, path string) string {
	if strings.HasSuffix(address, "/") && strings.HasPrefix(path, "/") && len(path) > 0 {
		path = path[1:]
	}
	return address + path
}
func (this *FileBedClient) getAddress(ctx context.Context) string {
	list := this.handler.ListAddress(ctx)
	if len(list) == 0 {
		return ""
	}
	logId := util.GetLogId(ctx)
	index := int(logId) % len(list)
	return list[index]
}
func (this *FileBedClient) genJWT(ctx context.Context) (string, string) {
	return util.GenAuthorizationJWT(ctx, this.timeout, this.handler.GetSecret(ctx))
}
