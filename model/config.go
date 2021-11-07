package model

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	ListenAddress = ":8880"
	FileBedPath   = "file_bed"

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
)

type Config struct {
	LogLevel logrus.Level  `yaml:"log_level" json:"log_level"`
	Retry    int           `yaml:"retry" json:"retry"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Sleep    time.Duration `yaml:"sleep" json:"sleep"`
	Secret   string        `yaml:"secret" json:"-"`

	LastFileCount int `yaml:"last_file_count" json:"last_file_count"`

	ImageTargetSize float64        `yaml:"image_target_size" json:"image_target_size"`
	JpegMinQuality  float64        `yaml:"jpeg_min_quality" json:"jpeg_min_quality"`
	JpegMaxQuality  float64        `yaml:"jpeg_max_quality" json:"jpeg_max_quality"`
	ImageSaveFormat imaging.Format `yaml:"image_save_format" json:"image_save_format"`

	PullSyncCron   string `yaml:"pull_sync_cron" json:"pull_sync_cron"`
	PullSyncHost   string `yaml:"pull_sync_host" json:"pull_sync_host"`
	PullSyncSecret string `yaml:"pull_sync_secret" json:"-"`
	PushSyncCron   string `yaml:"push_sync_cron" json:"push_sync_cron"`
	PushSyncHost   string `yaml:"push_sync_host" json:"push_sync_host"`
	PushSyncSecret string `yaml:"push_sync_secret" json:"-"`
}

func (this Config) String() string {
	return util.ToJsonString(this)
}
