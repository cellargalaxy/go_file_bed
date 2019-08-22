package service

import (
	"../cache"
	"../config"
	"../dao"
	"../model"
	"../utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strings"
	"time"
)

const FileUrl = "/file"
const LoginUrl = "/login"

const UploadFileUrl = "/admin/uploadFile"
const UploadUrlUrl = "/admin/uploadUrl"
const RemoveFileUrl = "/admin/removeFile"
const ClearAllCacheUrl = "/admin/clearAllCache"
const ListFileOrFolderInfoUrl = "/admin/listFileOrFolderInfo"
const ListAllFileInfoUrl = "/admin/listAllFileInfo"

var log = logrus.New()
var fileBedPath = config.GetConfig().FileBedPath

func AddFile(sort string, filename string, reader io.Reader) (model.FileOrFolderInfo, error) {
	filePath := path.Join(fileBedPath, sort, time.Now().Format("20060102"), filename)
	filePath = utils.ClearPath(filePath)
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("添加文件路径")

	if !strings.HasPrefix(filePath, fileBedPath) {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("添加文件路径不在指定路径下")
		return model.FileOrFolderInfo{}, errors.New(fmt.Sprintf("添加文件路径不在指定路径下: %v", filePath))
	}

	err := dao.InsertFile(filePath, reader)
	if err != nil {
		return model.FileOrFolderInfo{}, err
	}
	count, size, fileInfo, err := utils.GetFileOrFolderInfo(filePath)
	if err != nil {
		return model.FileOrFolderInfo{}, err
	}
	md5, err := utils.SumFileMd5(filePath)
	if err != nil {
		return model.FileOrFolderInfo{}, err
	}
	cache.DeleteFileOrFolderInfo(filePath)
	filePath = utils.ClearPath(strings.Replace(filePath, fileBedPath, "", 1))
	return model.FileOrFolderInfo{filePath, fileInfo.Name(), !fileInfo.IsDir(), count, size, md5, createUrl(filePath)}, nil
}

func AddUrl(sort string, url string) (model.FileOrFolderInfo, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(logrus.Fields{"url": url}).Error("文件下载请求创建失败")
		return model.FileOrFolderInfo{}, err
	}
	fileUrl := fmt.Sprintf("%v://%v%v", request.URL.Scheme, request.URL.Host, request.URL.Path)
	filename := utils.Url2Path(fileUrl)
	log.WithFields(logrus.Fields{"fileUrl": fileUrl, "filename": filename}).Info("文件下载的文件名")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.WithFields(logrus.Fields{"url": url}).Error("文件下载失败")
		return model.FileOrFolderInfo{}, err
	}
	defer response.Body.Close()

	return AddFile(sort, filename, response.Body)
}

func RemoveFile(filePath string) (model.FileOrFolderInfo, error) {
	filePath = path.Join(fileBedPath, filePath)
	filePath = utils.ClearPath(filePath)
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件路径")

	if !strings.HasPrefix(filePath, fileBedPath) {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("删除文件路径不在指定路径下")
		return model.FileOrFolderInfo{}, errors.New(fmt.Sprintf("删除文件路径不在指定路径下: %v", filePath))
	}

	existAndIsFile, _ := utils.ExistAndIsFile(filePath)
	if !existAndIsFile {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("删除文件不存在或者不是文件")
		return model.FileOrFolderInfo{}, errors.New(fmt.Sprintf("删除文件不存在或者不是文件: %v", filePath))
	}

	count, size, fileInfo, err := utils.GetFileOrFolderInfo(filePath)
	if err != nil {
		return model.FileOrFolderInfo{}, err
	}
	md5, err := utils.SumFileMd5(filePath)
	if err != nil {
		return model.FileOrFolderInfo{}, err
	}
	err = dao.DeleteFile(filePath)
	if err != nil {
		return model.FileOrFolderInfo{}, err
	}
	cache.DeleteFileOrFolderInfo(filePath)
	filePath = utils.ClearPath(strings.Replace(filePath, fileBedPath, "", 1))
	return model.FileOrFolderInfo{filePath, fileInfo.Name(), !fileInfo.IsDir(), count, size, md5, createUrl(filePath)}, nil
}

func ListFileOrFolderInfo(fileOrFolderPath string) ([]model.FileOrFolderInfo, error) {
	fileOrFolderPath = path.Join(fileBedPath, fileOrFolderPath)
	fileOrFolderPath = utils.ClearPath(fileOrFolderPath)
	log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("查询文件路径")

	if !strings.HasPrefix(fileOrFolderPath, fileBedPath) {
		log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Error("查询文件路径不在指定路径下")
		return nil, errors.New(fmt.Sprintf("查询文件路径不在指定路径下: %v", fileOrFolderPath))
	}

	fileOrFolderInfos := cache.SelectListFileOrFolderInfo(fileOrFolderPath)
	if fileOrFolderInfos != nil {
		return fileOrFolderInfos, nil
	}

	fileOrFolderInfos, err := dao.SelectFileOrFolder(fileOrFolderPath)
	if err != nil {
		return nil, err
	}

	for i := range fileOrFolderInfos {
		if fileOrFolderInfos[i].IsFile {
			md5, err := utils.SumFileMd5(fileOrFolderInfos[i].Path)
			if err != nil {
				return nil, err
			}
			fileOrFolderInfos[i].Md5 = md5
		}
		fileOrFolderInfos[i].Path = utils.ClearPath(strings.Replace(fileOrFolderInfos[i].Path, fileBedPath, "", 1))
		fileOrFolderInfos[i].Url = createUrl(fileOrFolderInfos[i].Path)
	}

	cache.InsertListFileOrFolderInfo(fileOrFolderPath, fileOrFolderInfos)
	return fileOrFolderInfos, err
}

func ListAllFileInfo() ([]model.FileOrFolderInfo, error) {
	fileInfos := cache.SelectListAllFileInfo()
	if fileInfos != nil {
		return fileInfos, nil
	}

	fileInfos, err := dao.SelectAllFile(fileBedPath)
	if err != nil {
		return nil, err
	}

	for i := range fileInfos {
		if fileInfos[i].IsFile {
			md5, err := utils.SumFileMd5(fileInfos[i].Path)
			if err != nil {
				return nil, err
			}
			fileInfos[i].Md5 = md5
		}
		fileInfos[i].Path = utils.ClearPath(strings.Replace(fileInfos[i].Path, fileBedPath, "", 1))
		fileInfos[i].Url = createUrl(fileInfos[i].Path)
	}

	cache.InsertListAllFileInfo(fileInfos)
	return fileInfos, nil
}

func ClearAllCache() error {
	return cache.ClearAllCache()
}

func SynFile() error {
	client := &http.Client{Timeout: time.Hour}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	client.Jar = jar

	loginUrl := config.GetConfig().SynUrl + LoginUrl
	log.WithFields(logrus.Fields{"loginUrl": loginUrl}).Info("登录远程端Url")
	postValues := url.Values{}
	postValues.Add("token", config.GetConfig().Token)
	response, err := client.PostForm(loginUrl, postValues)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("登录http状态码异常")
		return errors.New(fmt.Sprintf("登录http状态码异常: %v", response.StatusCode))
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	log.WithFields(logrus.Fields{"loginData": string(data)}).Info("登录请求结果")
	var loginResult struct {
		Code    int32       `json:"code"`
		Massage string      `json:"massage"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(data, &loginResult)
	if err != nil {
		return err
	}
	if loginResult.Code != model.SuccessCode {
		log.WithFields(logrus.Fields{"loginResult": loginResult}).Error("登录失败")
		return errors.New(fmt.Sprintf("登录失败: %v", loginResult))
	}
	log.WithFields(logrus.Fields{"loginResult": loginResult}).Info("登录成功")

	allFileInfoUrl := config.GetConfig().SynUrl + ListAllFileInfoUrl
	log.WithFields(logrus.Fields{"allFileInfoUrl": allFileInfoUrl}).Info("获取全部文件信息Url")
	response, err = client.Get(allFileInfoUrl)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("获取全部文件信息http状态码异常")
		return errors.New(fmt.Sprintf("获取全部文件信息http状态码异常: %v", response.StatusCode))
	}
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	log.WithFields(logrus.Fields{"loginData": string(data)}).Info("获取全部文件信息请求结果")
	var allFileInfoResult struct {
		Code    int32                    `json:"code"`
		Massage string                   `json:"massage"`
		Data    []model.FileOrFolderInfo `json:"data"`
	}
	err = json.Unmarshal(data, &allFileInfoResult)
	if err != nil {
		return err
	}
	if allFileInfoResult.Code != model.SuccessCode {
		log.WithFields(logrus.Fields{"allFileInfoResult": allFileInfoResult}).Error("获取全部文件信息失败")
		return errors.New(fmt.Sprintf("获取全部文件信息失败: %v", allFileInfoResult))
	}
	log.WithFields(logrus.Fields{"allFileInfoResult": allFileInfoResult}).Info("获取全部文件信息成功")
	fileOrFolderInfos := []model.FileOrFolderInfo{}
	for i := range allFileInfoResult.Data {
		isDownload, err := downloadFile(allFileInfoResult.Data[i])
		if isDownload && err == nil {
			fileOrFolderInfos = append(fileOrFolderInfos, allFileInfoResult.Data[i])
		}
		if err == nil {
			log.WithFields(logrus.Fields{"path": allFileInfoResult.Data[i].Path}).Info("文件下载成功")
		} else {
			log.WithFields(logrus.Fields{"err": err}).Error("文件下载失败")
		}
	}
	for i := range fileOrFolderInfos {
		log.WithFields(logrus.Fields{"path": fileOrFolderInfos[i].Path}).Info("下载了文件")
	}
	return nil
}

func downloadFile(fileOrFolderInfo model.FileOrFolderInfo) (bool, error) {
	filePath := path.Join(fileBedPath, fileOrFolderInfo.Path)
	filePath = utils.ClearPath(filePath)
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("添加文件路径")

	existAndIsFile, _ := utils.ExistAndIsFile(filePath)
	if existAndIsFile {
		md5, _ := utils.SumFileMd5(filePath)
		if fileOrFolderInfo.Md5 == md5 {
			log.WithFields(logrus.Fields{"filePath": filePath}).Info("文件已存在且MD5匹配，跳过下载")
			return false, nil
		}
	}

	if !strings.HasPrefix(filePath, fileBedPath) {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("添加文件路径不在指定路径下")
		return false, errors.New(fmt.Sprintf("添加文件路径不在指定路径下: %v", filePath))
	}

	fileUrl := config.GetConfig().SynUrl + fileOrFolderInfo.Url
	response, err := http.Get(fileUrl)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("文件下载请求失败")
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"filePath": filePath, "StatusCode": response.StatusCode}).Error("文件下载请求状态码异常")
		return false, errors.New(fmt.Sprintf("文件下载请求状态码异常: %v", response.StatusCode))
	}

	return true, dao.InsertFile(filePath, response.Body)
}

func createUrl(filePath string) string {
	return FileUrl + filePath
}
