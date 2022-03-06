package corn

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/cellargalaxy/go_file_bed/service"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func Init() {
	ctx := util.CreateLogCtx()
	cronObject := cron.New()

	if config.Config.PullSyncCron != "" {
		var job pullSyncFileJob
		job.Address = config.Config.PullSyncHost
		job.Secret = config.Config.PullSyncSecret
		entryId, err := cronObject.AddJob(config.Config.PullSyncCron, &job)
		if err != nil {
			panic(err)
		}
		logrus.WithContext(ctx).WithFields(logrus.Fields{"corn": job, "entryId": entryId}).Info("定时任务，添加定时")
	}

	if config.Config.PushSyncCron != "" {
		var job pushSyncFileJob
		job.Address = config.Config.PushSyncHost
		job.Secret = config.Config.PushSyncSecret
		entryId, err := cronObject.AddJob(config.Config.PushSyncCron, &job)
		if err != nil {
			panic(err)
		}
		logrus.WithContext(ctx).WithFields(logrus.Fields{"corn": job, "entryId": entryId}).Info("定时任务，添加定时")
	}

	if config.Config.TrashClearCron != "" {
		var job trashClearJob
		entryId, err := cronObject.AddJob(config.Config.TrashClearCron, &job)
		if err != nil {
			panic(err)
		}
		logrus.WithContext(ctx).WithFields(logrus.Fields{"corn": job, "entryId": entryId}).Info("定时任务，添加定时")
	}

	cronObject.Start()
	logrus.WithContext(ctx).WithFields(logrus.Fields{}).Info("定时任务，添加完成")
}

type pushSyncFileJob struct {
	Address string `json:"address"`
	Secret  string `json:"-"`
}

func (this pushSyncFileJob) String() string {
	return util.ToJsonString(this)
}

func (this *pushSyncFileJob) Run() {
	ctx := util.CreateLogCtx()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"pushSyncFileJob": this}).Info("定时任务，执行任务开始")
	service.PushSyncFile(ctx, this.Address, this.Secret, "")
	logrus.WithContext(ctx).WithFields(logrus.Fields{"pushSyncFileJob": this}).Info("定时任务，执行任务完成")
}

type pullSyncFileJob struct {
	Address string `json:"address"`
	Secret  string `json:"-"`
}

func (this pullSyncFileJob) String() string {
	return util.ToJsonString(this)
}

func (this *pullSyncFileJob) Run() {
	ctx := util.CreateLogCtx()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"pullSyncFileJob": this}).Info("定时任务，执行任务开始")
	service.PullSyncFile(ctx, this.Address, this.Secret, "")
	logrus.WithContext(ctx).WithFields(logrus.Fields{"pullSyncFileJob": this}).Info("定时任务，执行任务完成")
}

type trashClearJob struct {
}

func (this trashClearJob) String() string {
	return util.ToJsonString(this)
}

func (this *trashClearJob) Run() {
	ctx := util.CreateLogCtx()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"trashClearJob": this}).Info("定时任务，执行任务开始")
	service.ClearTrash(ctx)
	logrus.WithContext(ctx).WithFields(logrus.Fields{"trashClearJob": this}).Info("定时任务，执行任务完成")
}
