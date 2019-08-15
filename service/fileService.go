package service

import (
	"../config"
	"../dao"
	"../model"
	"../utils"
	"encoding/json"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const FileUrl = "/file"
const LoginUrl = "/login"

const UploadFileUrl = "/admin/uploadFile"
const RemoveFileUrl = "/admin/removeFile"
const ListFileOrFolderInfoUrl = "/admin/listFileOrFolderInfo"
const ListAllFileInfoUrl = "/admin/listAllFileInfo"

var log = logrus.New()
var fileBedPath = config.GetConfig().FileBedPath
var cache sync.Map
var selectAllFileKey = uuid.Must(uuid.NewV4()).String()

func AddFile(sort string, filename string, reader io.Reader) error {
	filePath := path.Join(fileBedPath, sort, time.Now().Format("20060102"), filename)
	filePath = utils.ClearPath(filePath)
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("添加文件路径")

	if !strings.HasPrefix(filePath, fileBedPath) {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("添加文件路径不在指定路径下")
		return errors.New(fmt.Sprintf("添加文件路径不在指定路径下: %v", filePath))
	}

	err := dao.InsertFile(filePath, reader)
	if err == nil {
		cache.Delete(selectAllFileKey)
		cache.Range(func(fileOrFolderPath, _ interface{}) bool {
			if strings.HasPrefix(filePath, fileOrFolderPath.(string)) {
				log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("删除缓存")
				cache.Delete(fileOrFolderPath)
			}
			return true
		})
	}
	return err
}

func RemoveFile(filePath string) error {
	filePath = path.Join(fileBedPath, filePath)
	filePath = utils.ClearPath(filePath)
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件路径")

	if !strings.HasPrefix(filePath, fileBedPath) {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("删除文件路径不在指定路径下")
		return errors.New(fmt.Sprintf("删除文件路径不在指定路径下: %v", filePath))
	}

	err := dao.DeleteFile(filePath)
	if err == nil {
		cache.Delete(selectAllFileKey)
		cache.Range(func(fileOrFolderPath, _ interface{}) bool {
			if strings.HasPrefix(filePath, fmt.Sprintf("%v", fileOrFolderPath)) {
				log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("删除缓存")
				cache.Delete(fileOrFolderPath)
			}
			return true
		})
	}
	return err
}

func ListFileOrFolderInfo(fileOrFolderPath string) ([]model.FileOrFolderInfo, error) {
	fileOrFolderPath = path.Join(fileBedPath, fileOrFolderPath)
	fileOrFolderPath = utils.ClearPath(fileOrFolderPath)
	log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("查询文件路径")

	if !strings.HasPrefix(fileOrFolderPath, fileBedPath) {
		log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Error("查询文件路径不在指定路径下")
		return nil, errors.New(fmt.Sprintf("查询文件路径不在指定路径下: %v", fileOrFolderPath))
	}

	fileOrFolderInfos, _ := cache.Load(fileOrFolderPath)
	if fileOrFolderInfos != nil {
		return fileOrFolderInfos.([]model.FileOrFolderInfo), nil
	}

	fileOrFolderInfos, err := dao.SelectFileOrFolder(fileOrFolderPath)
	if err == nil {
		cache.Store(fileOrFolderPath, fileOrFolderInfos)
	}

	infos := fileOrFolderInfos.([]model.FileOrFolderInfo)
	if fileOrFolderInfos != nil && err == nil {
		for i := range infos {
			infos[i].Path = utils.ClearPath(strings.Replace(infos[i].Path, fileBedPath, "", 1))
		}
	}
	return infos, err
}

func ListAllFileInfo() ([]model.FileOrFolderInfo, error) {
	fileInfos, _ := cache.Load(selectAllFileKey)
	if fileInfos != nil {
		return fileInfos.([]model.FileOrFolderInfo), nil
	}

	fileInfos, err := dao.SelectAllFile(fileBedPath)
	if err == nil {
		cache.Store(selectAllFileKey, fileInfos)
	}

	infos := fileInfos.([]model.FileOrFolderInfo)
	if fileInfos != nil && err == nil {
		for i := range infos {
			infos[i].Path = utils.ClearPath(strings.Replace(infos[i].Path, fileBedPath, "", 1))
		}
	}
	return infos, nil
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
	for i := range allFileInfoResult.Data {
		err := downloadFile(allFileInfoResult.Data[i])
		if err == nil {
			log.WithFields(logrus.Fields{"path": allFileInfoResult.Data[i].Path}).Info("文件下载成功")
		} else {
			log.WithFields(logrus.Fields{"err": err}).Error("文件下载失败")
		}
	}
	return nil
}

func downloadFile(fileOrFolderInfo model.FileOrFolderInfo) error {
	filePath := path.Join(fileBedPath, fileOrFolderInfo.Path)
	filePath = utils.ClearPath(filePath)
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("添加文件路径")

	existAndIsFile, _ := utils.ExistAndIsFile(filePath)
	if existAndIsFile {
		md5, _ := utils.SumFileMd5(filePath)
		if fileOrFolderInfo.Md5 == md5 {
			log.WithFields(logrus.Fields{"filePath": filePath}).Info("文件已存在且MD5匹配，跳过下载")
			return nil
		}
	}

	if !strings.HasPrefix(filePath, fileBedPath) {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("添加文件路径不在指定路径下")
		return errors.New(fmt.Sprintf("添加文件路径不在指定路径下: %v", filePath))
	}

	fileUrl := config.GetConfig().SynUrl + FileUrl + fileOrFolderInfo.Path
	response, err := http.Get(fileUrl)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("文件下载请求失败")
		return err
	}
	defer response.Body.Close()

	return dao.InsertFile(filePath, response.Body)
}
