package cache

import (
	"../config"
	"../model"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

var log = logrus.New()
var fileBedPath = config.GetConfig().FileBedPath
var cache sync.Map
var selectAllFileKey = uuid.Must(uuid.NewV4()).String()

func DeleteFileOrFolderInfo(filePath string) {
	cache.Delete(selectAllFileKey)
	cache.Range(func(fileOrFolderPath, _ interface{}) bool {
		if strings.HasPrefix(filePath, fileOrFolderPath.(string)) {
			log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("删除缓存")
			cache.Delete(fileOrFolderPath)
		}
		return true
	})
}

func SelectListFileOrFolderInfo(fileOrFolderPath string) []model.FileOrFolderInfo {
	fileOrFolderInfos, _ := cache.Load(fileOrFolderPath)
	if fileOrFolderInfos == nil {
		return nil
	}
	return fileOrFolderInfos.([]model.FileOrFolderInfo)
}

func InsertListFileOrFolderInfo(fileOrFolderPath string, fileOrFolderInfos []model.FileOrFolderInfo) {
	if fileOrFolderPath != "" && fileOrFolderInfos != nil {
		cache.Store(fileOrFolderPath, fileOrFolderInfos)
	}
}

func SelectListAllFileInfo() []model.FileOrFolderInfo {
	fileInfos, _ := cache.Load(selectAllFileKey)
	if fileInfos == nil {
		return nil
	}
	return fileInfos.([]model.FileOrFolderInfo)
}

func InsertListAllFileInfo(fileOrFolderInfos []model.FileOrFolderInfo) {
	if fileOrFolderInfos != nil {
		cache.Store(selectAllFileKey, fileOrFolderInfos)
	}
}

func ClearAllCache() error {
	cache = sync.Map{}
	return nil
}
