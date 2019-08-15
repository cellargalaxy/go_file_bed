package service

import (
	"../cache"
	"../model"
	"../utils"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"path"
	"strings"
	"time"
)

var FileBedPath = path.Clean("n1")
var log = logrus.New()

func AddFile(sort string, filename string, reader io.Reader) error {
	filePath := path.Join(FileBedPath, sort, time.Now().Format("20060102"), filename)
	filePath = utils.ClearPath(filePath)
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("添加文件路径")

	if !strings.HasPrefix(filePath, FileBedPath) {
		log.WithFields(logrus.Fields{"filePath": filePath}).Info("添加文件路径不在指定路径下")
		return errors.New(fmt.Sprintf("添加文件路径不在指定路径下: %v", filePath))
	}
	return cache.InsertFile(filePath, reader)
}

func RemoveFile(filePath string) error {
	filePath = path.Join(FileBedPath, filePath)
	filePath = utils.ClearPath(filePath)
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件路径")

	if !strings.HasPrefix(filePath, FileBedPath) {
		log.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件路径不在指定路径下")
		return errors.New(fmt.Sprintf("删除文件路径不在指定路径下: %v", filePath))
	}
	return cache.DeleteFile(filePath)
}

func ListFileOrFolderInfo(fileOrFolderPath string) ([]model.FileOrFolderInfo, error) {
	fileOrFolderPath = path.Join(FileBedPath, fileOrFolderPath)
	fileOrFolderPath = utils.ClearPath(fileOrFolderPath)
	log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("查询文件路径")

	if !strings.HasPrefix(fileOrFolderPath, FileBedPath) {
		log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("查询文件路径不在指定路径下")
		return nil, errors.New(fmt.Sprintf("查询文件路径不在指定路径下: %v", fileOrFolderPath))
	}

	fileOrFolderInfos, err := cache.SelectFileOrFolder(fileOrFolderPath)
	if fileOrFolderInfos != nil && err == nil {
		for i := range fileOrFolderInfos {
			fileOrFolderInfos[i].Path = utils.ClearPath(strings.Replace(fileOrFolderInfos[i].Path, FileBedPath, "", 1))
		}
	}
	return fileOrFolderInfos, nil
}

func ListAllFileInfo() ([]model.FileOrFolderInfo, error) {
	fileInfos, err := cache.SelectAllFile(FileBedPath)
	if fileInfos != nil && err == nil {
		for i := range fileInfos {
			fileInfos[i].Path = utils.ClearPath(strings.Replace(fileInfos[i].Path, FileBedPath, "", 1))
		}
	}
	return fileInfos, nil
}
