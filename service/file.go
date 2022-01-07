package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/cellargalaxy/go_file_bed/dao"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"path"
	"sync"
)

const (
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"
)

var lastFileInfos []model.FileSimpleInfo
var lastFileInfoLock sync.Mutex

func AddUrl(ctx context.Context, filePath string, url string, raw bool) (*model.FileSimpleInfo, error) {
	if url == "" {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Error("添加链接，文件下载连接为空")
		return nil, fmt.Errorf("添加链接，文件下载连接为空")
	}

	response, err := httpClient.R().SetContext(ctx).
		SetHeader("User-Agent", userAgent).
		SetDoNotParseResponse(true).
		Get(url)

	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加链接，文件下载异常")
		return nil, fmt.Errorf("添加链接，文件下载异常")
	}
	if response == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加链接，文件下载响应为空")
		return nil, fmt.Errorf("添加链接，文件下载响应为空")
	}
	statusCode := response.StatusCode()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode}).Info("添加链接，文件下载响应")
	if statusCode != http.StatusOK {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"StatusCode": statusCode}).Error("添加链接，文件下载响应码失败")
		return nil, fmt.Errorf("添加链接，文件下载响应码失败: %+v", statusCode)
	}

	return AddFile(ctx, filePath, response.RawBody(), raw)
}

func AddFile(ctx context.Context, filePath string, reader io.Reader, raw bool) (*model.FileSimpleInfo, error) {
	fileExt := path.Ext(filePath)
	logrus.WithContext(ctx).WithFields(logrus.Fields{"fileExt": fileExt}).Info("添加文件，文件拓展名")
	format, err := imaging.FormatFromExtension(fileExt)
	logrus.WithContext(ctx).WithFields(logrus.Fields{"format": format, "err": err}).Info("添加文件，解析图片拓展名")

	if !raw && err == nil && format != imaging.GIF {
		buffer := &bytes.Buffer{}
		_, err = io.Copy(buffer, reader)
		if err != nil {
			logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加文件，读取图片数据异常")
			return nil, err
		}

		imageBuffer, err := CompressionImage(ctx, buffer)
		if err != nil {
			reader = buffer
		} else {
			reader = imageBuffer
			filePath = AddImageExtension(ctx, filePath)
		}
	}

	info, err := dao.InsertFile(ctx, filePath, reader)
	if err != nil {
		return nil, err
	}
	info = initFileSimpleInfo(ctx, info)
	addLastFileInfo(ctx, info)
	return info, err
}

func RemoveFile(ctx context.Context, filePath string) (*model.FileSimpleInfo, error) {
	info, err := dao.DeleteFile(ctx, filePath)
	if err != nil {
		return nil, err
	}
	info = initFileSimpleInfo(ctx, info)
	return info, err
}

func GetFileCompleteInfo(ctx context.Context, fileOrFolderPath string) (*model.FileCompleteInfo, error) {
	info, err := dao.SelectFileCompleteInfo(ctx, fileOrFolderPath)
	if err != nil {
		return nil, err
	}
	info = initFileCompleteInfo(ctx, info)
	return info, err
}

func ListFileSimpleInfo(ctx context.Context, folderPath string) ([]model.FileSimpleInfo, error) {
	infos, err := dao.SelectFolderSimpleInfo(ctx, folderPath)
	if err != nil {
		return nil, err
	}
	infos = initFileSimpleInfos(ctx, infos)
	return infos, err
}

func ListFileCompleteInfo(ctx context.Context, folderPath string) ([]model.FileCompleteInfo, error) {
	infos, err := dao.SelectFolderCompleteInfo(ctx, folderPath)
	if err != nil {
		return nil, err
	}
	infos = initFileCompleteInfos(ctx, infos)
	return infos, err
}

func ListLastFileInfo(ctx context.Context) ([]model.FileSimpleInfo, error) {
	return lastFileInfos, nil
}

func addLastFileInfo(ctx context.Context, info *model.FileSimpleInfo) {
	if info == nil {
		return
	}
	lastFileInfoLock.Lock()
	defer lastFileInfoLock.Unlock()

	if len(lastFileInfos) < config.Config.LastFileCount {
		lastFileInfos = append([]model.FileSimpleInfo{*info}, lastFileInfos...)
		return
	}

	infos := lastFileInfos
	for i := len(infos) - 1; i > 0; i-- {
		infos[i] = infos[i-1]
	}
	infos[0] = *info
	lastFileInfos = infos
}

func initFileCompleteInfos(ctx context.Context, infos []model.FileCompleteInfo) []model.FileCompleteInfo {
	for i := range infos {
		infos[i].Url = createUrl(ctx, infos[i].Path)
	}
	return infos
}

func initFileCompleteInfo(ctx context.Context, info *model.FileCompleteInfo) *model.FileCompleteInfo {
	if info == nil {
		return nil
	}
	info.Url = createUrl(ctx, info.Path)
	return info
}

func initFileSimpleInfos(ctx context.Context, infos []model.FileSimpleInfo) []model.FileSimpleInfo {
	for i := range infos {
		infos[i].Url = createUrl(ctx, infos[i].Path)
	}
	return infos
}

func initFileSimpleInfo(ctx context.Context, info *model.FileSimpleInfo) *model.FileSimpleInfo {
	if info == nil {
		return nil
	}
	info.Url = createUrl(ctx, info.Path)
	return info
}

func createUrl(ctx context.Context, filePath string) string {
	return util.ClearPath(ctx, path.Join(model.FileUrl, filePath))
}
