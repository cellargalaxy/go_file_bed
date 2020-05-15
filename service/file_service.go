package service

import (
	"bytes"
	"fmt"
	"github.com/cellargalaxy/go-file-bed/config"
	"github.com/cellargalaxy/go-file-bed/dao"
	"github.com/cellargalaxy/go-file-bed/model"
	"github.com/cellargalaxy/go-file-bed/utils"
	"github.com/disintegration/imaging"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"io"
	"path"
)

var lastFileInfos []model.FileSimpleInfo
var lastInfoLock = make(chan int, 1)

func AddUrl(filePath string, url string) (model.FileSimpleInfo, error) {
	request := gorequest.New()
	response, _, errs := request.Get(url).
		Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36").
		Timeout(config.Timeout).End()

	logrus.WithFields(logrus.Fields{"url": url, "errs": errs}).Info("url下载请求")
	if errs != nil && len(errs) > 0 {
		return model.FileSimpleInfo{}, fmt.Errorf("url下载请求失败: %+v", errs)
	}
	defer response.Body.Close()

	logrus.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Info("url下载请求")
	if response.StatusCode != 200 {
		return model.FileSimpleInfo{}, fmt.Errorf("url下载请求响应码异常: %+v", response.StatusCode)
	}

	return AddFile(filePath, response.Body)
}

func AddFile(filePath string, reader io.Reader) (model.FileSimpleInfo, error) {
	fileExt := path.Ext(filePath)
	logrus.WithFields(logrus.Fields{"fileExt": fileExt}).Info("文件拓展名")
	format, err := imaging.FormatFromExtension(fileExt)
	logrus.WithFields(logrus.Fields{"format": format, "err": err}).Info("解析图片拓展名")

	if err == nil {
		buffer := &bytes.Buffer{}
		_, err := io.Copy(buffer, reader)
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err}).Error("读取文件数据异常")
			return model.FileSimpleInfo{}, err
		}
		imageBuffer, err := CompressionImage(buffer)
		if err != nil {
			imageBuffer = buffer
		}
		reader = imageBuffer

		filePath = AddImageExtension(filePath)
	}

	info, err := dao.InsertFile(filePath, reader)
	info = initFileSimpleInfo(info)
	addLastFileInfo(info)
	return info, err
}

func RemoveFile(filePath string) (model.FileSimpleInfo, error) {
	info, err := dao.DeleteFile(filePath)
	info = initFileSimpleInfo(info)
	return info, err
}

func GetFileCompleteInfo(fileOrFolderPath string) (model.FileCompleteInfo, error) {
	info, err := dao.GetFileCompleteInfo(fileOrFolderPath)
	info = initFileCompleteInfo(info)
	return info, err
}

func ListLastFileInfos() ([]model.FileSimpleInfo, error) {
	getLastInfoLock()
	defer releaseLastInfoLock()
	return lastFileInfos, nil
}

func ListFolderInfo(folderPath string) ([]model.FileSimpleInfo, error) {
	infos, err := dao.ListFileSimpleInfo(folderPath)
	infos = initFileSimpleInfos(infos)
	return infos, err
}

func ListAllFileSimpleInfo() ([]model.FileSimpleInfo, error) {
	infos, err := dao.ListAllFileSimpleInfo("")
	infos = initFileSimpleInfos(infos)
	return infos, err
}

func ListAllFileCompleteInfo() ([]model.FileCompleteInfo, error) {
	infos, err := dao.ListAllFileCompleteInfo("")
	infos = initFileCompleteInfos(infos)
	return infos, err
}

func addLastFileInfo(info model.FileSimpleInfo) {
	getLastInfoLock()
	defer releaseLastInfoLock()
	if len(lastFileInfos) < config.LastFileInfoCount {
		lastFileInfos = append([]model.FileSimpleInfo{info}, lastFileInfos...)
		return
	}
	for i := len(lastFileInfos) - 1; i > 0; i-- {
		lastFileInfos[i] = lastFileInfos[i-1]
	}
	lastFileInfos[0] = info
}

func getLastInfoLock() {
	lastInfoLock <- 1
}

func releaseLastInfoLock() {
	<-lastInfoLock
}

func initFileCompleteInfos(infos []model.FileCompleteInfo) []model.FileCompleteInfo {
	if infos == nil {
		return []model.FileCompleteInfo{}
	}
	for i := range infos {
		infos[i] = initFileCompleteInfo(infos[i])
	}
	return infos
}

func initFileCompleteInfo(info model.FileCompleteInfo) model.FileCompleteInfo {
	info.Url = createUrl(info.Path)
	return info
}

func initFileSimpleInfos(infos []model.FileSimpleInfo) []model.FileSimpleInfo {
	if infos == nil {
		return []model.FileSimpleInfo{}
	}
	for i := range infos {
		infos[i] = initFileSimpleInfo(infos[i])
	}
	return infos
}

func initFileSimpleInfo(info model.FileSimpleInfo) model.FileSimpleInfo {
	info.Url = createUrl(info.Path)
	return info
}

func createUrl(filePath string) string {
	return utils.ClearPath(path.Join(config.FileUrl, filePath))
}
