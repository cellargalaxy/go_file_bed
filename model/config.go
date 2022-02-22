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
	TrashPath     = "/.trash"

	FileUrl = "/file"

	AddUrlUrl              = "/api/addUrl"
	AddFileUrl             = "/api/addFile"
	RemoveFileUrl          = "/api/removeFile"
	GetFileCompleteInfoUrl = "/api/getFileCompleteInfo"
	ListFileSimpleInfoUrl  = "/api/listFileSimpleInfo"
	ListLastFileInfoUrl    = "/api/listLastFileInfo"
	PushSyncFileUrl        = "/api/pushSyncFile"
	PullSyncFileUrl        = "/api/pullSyncFile"
)

type Config struct {
	LogLevel logrus.Level  `yaml:"log_level" json:"log_level"`
	Retry    int           `yaml:"retry" json:"retry"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Sleep    time.Duration `yaml:"sleep" json:"sleep"`
	Secret   string        `yaml:"secret" json:"-"`

	LastFileCount  int           `yaml:"last_file_count" json:"last_file_count"`
	MaxHashLimit   int64         `yaml:"max_hash_limit" json:"max_hash_limit"`
	TrashSaveTime  time.Duration `yaml:"trash_save_time" json:"trash_save_time"`
	TrashClearCron string        `yaml:"trash_clear_cron" json:"trash_clear_cron"`

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
