package config

import (
	"github.com/disintegration/imaging"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

const (
	SuccessCode = 1
	FailCode    = 2

	FileUrl  = "/file"
	LoginUrl = "/login"

	UploadUrlUrl             = "/admin/uploadUrl"
	UploadFileUrl            = "/admin/uploadFile"
	RemoveFileUrl            = "/admin/removeFile"
	GetFileCompleteInfoUrl   = "/admin/getFileCompleteInfo"
	ListLastFileInfoUrl      = "/admin/listLastFileInfo"
	ListFolderInfoUrl        = "/admin/listFolderInfo"
	ListAllFileSimpleInfoUrl = "/admin/listAllFileSimpleInfo"
	ReceivePushSyncFileUrl   = "/admin/receivePushSyncFile"
	PushSyncFileUrl          = "/admin/pushSyncFile"
	PullSyncFileUrl          = "/admin/pullSyncFile"

	ConfigFilePath    = "config.ini"
	ImageTargetSize   = 1024 * 200
	ImageSaveFormat   = imaging.JPEG
	Timeout           = 5 * time.Second
	PullOrPushTimeout = 60 * 60 * time.Second

	defaultToken             = "token"
	defaultListenAddress     = ":8880"
	defaultFileBedPath       = "file_bed"
	defaultLastFileInfoCount = 10
)

var Token = defaultToken
var ListenAddress = defaultListenAddress
var FileBedPath = defaultFileBedPath
var LastFileInfoCount = defaultLastFileInfoCount

var PullSyncCron string
var PullSyncHost string
var PullSyncToken string
var PushSyncCron string
var PushSyncHost string
var PushSyncToken string

func init() {
	logrus.Info("加载配置开始")

	configFile, err := ini.Load(ConfigFilePath)
	if err == nil {
		Token = configFile.Section("").Key("token").MustString(defaultToken)
		ListenAddress = configFile.Section("").Key("listenAddress").MustString(defaultListenAddress)
		FileBedPath = configFile.Section("").Key("fileBedPath").MustString(defaultFileBedPath)
		LastFileInfoCount = configFile.Section("").Key("lastFileInfoCount").MustInt(defaultLastFileInfoCount)
		PullSyncCron = configFile.Section("").Key("pullSyncCron").String()
		PullSyncHost = configFile.Section("").Key("pullSyncHost").String()
		PullSyncToken = configFile.Section("").Key("pullSyncToken").String()
		PushSyncCron = configFile.Section("").Key("pushSyncCron").String()
		PushSyncHost = configFile.Section("").Key("pushSyncHost").String()
		PushSyncToken = configFile.Section("").Key("pushSyncToken").String()
	}

	token := os.Getenv("TOKEN")
	logrus.WithFields(logrus.Fields{"token": len(token)}).Info("环境变量读取配置Token")
	if token != "" {
		Token = token
	}
	listenAddress := os.Getenv("LISTEN_ADDRESS")
	logrus.WithFields(logrus.Fields{"listenAddress": listenAddress}).Info("环境变量读取配置ListenAddress")
	if listenAddress != "" {
		ListenAddress = listenAddress
	}
	fileBedPath := os.Getenv("FILE_BED_PATH")
	logrus.WithFields(logrus.Fields{"fileBedPath": fileBedPath}).Info("环境变量读取配置FileBedPath")
	if fileBedPath != "" {
		FileBedPath = fileBedPath
	}
	lastFileInfoCountString := os.Getenv("LAST_FILE_INFO_COUNT")
	logrus.WithFields(logrus.Fields{"lastFileInfoCountString": lastFileInfoCountString}).Info("环境变量读取配置LastFileInfoCount")
	lastFileInfoCount, err := strconv.Atoi(lastFileInfoCountString)
	if err == nil && lastFileInfoCount > 0 {
		LastFileInfoCount = lastFileInfoCount
	}

	pullSyncCron := os.Getenv("PULL_SYNC_CRON")
	logrus.WithFields(logrus.Fields{"pullSyncCron": pullSyncCron}).Info("环境变量读取配置PullSyncCron")
	if pullSyncCron != "" {
		PullSyncCron = pullSyncCron
	}
	pullSyncHost := os.Getenv("PULL_SYNC_HOST")
	logrus.WithFields(logrus.Fields{"pullSyncHost": pullSyncHost}).Info("环境变量读取配置PullSyncHost")
	if pullSyncHost != "" {
		PullSyncHost = pullSyncHost
	}
	pullSyncToken := os.Getenv("PULL_SYNC_TOKEN")
	logrus.WithFields(logrus.Fields{"pullSyncToken": pullSyncToken}).Info("环境变量读取配置PullSyncToken")
	if pullSyncToken != "" {
		PullSyncToken = pullSyncToken
	}
	pushSyncCron := os.Getenv("PUSH_SYNC_CRON")
	logrus.WithFields(logrus.Fields{"pushSyncCron": pushSyncCron}).Info("环境变量读取配置PushSyncCron")
	if pushSyncCron != "" {
		PushSyncCron = pushSyncCron
	}
	pushSyncHost := os.Getenv("PUSH_SYNC_HOST")
	logrus.WithFields(logrus.Fields{"pushSyncHost": pushSyncHost}).Info("环境变量读取配置PushSyncHost")
	if pushSyncHost != "" {
		PushSyncHost = pushSyncHost
	}
	pushSyncToken := os.Getenv("PUSH_SYNC_TOKEN")
	logrus.WithFields(logrus.Fields{"pushSyncToken": pushSyncToken}).Info("环境变量读取配置PushSyncToken")
	if pushSyncToken != "" {
		PushSyncToken = pushSyncToken
	}

	logrus.WithFields(logrus.Fields{"Token": len(Token)}).Info("配置Token")
	logrus.WithFields(logrus.Fields{"ListenAddress": ListenAddress}).Info("配置ListenAddress")
	logrus.WithFields(logrus.Fields{"FileBedPath": FileBedPath}).Info("配置FileBedPath")
	logrus.WithFields(logrus.Fields{"LastFileInfoCount": LastFileInfoCount}).Info("配置LastFileInfoCount")
	logrus.WithFields(logrus.Fields{"PullSyncCron": PullSyncCron}).Info("配置PullSyncCron")
	logrus.WithFields(logrus.Fields{"PullSyncHost": PullSyncHost}).Info("配置PullSyncHost")
	logrus.WithFields(logrus.Fields{"PullSyncToken": len(PullSyncToken)}).Info("配置PullSyncToken")
	logrus.WithFields(logrus.Fields{"PushSyncCron": PushSyncCron}).Info("配置PushSyncCron")
	logrus.WithFields(logrus.Fields{"PushSyncHost": PushSyncHost}).Info("配置PushSyncHost")
	logrus.WithFields(logrus.Fields{"PushSyncToken": len(PushSyncToken)}).Info("配置PushSyncToken")

	err = os.MkdirAll(FileBedPath, 0666)
	if err != nil {
		logrus.WithFields(logrus.Fields{"folderPath": FileBedPath, "err": err}).Error("创建文件夹失败")
	}

	logrus.Info("加载配置完成")
}
