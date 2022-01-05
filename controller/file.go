package controller

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/cellargalaxy/go_file_bed/service/controller"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func addUrl(ctx *gin.Context) {
	var request model.UrlAddRequest
	err := ctx.Bind(&request)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加链接，请求参数解析异常")
		ctx.JSON(http.StatusOK, util.CreateErrResponse(err.Error()))
		return
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"request": request}).Info("添加链接")
	ctx.JSON(http.StatusOK, util.CreateResponse(controller.AddUrl(ctx, request)))
}

func addFile(ctx *gin.Context) {
	filePath := ctx.Request.FormValue("path")
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("添加文件，读取表单文件异常")
		ctx.JSON(http.StatusOK, util.CreateErrResponse(err.Error()))
		return
	}
	defer file.Close()
	logrus.WithContext(ctx).WithFields(logrus.Fields{"filePath": filePath, "filename": header.Filename}).Info("添加文件")

	ctx.JSON(http.StatusOK, util.CreateResponse(controller.AddFile(ctx, filePath, file)))
}

func removeFile(ctx *gin.Context) {
	var request model.FileRemoveRequest
	err := ctx.Bind(&request)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("删除文件，请求参数解析异常")
		ctx.JSON(http.StatusOK, util.CreateErrResponse(err.Error()))
		return
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"request": request}).Info("删除文件")
	ctx.JSON(http.StatusOK, util.CreateResponse(controller.RemoveFile(ctx, request)))
}

func getFileCompleteInfo(ctx *gin.Context) {
	var request model.FileCompleteInfoGetRequest
	err := ctx.Bind(&request)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询文件完整信息，请求参数解析异常")
		ctx.JSON(http.StatusOK, util.CreateErrResponse(err.Error()))
		return
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"request": request}).Info("查询文件完整信息")
	ctx.JSON(http.StatusOK, util.CreateResponse(controller.GetFileCompleteInfo(ctx, request)))
}

func listFileSimpleInfo(ctx *gin.Context) {
	var request model.FileSimpleInfoListRequest
	err := ctx.Bind(&request)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询文件简单信息，请求参数解析异常")
		ctx.JSON(http.StatusOK, util.CreateErrResponse(err.Error()))
		return
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"request": request}).Info("查询文件简单信息")
	ctx.JSON(http.StatusOK, util.CreateResponse(controller.ListFileSimpleInfo(ctx, request)))
}

func listLastFileInfo(ctx *gin.Context) {
	var request model.LastFileInfoListRequest
	err := ctx.Bind(&request)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("查询最近文件信息，请求参数解析异常")
		ctx.JSON(http.StatusOK, util.CreateErrResponse(err.Error()))
		return
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"request": request}).Info("查询最近文件信息")
	ctx.JSON(http.StatusOK, util.CreateResponse(controller.ListLastFileInfo(ctx, request)))
}
