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
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const FileUrl = "/file"
const LoginUrl = "/login"

const UploadUrlUrl = "/admin/uploadUrl"
const UploadFileUrl = "/admin/uploadFile"
const UploadFileByFilePathUrl = "/admin/uploadFileByFilePath"
const RemoveFileUrl = "/admin/removeFile"
const ListFolderInfoUrl = "/admin/listFolderInfo"
const ListAllFileInfoUrl = "/admin/listAllFileInfo"
const ReceivePushSynFileUrl = "/admin/receivePushSynFile"
const PushSynFileUrl = "/admin/pushSynFile"
const PullSynFileUrl = "/admin/pullSynFile"
const ClearAllCacheUrl = "/admin/clearAllCache"

var log = logrus.New()
var fileBedPath = config.GetConfig().FileBedPath

func init() {
	os.MkdirAll(fileBedPath, 0666)
	go func() {
		log.Info("开始索引文件")
		ListFolderInfo("")
		ListAllFileInfo()
		log.Info("成功索引文件")
	}()
}

func AddUrl(sort string, url string) ([]model.FileOrFolderInfo, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(logrus.Fields{"url": url}).Error("文件下载请求创建失败")
		return nil, err
	}
	fileUrl := fmt.Sprintf("%v://%v%v", request.URL.Scheme, request.URL.Host, request.URL.Path)
	filename := utils.Url2Path(fileUrl)
	log.WithFields(logrus.Fields{"fileUrl": fileUrl, "filename": filename}).Info("文件下载的文件名")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.WithFields(logrus.Fields{"url": url}).Error("文件下载请求失败")
		return nil, err
	}
	defer response.Body.Close()

	return AddFile(sort, filename, response.Body)
}

func AddFile(sort string, filename string, reader io.Reader) ([]model.FileOrFolderInfo, error) {
	sort = strings.Replace(sort, " ", "", -1)
	filePath := path.Join(sort, time.Now().Format("20060102"), filename)
	return AddFileByFilePath(filePath, reader)
}

func AddFileByFilePath(filePath string, reader io.Reader) ([]model.FileOrFolderInfo, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("添加文件路径为空")
		return nil, errors.New(fmt.Sprintf("添加文件路径为空: %v", filePath))
	}

	bedFilePath := utils.ClearPath(path.Join(fileBedPath, filePath))
	log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Info("添加文件路径")

	if !strings.HasPrefix(bedFilePath, fileBedPath) {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("添加文件路径不在指定路径下")
		return nil, errors.New(fmt.Sprintf("添加文件路径不在指定路径下: %v", filePath))
	}

	err := dao.InsertFile(bedFilePath, reader)
	if err != nil {
		return nil, err
	}

	existAndIsFile, fileInfo := utils.ExistAndIsFile(bedFilePath)
	if !existAndIsFile {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("保存文件了，但文件不存在")
		return nil, errors.New(fmt.Sprintf("保存文件了，但文件不存在: %v", filePath))
	}
	fileOrFolderInfo := model.FileOrFolderInfo{bedFilePath, fileInfo.Name(), existAndIsFile, 1, fileInfo.Size(), "", ""}
	fileOrFolderInfo = initFileOrFolderInfo(fileOrFolderInfo)
	fileOrFolderInfos := []model.FileOrFolderInfo{fileOrFolderInfo}
	folder, _ := path.Split(filePath)
	folder = path.Clean(folder)
	cache.DeleteCache(folder)
	cache.InsertFileCache(filePath, fileOrFolderInfos)
	cache.InsertFolderCache(filePath, fileOrFolderInfos)

	return fileOrFolderInfos, nil
}

func RemoveFile(filePath string) ([]model.FileOrFolderInfo, error) {
	bedFilePath := utils.ClearPath(path.Join(fileBedPath, filePath))
	log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Info("删除文件路径")

	if !strings.HasPrefix(bedFilePath, fileBedPath) {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("删除文件路径不在指定路径下")
		return nil, errors.New(fmt.Sprintf("删除文件路径不在指定路径下: %v", filePath))
	}

	existAndIsFile, fileInfo := utils.ExistAndIsFile(bedFilePath)
	if !existAndIsFile {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("删除文件不存在或者不是文件")
		return nil, errors.New(fmt.Sprintf("删除文件不存在或者不是文件: %v", filePath))
	}

	fileInfos := cache.SelectFileCache(filePath)
	folder, _ := path.Split(filePath)
	folder = path.Clean(folder)
	cache.DeleteCache(folder)
	if fileInfos == nil {
		fileInfo := model.FileOrFolderInfo{bedFilePath, fileInfo.Name(), existAndIsFile, 1, fileInfo.Size(), "", ""}
		fileInfo = initFileOrFolderInfo(fileInfo)
	}

	err := dao.DeleteFile(bedFilePath)
	return fileInfos, err
}

func ListFolderInfo(folderPath string) ([]model.FileOrFolderInfo, error) {
	bedFolderPath := utils.ClearPath(path.Join(fileBedPath, folderPath))
	log.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath}).Info("查询文件夹路径")

	if !strings.HasPrefix(bedFolderPath, fileBedPath) {
		log.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath}).Error("查询文件夹路径不在指定路径下")
		return nil, errors.New(fmt.Sprintf("查询文件夹路径不在指定路径下: %v", bedFolderPath))
	}

	folderInfos := cache.SelectFolderCache(folderPath)
	if folderInfos != nil {
		return folderInfos, nil
	}

	existAndIsFile, fileInfo := utils.ExistAndIsFile(bedFolderPath)
	if existAndIsFile {
		folderInfos := cache.SelectFileCache(folderPath)
		if folderInfos != nil {
			return folderInfos, nil
		}

		folderInfo := model.FileOrFolderInfo{bedFolderPath, fileInfo.Name(), existAndIsFile, 1, fileInfo.Size(), "", ""}
		folderInfo = initFileOrFolderInfo(folderInfo)
		folderInfos = []model.FileOrFolderInfo{folderInfo}
		cache.InsertFileCache(folderPath, folderInfos)
		cache.InsertFolderCache(folderPath, folderInfos)
		return folderInfos, nil
	}

	existAndIsFolder, _ := utils.ExistAndIsFolder(bedFolderPath)
	if !existAndIsFolder {
		log.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath}).Error("查询文件夹路径不存在")
		return nil, errors.New(fmt.Sprintf("查询文件夹路径不存在: %v", folderPath))
	}

	files, err := ioutil.ReadDir(bedFolderPath)
	if err != nil {
		return nil, err
	}

	folderInfos = []model.FileOrFolderInfo{}
	for i := range files {
		childFolderPath := path.Join(folderPath, files[i].Name())
		bedChildFolderPath := utils.ClearPath(path.Join(fileBedPath, childFolderPath))
		childFolderInfos, err := ListFolderInfo(childFolderPath)
		if err != nil {
			log.WithFields(logrus.Fields{"bedChildFolderPath": bedChildFolderPath, "err": err}).Error("查询文件夹路径下的信息失败")
			continue
		}

		//如果是文件
		existAndIsFile, _ := utils.ExistAndIsFile(bedChildFolderPath)
		if existAndIsFile {
			folderInfos = append(folderInfos, childFolderInfos[0])
			continue
		}
		//如果是文件夹，将文件夹里面的文件信息相加，即这个文件夹的信息
		folderInfo := model.FileOrFolderInfo{bedChildFolderPath, files[i].Name(), false, 0, 0, "", ""}
		for j := range childFolderInfos {
			folderInfo.FileCount = folderInfo.FileCount + childFolderInfos[j].FileCount
			folderInfo.FileSize = folderInfo.FileSize + childFolderInfos[j].FileSize
		}
		folderInfo = initFileOrFolderInfo(folderInfo)
		folderInfos = append(folderInfos, folderInfo)
	}
	cache.InsertFolderCache(folderPath, folderInfos)

	return folderInfos, err
}

func ListAllFileInfo() ([]model.FileOrFolderInfo, error) {
	return listFileInfo("")
}

//查询某个文件夹下的全部文件的信息
func listFileInfo(folderPath string) ([]model.FileOrFolderInfo, error) {
	bedFolderPath := utils.ClearPath(path.Join(fileBedPath, folderPath))
	log.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath}).Info("查询文件信息路径")

	if !strings.HasPrefix(bedFolderPath, fileBedPath) {
		log.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath}).Error("查询文件信息路径不在指定路径下")
		return nil, errors.New(fmt.Sprintf("查询文件信息路径不在指定路径下: %v", folderPath))
	}

	fileInfos := cache.SelectFileCache(folderPath)
	if fileInfos != nil {
		return fileInfos, nil
	}

	existAndIsFile, fileInfo := utils.ExistAndIsFile(bedFolderPath)
	if existAndIsFile {
		fileInfo := model.FileOrFolderInfo{bedFolderPath, fileInfo.Name(), existAndIsFile, 1, fileInfo.Size(), "", ""}
		fileInfo = initFileOrFolderInfo(fileInfo)
		fileInfos = []model.FileOrFolderInfo{fileInfo}
		cache.InsertFileCache(folderPath, fileInfos)
		cache.InsertFolderCache(folderPath, fileInfos)
		return fileInfos, nil
	}

	existAndIsFolder, _ := utils.ExistAndIsFolder(bedFolderPath)
	if !existAndIsFolder {
		log.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath}).Error("所查询路径不存在")
		return nil, errors.New(fmt.Sprintf("所查询路径不存在: %v", folderPath))
	}

	files, err := ioutil.ReadDir(bedFolderPath)
	if err != nil {
		return nil, err
	}

	fileInfos = []model.FileOrFolderInfo{}
	for i := range files {
		childFolderPath := path.Join(folderPath, files[i].Name())
		bedChildFolderPath := utils.ClearPath(path.Join(fileBedPath, childFolderPath))
		childFolderInfos, err := listFileInfo(childFolderPath)
		if err != nil {
			log.WithFields(logrus.Fields{"bedChildFolderPath": bedChildFolderPath, "err": err}).Error("查询文件的信息失败")
			continue
		}
		fileInfos = append(fileInfos, childFolderInfos...)
	}
	cache.InsertFileCache(folderPath, fileInfos)

	return fileInfos, nil
}

func ReceivePushSynFile(filePath string, md5 string, reader io.Reader) error {
	exist, err := checkFile(filePath, md5)
	if err != nil {
		return err
	}
	if exist {
		log.WithFields(logrus.Fields{"filePath": filePath}).Info("文件存在且md5相同")
		return nil
	}
	_, err = AddFileByFilePath(filePath, reader)
	return err
}

//检查文件是否存在且md5相同，true:存在且md5相同
func checkFile(filePath string, md5 string) (bool, error) {
	bedFilePath := utils.ClearPath(path.Join(fileBedPath, filePath))
	log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Info("所检查的文件路径")

	if !strings.HasPrefix(bedFilePath, fileBedPath) {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("所检查的文件不在指定路径下")
		return false, errors.New(fmt.Sprintf("所检查的文件不在指定路径下: %v", filePath))
	}

	existAndIsFolder, _ := utils.ExistAndIsFolder(bedFilePath)
	if existAndIsFolder {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("所检查文件是文件夹")
		return false, errors.New(fmt.Sprintf("所检查文件是文件夹: %v", filePath))
	}

	existAndIsFile, fileInfo := utils.ExistAndIsFile(bedFilePath)
	if !existAndIsFile {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Info("所检查文件不存在")
		return false, nil
	}

	fileInfos := cache.SelectFileCache(filePath)
	if fileInfos == nil {
		fileInfo := model.FileOrFolderInfo{bedFilePath, fileInfo.Name(), existAndIsFile, 1, fileInfo.Size(), "", ""}
		fileInfo = initFileOrFolderInfo(fileInfo)
		fileInfos = []model.FileOrFolderInfo{fileInfo}
		cache.InsertFileCache(filePath, fileInfos)
		cache.InsertFolderCache(filePath, fileInfos)
	}

	return fileInfos[0].Md5 == md5, nil
}

func PushSynFile(pushSynHost string, token string) (int, error) {
	fileInfos, err := ListAllFileInfo()
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("获取全部文件信息失败")
		return 0, err
	}

	client := &http.Client{Timeout: time.Hour}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("http client jar创建失败")
		return 0, err
	}
	client.Jar = jar

	err = synLogin(client, pushSynHost, token)
	if err != nil {
		return 0, err
	}

	receivePushSynFileUrl := pushSynHost + ReceivePushSynFileUrl
	log.WithFields(logrus.Fields{"receivePushSynFileUrl": receivePushSynFileUrl}).Info("远端接收push文件的url")
	var failFileInfos []model.FileOrFolderInfo
	var failFileErrors []error
	for i := range fileInfos {
		err := pushFile(client, receivePushSynFileUrl, fileInfos[i])
		if err != nil {
			failFileInfos = append(failFileInfos, fileInfos[i])
			failFileErrors = append(failFileErrors, err)
		}
	}
	for i := range failFileInfos {
		log.WithFields(logrus.Fields{"path": failFileInfos[i].Path, "err": failFileErrors[i]}).Error("push文件失败")
	}
	return len(failFileInfos), nil
}

func pushFile(client *http.Client, receivePushSynFileUrl string, fileOrFolderInfo model.FileOrFolderInfo) error {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	go func() {
		defer pipeWriter.Close()
		defer writer.Close()

		err := writer.WriteField("filePath", fileOrFolderInfo.Path)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("写入表单参数filePath失败")
			return
		}
		err = writer.WriteField("md5", fileOrFolderInfo.Md5)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("写入表单参数md5失败")
			return
		}

		formFile, err := writer.CreateFormFile("file", fileOrFolderInfo.Name)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("创建文件表单失败")
			return
		}

		filePath := fileOrFolderInfo.Path
		bedFilePath := utils.ClearPath(path.Join(fileBedPath, filePath))
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Info("所push的文件路径")
		file, err := os.Open(bedFilePath)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("打开push文件失败")
			return
		}
		defer file.Close()

		_, err = io.Copy(formFile, file)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("将文件写入表单失败")
			return
		}
	}()

	contentType := writer.FormDataContentType()
	response, err := client.Post(receivePushSynFileUrl, contentType, pipeReader)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("创建push文件请求失败")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("push文件http状态码异常")
		return errors.New(fmt.Sprintf("push文件http状态码异常: %v", response.StatusCode))
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("push文件http请求读取异常")
		return err
	}
	log.WithFields(logrus.Fields{"pushData": string(data)}).Info("push文件请求结果")

	var pushResult struct {
		Code    int32       `json:"code"`
		Massage string      `json:"massage"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(data, &pushResult)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("push文件http请求反序列化异常")
		return err
	}
	if pushResult.Code != model.SuccessCode {
		log.WithFields(logrus.Fields{"pushResult": pushResult}).Error("push文件失败")
		return errors.New(fmt.Sprintf("push文件失败: %v", pushResult))
	}
	log.WithFields(logrus.Fields{"pushResult": pushResult}).Info("push文件成功")
	return nil
}

func PullSynFile(pullSynHost string, token string) (int, error) {
	client := &http.Client{Timeout: time.Hour}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("http client jar创建失败")
		return 0, err
	}
	client.Jar = jar

	err = synLogin(client, pullSynHost, token)
	if err != nil {
		return 0, err
	}

	allFileInfoUrl := pullSynHost + ListAllFileInfoUrl
	log.WithFields(logrus.Fields{"allFileInfoUrl": allFileInfoUrl}).Info("获取全部文件信息Url")
	response, err := client.Get(allFileInfoUrl)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("获取全部文件信息http状态码异常")
		return 0, errors.New(fmt.Sprintf("获取全部文件信息http状态码异常: %v", response.StatusCode))
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	log.WithFields(logrus.Fields{"loginData": string(data)}).Info("获取全部文件信息请求结果")
	var allFileInfoResult struct {
		Code    int32                    `json:"code"`
		Massage string                   `json:"massage"`
		Data    []model.FileOrFolderInfo `json:"data"`
	}
	err = json.Unmarshal(data, &allFileInfoResult)
	if err != nil {
		return 0, err
	}
	if allFileInfoResult.Code != model.SuccessCode {
		log.WithFields(logrus.Fields{"allFileInfoResult": allFileInfoResult}).Error("获取全部文件信息失败")
		return 0, errors.New(fmt.Sprintf("获取全部文件信息失败: %v", allFileInfoResult))
	}
	log.WithFields(logrus.Fields{"allFileInfoResult": allFileInfoResult}).Info("获取全部文件信息成功")
	var fileOrFolderInfos []model.FileOrFolderInfo
	var errs []error
	for i := range allFileInfoResult.Data {
		fileOrFolderInfo := allFileInfoResult.Data[i]
		pullFileUrl := pullSynHost + fileOrFolderInfo.Url
		log.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl}).Info("文件下载url")
		err := pullFile(pullFileUrl, fileOrFolderInfo)
		if err != nil {
			fileOrFolderInfos = append(fileOrFolderInfos, fileOrFolderInfo)
			errs = append(errs, err)
		} else {
			log.WithFields(logrus.Fields{"path": fileOrFolderInfo.Path}).Info("文件下载成功")
		}
	}
	for i := range fileOrFolderInfos {
		log.WithFields(logrus.Fields{"path": fileOrFolderInfos[i].Path, "err": errs[i]}).Error("下载文件失败")
	}
	return len(fileOrFolderInfos), nil
}

func pullFile(pullFileUrl string, fileOrFolderInfo model.FileOrFolderInfo) error {
	filePath := fileOrFolderInfo.Path
	bedFilePath := utils.ClearPath(path.Join(fileBedPath, filePath))
	log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Info("下载文件的路径")

	if !strings.HasPrefix(bedFilePath, fileBedPath) {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("下载文件的路径不在指定路径下")
		return errors.New(fmt.Sprintf("下载文件的路径不在指定路径下: %v", filePath))
	}

	existAndIsFolder, _ := utils.ExistAndIsFolder(bedFilePath)
	if existAndIsFolder {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("下载文件的路径是文件夹")
		return errors.New(fmt.Sprintf("下载文件的路径是文件夹: %v", filePath))
	}

	existAndIsFile, fileInfo := utils.ExistAndIsFile(bedFilePath)
	if existAndIsFile {
		fileInfos := cache.SelectFileCache(filePath)
		if fileInfos == nil {
			fileInfo := model.FileOrFolderInfo{bedFilePath, fileInfo.Name(), existAndIsFile, 1, fileInfo.Size(), "", ""}
			fileInfo = initFileOrFolderInfo(fileInfo)
			fileInfos = []model.FileOrFolderInfo{fileInfo}
			cache.InsertFileCache(filePath, fileInfos)
			cache.InsertFolderCache(filePath, fileInfos)
		}

		if fileInfos[0].Md5 == fileOrFolderInfo.Md5 {
			log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Info("文件已存在且MD5匹配，跳过下载")
			return nil
		}
	}

	response, err := http.Get(pullFileUrl)
	if err != nil {
		log.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl}).Error("文件下载请求失败")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl, "StatusCode": response.StatusCode}).Error("文件下载请求状态码异常")
		return errors.New(fmt.Sprintf("文件下载请求状态码异常: %v", response.StatusCode))
	}

	fileOrFolderInfos, err := AddFileByFilePath(filePath, response.Body)
	if err != nil {
		log.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl, "err": err}).Error("文件下载失败")
		return err
	}

	if fileOrFolderInfos[0].Md5 != fileOrFolderInfo.Md5 {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath, "localMd5": fileOrFolderInfos[0].Md5, "remoteMd5": fileOrFolderInfo.Md5}).Info("文件下载了，但MD5不匹配")
		return errors.New(fmt.Sprintf("文件下载了，但MD5不匹配: %v", filePath))
	}
	return nil
}

func synLogin(client *http.Client, synUrl string, token string) error {
	loginUrl := synUrl + LoginUrl
	log.WithFields(logrus.Fields{"loginUrl": loginUrl}).Info("登录远程端Url")
	postValues := url.Values{}
	postValues.Add("token", token)
	response, err := client.PostForm(loginUrl, postValues)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("登录http请求异常")
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("登录http状态码异常")
		return errors.New(fmt.Sprintf("登录http状态码异常: %v", response.StatusCode))
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("登录http请求读取异常")
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
		log.WithFields(logrus.Fields{"err": err}).Error("登录http请求反序列化异常")
		return err
	}
	if loginResult.Code != model.SuccessCode {
		log.WithFields(logrus.Fields{"loginResult": loginResult}).Error("登录失败")
		return errors.New(fmt.Sprintf("登录失败: %v", loginResult))
	}
	log.WithFields(logrus.Fields{"loginResult": loginResult}).Info("登录成功")
	return nil
}

func initFileOrFolderInfo(fileOrFolderInfo model.FileOrFolderInfo) model.FileOrFolderInfo {
	bedFilePath := fileOrFolderInfo.Path
	if fileOrFolderInfo.IsFile {
		md5, err := utils.SumFileMd5(bedFilePath)
		fileOrFolderInfo.Md5 = md5
		if err != nil {
			log.WithFields(logrus.Fields{"bedFilePath": bedFilePath, "err": err}).Error("计算文件md5失败")
		}
	}
	filePath := utils.ClearPath(strings.Replace(bedFilePath, fileBedPath, "", 1))
	fileOrFolderInfo.Path = filePath
	fileOrFolderInfo.Url = utils.ClearPath(path.Join(FileUrl, filePath))
	return fileOrFolderInfo
}

func ClearAllCache() {
	cache.ClearAllCache()
}
