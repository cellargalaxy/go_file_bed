package service

import (
	"encoding/json"
	"fmt"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/cellargalaxy/go_file_bed/dao"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

func init() {
	c := cron.New()
	if config.PullSyncCron != "" {
		err := c.AddFunc(config.PullSyncCron, func() { PullSyncFile(config.PullSyncHost, config.PullSyncToken) })
		if err == nil {
			logrus.WithFields(logrus.Fields{}).Info("添加自动pull任务成功")
		} else {
			logrus.WithFields(logrus.Fields{"err": err}).Error("添加自动pull任务失败")
		}
	} else {
		logrus.WithFields(logrus.Fields{}).Info("自动pull任务cron为空")
	}
	if config.PushSyncCron != "" {
		err := c.AddFunc(config.PushSyncCron, func() { PushSyncFile(config.PushSyncHost, config.PushSyncToken) })
		if err == nil {
			logrus.WithFields(logrus.Fields{}).Info("添加自动push任务成功")
		} else {
			logrus.WithFields(logrus.Fields{"err": err}).Error("添加自动push任务失败")
		}
	} else {
		logrus.WithFields(logrus.Fields{}).Info("自动push任务cron为空")
	}
	c.Start()
}

func ReceivePushSyncFile(filePath string, md5 string, reader io.Reader) (model.FileSimpleInfo, error) {
	existAndSame, err := checkFile(filePath, md5)
	if err != nil {
		return model.FileSimpleInfo{}, err
	}
	if existAndSame {
		logrus.WithFields(logrus.Fields{"filePath": filePath}).Info("文件存在且md5相同")
		return model.FileSimpleInfo{}, nil
	}
	return dao.InsertFile(filePath, reader)
}

func PushSyncFile(pushSyncHost string, token string) (int, error) {
	pushSyncHost = strings.ReplaceAll(pushSyncHost, "\\", "/")
	pushSyncHost = strings.TrimRight(pushSyncHost, "/")

	infos, err := ListAllFileCompleteInfo()
	if err != nil {
		return -1, err
	}

	client := &http.Client{Timeout: config.PullOrPushTimeout}
	jar, err := cookiejar.New(nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("http client jar创建失败")
		return -1, err
	}
	client.Jar = jar

	err = syncLogin(client, pushSyncHost, token)
	if err != nil {
		return -1, err
	}

	receivePushSynFileUrl := pushSyncHost + config.ReceivePushSyncFileUrl
	logrus.WithFields(logrus.Fields{"receivePushSynFileUrl": receivePushSynFileUrl}).Info("远端接收push文件的url")
	var failInfos []model.FileCompleteInfo
	var failErrors []error
	for _, info := range infos {
		err := pushFile(client, receivePushSynFileUrl, info)
		if err != nil {
			failInfos = append(failInfos, info)
			failErrors = append(failErrors, err)
		}
	}

	for i := range failInfos {
		logrus.WithFields(logrus.Fields{"path": failInfos[i].Path, "err": failErrors[i]}).Error("push文件失败")
	}
	return len(failInfos), nil
}

func pushFile(client *http.Client, receivePushSyncFileUrl string, info model.FileCompleteInfo) error {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	go func() {
		defer pipeWriter.Close()
		defer writer.Close()

		err := writer.WriteField("filePath", info.Path)
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err}).Error("写入表单参数filePath失败")
			return
		}
		err = writer.WriteField("md5", info.Md5)
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err}).Error("写入表单参数md5失败")
			return
		}

		formFile, err := writer.CreateFormFile("file", info.Name)
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err}).Error("创建文件表单失败")
			return
		}
		dao.ReadFileWithWriter(info.Path, formFile)
	}()

	contentType := writer.FormDataContentType()
	response, err := client.Post(receivePushSyncFileUrl, contentType, pipeReader)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("创建push文件请求失败")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("push文件http状态码异常")
		return fmt.Errorf("push文件http状态码异常")
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("push文件http读取响应异常")
		return err
	}
	logrus.WithFields(logrus.Fields{"data": string(data)}).Info("push文件请求结果")

	var pushResult struct {
		Code    int32       `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(data, &pushResult)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("push文件http响应反序列化失败")
		return err
	}
	if pushResult.Code != config.SuccessCode {
		logrus.WithFields(logrus.Fields{"pushResult": pushResult, "fileOrFolderInfo": info}).Error("push文件失败")
		return fmt.Errorf("push文件失败: %+v", pushResult)
	}

	logrus.WithFields(logrus.Fields{"pushResult": pushResult, "fileOrFolderInfo": info}).Info("push文件成功")
	return nil
}

func PullSyncFile(pullSyncHost string, token string) (int, error) {
	pullSyncHost = strings.ReplaceAll(pullSyncHost, "\\", "/")
	pullSyncHost = strings.TrimRight(pullSyncHost, "/")

	client := &http.Client{Timeout: config.PullOrPushTimeout}
	jar, err := cookiejar.New(nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("http client jar创建失败")
		return -1, err
	}
	client.Jar = jar

	err = syncLogin(client, pullSyncHost, token)
	if err != nil {
		return -1, err
	}

	allFileInfoUrl := pullSyncHost + config.ListAllFileSimpleInfoUrl
	logrus.WithFields(logrus.Fields{"allFileInfoUrl": allFileInfoUrl}).Info("获取全部文件信息Url")
	response, err := client.Get(allFileInfoUrl)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("获取全部文件信息失败")
		return -1, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("获取全部文件信息http状态码异常")
		return -1, fmt.Errorf("获取全部文件信息http状态码异常")
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("读取全部文件信息http状态码异常")
		return -1, err
	}
	logrus.WithFields(logrus.Fields{"data": string(data)}).Info("获取全部文件信息请求结果")

	var allFileInfoResult struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    []model.FileSimpleInfo `json:"data"`
	}
	err = json.Unmarshal(data, &allFileInfoResult)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("反序列化全部文件信息失败")
		return -1, err
	}

	if allFileInfoResult.Code != config.SuccessCode {
		logrus.WithFields(logrus.Fields{"allFileInfoResult": allFileInfoResult}).Error("获取全部文件信息失败")
		return -1, fmt.Errorf("获取全部文件信息失败")
	}
	logrus.WithFields(logrus.Fields{"allFileInfoResult": allFileInfoResult}).Info("获取全部文件信息成功")

	getFileCompleteInfoUrl := pullSyncHost + config.GetFileCompleteInfoUrl
	logrus.WithFields(logrus.Fields{"getFileCompleteInfoUrl": getFileCompleteInfoUrl}).Info("获取文件完整信息Url")

	var failInfos []model.FileSimpleInfo
	var failErrors []error
	for _, info := range allFileInfoResult.Data {
		pullFileUrl := pullSyncHost + info.Url
		logrus.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl}).Info("文件下载url")
		err := pullFile(client, info, getFileCompleteInfoUrl, pullFileUrl)
		if err != nil {
			failInfos = append(failInfos, info)
			failErrors = append(failErrors, err)
		} else {
			logrus.WithFields(logrus.Fields{"path": info.Path}).Info("文件下载成功")
		}
	}

	for i := range failInfos {
		logrus.WithFields(logrus.Fields{"path": failInfos[i].Path, "err": failErrors[i]}).Error("下载文件失败")
	}
	return len(failInfos), nil
}

func pullFile(client *http.Client, info model.FileSimpleInfo, getFileCompleteInfoUrl string, pullFileUrl string) error {
	filePath := info.Path
	existAndIsFolder, err := dao.ExistAndIsFolder(filePath)
	if err != nil {
		return err
	}
	if existAndIsFolder {
		logrus.WithFields(logrus.Fields{"filePath": filePath}).Error("下载文件的路径是文件夹")
		return fmt.Errorf("下载文件的路径是文件夹")
	}

	request, err := http.NewRequest("GET", getFileCompleteInfoUrl, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("创建http请求对象失败")
		return err
	}
	query := request.URL.Query()
	query.Add("fileOrFolderPath", info.Path)
	request.URL.RawQuery = query.Encode()

	response, err := client.Do(request)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("获取文件完整信息失败")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("获取文件完整信息http状态码异常")
		return fmt.Errorf("获取文件完整信息http状态码异常")
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("获取文件完整信息读取响应失败")
		return err
	}
	logrus.WithFields(logrus.Fields{"data": string(data)}).Info("获取文件完整信息")

	var completeInfoResult struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    model.FileCompleteInfo `json:"data"`
	}
	err = json.Unmarshal(data, &completeInfoResult)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("反序列化文件完整信息失败")
		return err
	}
	completeInfo := completeInfoResult.Data

	existAndSame, err := checkFile(filePath, completeInfo.Md5)
	if err != nil {
		return err
	}
	if existAndSame {
		logrus.WithFields(logrus.Fields{"filePath": filePath}).Info("文件存在且md5相同")
		return nil
	}

	response, err = http.Get(pullFileUrl)
	if err != nil {
		logrus.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl, "err": err}).Error("文件下载请求失败")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl, "StatusCode": response.StatusCode}).Error("文件下载请求状态码异常")
		return fmt.Errorf("文件下载请求状态码异常")
	}

	_, err = dao.InsertFile(filePath, response.Body)
	if err != nil {
		return err
	}

	localInfo, err := GetFileCompleteInfo(filePath)
	if err != nil {
		return err
	}
	if localInfo.Md5 != completeInfo.Md5 {
		logrus.WithFields(logrus.Fields{"filePath": filePath, "localMd5": localInfo.Md5, "remoteMd5": completeInfo.Md5}).Info("文件下载了，但MD5不匹配")
		return fmt.Errorf("文件下载了，但MD5不匹配")
	}

	return nil
}

func syncLogin(client *http.Client, syncUrl string, token string) error {
	loginUrl := syncUrl + config.LoginUrl
	logrus.WithFields(logrus.Fields{"loginUrl": loginUrl}).Info("登录远程端url")

	postValues := url.Values{}
	postValues.Add("token", token)

	response, err := client.PostForm(loginUrl, postValues)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("登录http请求异常")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("登录http状态码异常")
		return fmt.Errorf("登录http状态码异常: %+v", response.StatusCode)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("登录http请求读取异常")
		return err
	}
	logrus.WithFields(logrus.Fields{"loginData": string(data)}).Info("登录请求结果")

	var loginResult struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(data, &loginResult)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("登录http请求反序列化异常")
		return err
	}

	if loginResult.Code != config.SuccessCode {
		logrus.WithFields(logrus.Fields{"loginResult": loginResult}).Error("登录失败")
		return fmt.Errorf("登录失败: %+v", loginResult)
	}

	logrus.WithFields(logrus.Fields{"loginResult": loginResult}).Info("登录成功")
	return nil
}

//检查文件是否存在且md5相同，true:存在且md5相同
func checkFile(filePath string, md5 string) (bool, error) {
	existAndIsFile, err := dao.ExistAndIsFile(filePath)
	if err != nil {
		return false, err
	}

	if !existAndIsFile {
		logrus.WithFields(logrus.Fields{"filePath": filePath}).Info("所检查文件不存在")
		return false, nil
	}

	info, err := GetFileCompleteInfo(filePath)
	if err != nil {
		return false, err
	}

	return info.Md5 == md5, nil
}
