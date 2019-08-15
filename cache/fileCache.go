package cache

import (
	"../dao"
	"../model"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
)

var log = logrus.New()
var cache sync.Map

//添加文件
func InsertFile(filePath string, reader io.Reader) error {
	err := dao.InsertFile(filePath, reader)
	if err == nil {
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

//删除文件，随便删除文件其上的空文件夹
func DeleteFile(filePath string) error {
	err := dao.DeleteFile(filePath)
	if err == nil {
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

//查询文件夹下的情况
func SelectFileOrFolder(fileOrFolderPath string) ([]model.FileOrFolderInfo, error) {
	fileOrFolderInfos, _ := cache.Load(fileOrFolderPath)
	if fileOrFolderInfos != nil {
		return fileOrFolderInfos.([]model.FileOrFolderInfo), nil
	}
	fileOrFolderInfos, err := dao.SelectFileOrFolder(fileOrFolderPath)
	if err == nil {
		cache.Store(fileOrFolderPath, fileOrFolderInfos)
	}
	return fileOrFolderInfos.([]model.FileOrFolderInfo), err
}
