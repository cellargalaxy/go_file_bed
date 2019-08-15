package cache

import (
	"../dao"
	"../model"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
)

var log = logrus.New()
var cache sync.Map
var selectAllFileKey = uuid.Must(uuid.NewV4()).String()

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

//查询某个文件夹下的全部文件的信息
func SelectAllFile(folderPath string) ([]model.FileOrFolderInfo, error) {
	fileInfos, _ := cache.Load(selectAllFileKey)
	if fileInfos != nil {
		return fileInfos.([]model.FileOrFolderInfo), nil
	}
	fileInfos, err := dao.SelectAllFile(folderPath)
	if err == nil {
		cache.Store(selectAllFileKey, fileInfos)
	}
	return fileInfos.([]model.FileOrFolderInfo), err
}
