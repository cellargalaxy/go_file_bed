package config

import (
	"context"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	sc_model "github.com/cellargalaxy/server_center/model"
	"github.com/cellargalaxy/server_center/sdk"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"time"
)

var Config = model.Config{}

func init() {
	ctx := util.CreateLogCtx()
	client, err := sdk.NewDefaultServerCenterClient(ctx, &ServerCenterHandler{})
	if err != nil {
		panic(err)
	}
	client.StartConfWithInitConf(ctx)
}

func checkAndResetConfig(ctx context.Context, config model.Config) (model.Config, error) {
	if config.Timeout <= 0 {
		config.Timeout = 3 * time.Second
	}
	if config.Sleep < 0 {
		config.Sleep = 3 * time.Second
	}

	if config.LastFileCount <= 0 {
		config.LastFileCount = 10
	}
	if config.MaxHashLimit <= 0 {
		config.MaxHashLimit = 1024 * 1024 * 128 //128M
	}
	if config.TrashSaveTime <= 0 {
		config.TrashSaveTime = 30 * 24 * time.Hour
	}

	if config.ImageTargetSize <= 0 {
		config.ImageTargetSize = 1024 * 200 //200K
	}
	if config.JpegMinQuality <= 0 {
		config.JpegMinQuality = 20
	}
	if config.JpegMaxQuality <= 0 {
		config.JpegMaxQuality = 80
	}
	if config.JpegMaxQuality <= config.JpegMinQuality {
		config.JpegMinQuality = 20
		config.JpegMaxQuality = 80
	}
	if config.ImageSaveFormat < imaging.JPEG || imaging.BMP < config.ImageSaveFormat {
		config.ImageSaveFormat = imaging.JPEG
	}

	if config.Secret == "" {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Error("secret为空")
		return config, fmt.Errorf("secret为空")
	}

	err := util.CreateFolderPath(ctx, model.FileBedPath)
	return config, err
}

type ServerCenterHandler struct {
}

func (this *ServerCenterHandler) GetAddress(ctx context.Context) string {
	return sdk.GetEnvServerCenterAddress(ctx)
}
func (this *ServerCenterHandler) GetSecret(ctx context.Context) string {
	return sdk.GetEnvServerCenterSecret(ctx)
}
func (this *ServerCenterHandler) GetServerName(ctx context.Context) string {
	return sdk.GetEnvServerName(ctx, model.DefaultServerName)
}
func (this *ServerCenterHandler) GetInterval(ctx context.Context) time.Duration {
	return 5 * time.Minute
}
func (this *ServerCenterHandler) ParseConf(ctx context.Context, object sc_model.ServerConfModel) error {
	var config model.Config
	err := util.UnmarshalYamlString(object.ConfText, &config)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("反序列化配置异常")
		return err
	}
	config, err = checkAndResetConfig(ctx, config)
	if err != nil {
		return err
	}
	Config = config
	logrus.WithContext(ctx).WithFields(logrus.Fields{"Config": Config}).Info("加载配置")
	return nil
}
func (this *ServerCenterHandler) GetDefaultConf(ctx context.Context) string {
	var config model.Config
	config, _ = checkAndResetConfig(ctx, config)
	return util.ToYamlString(config)
}
func (this *ServerCenterHandler) GetLocalFilePath(ctx context.Context) string {
	return ""
}
