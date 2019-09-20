package cache

import (
	"../model"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

var log = logrus.New()
var FileCache sync.Map
var FolderCache sync.Map

func InsertFileCache(filePath string, fileOrFolderInfos []model.FileOrFolderInfo) {
	if fileOrFolderInfos == nil || len(fileOrFolderInfos) == 0 {
		log.Info("fileOrFolderInfos为nil或者长度为0")
		return
	}
	FileCache.Store(filePath, fileOrFolderInfos)
}

func SelectFileCache(filePath string) []model.FileOrFolderInfo {
	fileOrFolderInfos, _ := FileCache.Load(filePath)
	if fileOrFolderInfos == nil {
		return nil
	}
	return fileOrFolderInfos.([]model.FileOrFolderInfo)
}

func InsertFolderCache(folderPath string, fileOrFolderInfos []model.FileOrFolderInfo) {
	if fileOrFolderInfos == nil || len(fileOrFolderInfos) == 0 {
		log.Info("fileOrFolderInfos为nil或者长度为0")
		return
	}
	FolderCache.Store(folderPath, fileOrFolderInfos)
}

func SelectFolderCache(folderPath string) []model.FileOrFolderInfo {
	fileOrFolderInfos, _ := FolderCache.Load(folderPath)
	if fileOrFolderInfos == nil {
		return nil
	}
	return fileOrFolderInfos.([]model.FileOrFolderInfo)
}

func DeleteCache(fileOrFolderPath string) {
	FileCache.Range(func(filePath, _ interface{}) bool {
		if strings.HasPrefix(fileOrFolderPath, filePath.(string)) {
			log.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件缓存")
			FileCache.Delete(filePath)
		}
		return true
	})
	FolderCache.Range(func(folderPath, _ interface{}) bool {
		if strings.HasPrefix(fileOrFolderPath, folderPath.(string)) {
			log.WithFields(logrus.Fields{"folderPath": folderPath}).Info("删除文件夹缓存")
			FolderCache.Delete(folderPath)
		}
		return true
	})
}

func ClearAllCache() {
	FileCache = sync.Map{}
	FolderCache = sync.Map{}
}
