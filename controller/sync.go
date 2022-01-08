package controller

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/cellargalaxy/go_file_bed/service/controller"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func pushSyncFile(ctx *gin.Context) {
	var request model.PushSyncFileRequest
	err := ctx.Bind(&request)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("push同步文件，请求参数解析异常")
		ctx.JSON(http.StatusOK, util.CreateErrResponse(err.Error()))
		return
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"address": request.Address, "path": request.Path}).Info("push同步文件")
	ctx.JSON(http.StatusOK, util.CreateResponse(controller.PushSyncFile(ctx, request)))
}

func pullSyncFile(ctx *gin.Context) {
	var request model.PullSyncFileRequest
	err := ctx.Bind(&request)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("pull同步文件，请求参数解析异常")
		ctx.JSON(http.StatusOK, util.CreateErrResponse(err.Error()))
		return
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"address": request.Address, "path": request.Path}).Info("pull同步文件")
	ctx.JSON(http.StatusOK, util.CreateResponse(controller.PullSyncFile(ctx, request)))
}
