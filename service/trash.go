package service

import (
	"context"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/sirupsen/logrus"
	"path"
	"strconv"
	"strings"
	"time"
)

func ClearTrash(ctx context.Context) {
	clearTrash(ctx, model.TrashPath)
}

func clearTrash(ctx context.Context, folderPath string) error {
	infos, err := ListFileSimpleInfo(ctx, folderPath)
	if err != nil {
		return err
	}
	for i := range infos {
		if !infos[i].IsFile {
			clearTrash(ctx, infos[i].Path)
			continue
		}

		filePath := util.ClearPath(ctx, path.Join("/", infos[i].Path))
		if !strings.HasPrefix(filePath, model.TrashPath) {
			continue
		}
		_, logId := parseTrashPath(ctx, filePath)
		trashTime, err := util.ParseId(ctx, logId)
		if config.Config.TrashSaveTime <= time.Now().Sub(trashTime) || err != nil {
			RemoveFile(ctx, filePath)
		}
	}
	return nil
}

func parseTrashPath(ctx context.Context, filePath string) (string, int64) {
	fileExt := path.Ext(filePath)
	logIdPath := strings.TrimRight(filePath, fileExt)
	logIdExt := path.Ext(logIdPath)
	namePath := strings.TrimRight(logIdPath, logIdExt)
	filePath = fmt.Sprintf("%+v%+v", namePath, fileExt)
	var logId int64
	if logIdExt != "" {
		logId = util.String2Int64(logIdExt[1:])
	}
	return filePath, logId
}
func genTrashPath(ctx context.Context, filePath string) string {
	logId := util.GetLogId(ctx)
	return genTrashPathByLogId(ctx, filePath, logId)
}
func genTrashPathByLogId(ctx context.Context, filePath string, logId int64) string {
	fileExt := path.Ext(filePath)
	namePath := strings.TrimRight(filePath, fileExt)
	trashPath := fmt.Sprintf("%+v.%+v%+v", namePath, strconv.Itoa(int(logId)), fileExt)
	trashPath = path.Join(model.TrashPath, trashPath)
	logrus.WithContext(ctx).WithFields(logrus.Fields{"trashPath": trashPath}).Info("创建回收站路径")
	return trashPath
}
