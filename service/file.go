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
	"strings"
	"sync"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36"
)

var lastFileInfos []model.FileSimpleInfo
var lastFileInfoLock sync.Mutex

func AddUrl(ctx context.Context, filePath string, url string, raw bool) (*model.FileSimpleInfo, error) {
	if url == "" {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Error("添加链接，文件下载连接为空")
		return nil, fmt.Errorf("添加链接，文件下载连接为空")
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加链接，创建http请求异常")
		return nil, fmt.Errorf("添加链接，创建http请求异常: %+v", err)
	}
	if request == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加链接，创建http请求为空")
		return nil, fmt.Errorf("添加链接，创建http请求为空")
	}
	if request.Header == nil {
		request.Header = http.Header{}
	}
	request.Header.Set("User-Agent", userAgent)

	response, err := http.DefaultClient.Do(request.WithContext(ctx))
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加链接，http请求异常")
		return nil, fmt.Errorf("添加链接，http请求异常: %+v", err)
	}
	if response == nil || response.Body == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加链接，http响应为空")
		return nil, fmt.Errorf("添加链接，http响应为空")
	}
	defer response.Body.Close()

	statusCode := response.StatusCode
	logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode}).Info("添加链接，http响应码")
	if statusCode != http.StatusOK {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"statusCode": statusCode}).Error("添加链接，http响应码失败")
		return nil, fmt.Errorf("添加链接，http响应码失败: %+v", statusCode)
	}

	return AddFile(ctx, filePath, response.Body, raw)
}

//func AddTmpFile(ctx context.Context, reader io.Reader) (*model.FileSimpleInfo, error) {
//	return AddFile(ctx, path.Join(".tmp", strconv.Itoa(int(util.GenId()))), reader, true)
//}

func AddFile(ctx context.Context, filePath string, reader io.Reader, raw bool) (*model.FileSimpleInfo, error) {
	filePath = util.ClearPath(ctx, path.Join("/", filePath))
	logrus.WithContext(ctx).WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件")

	if !strings.HasPrefix(filePath, model.TrashPath) {
		fileExt := path.Ext(filePath)
		logrus.WithContext(ctx).WithFields(logrus.Fields{"fileExt": fileExt}).Info("添加文件，文件拓展名")
		format, err := imaging.FormatFromExtension(fileExt)
		logrus.WithContext(ctx).WithFields(logrus.Fields{"format": format, "err": err}).Info("添加文件，解析图片拓展名")

		if !raw && err == nil && format != imaging.GIF {
			buffer := &bytes.Buffer{}
			reader = util.NewTimeoutReader(reader, config.Config.Timeout)
			_, err = io.Copy(buffer, reader)
			if err != nil {
				logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加文件，读取图片数据异常")
				return nil, fmt.Errorf("添加文件，读取图片数据异常: %+v", err)
			}

			imageBuffer, err := CompressionImage(ctx, buffer)
			if err != nil {
				reader = buffer
			} else {
				reader = imageBuffer
				filePath = AddImageExtension(ctx, filePath)
			}
		}

		_, err = RemoveFile(ctx, filePath)
		if err != nil {
			return nil, err
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
	filePath = util.ClearPath(ctx, path.Join("/", filePath))
	logrus.WithContext(ctx).WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件")

	info, err := GetFileSimpleInfo(ctx, filePath)
	if info == nil || err != nil {
		return nil, err
	}
	if !info.IsFile {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Info("删除文件，不允许删除文件夹")
		return nil, fmt.Errorf("删除文件，不允许删除文件夹")
	}

	if !config.Config.TrashEnable || strings.HasPrefix(filePath, model.TrashPath) {
		info, err := dao.DeleteFile(ctx, filePath)
		if err != nil {
			return nil, err
		}
		info = initFileSimpleInfo(ctx, info)
		return info, err
	}

	trashPath := genTrashPath(ctx, filePath)

	err = dao.MoveFile(ctx, filePath, trashPath)
	if err == nil {
		return info, nil
	}
	dao.MoveFile(ctx, trashPath, filePath)
	return info, err
}

func GetFileSimpleInfo(ctx context.Context, fileOrFolderPath string) (*model.FileSimpleInfo, error) {
	info, err := dao.SelectFileSimpleInfo(ctx, fileOrFolderPath)
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
