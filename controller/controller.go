package controller

import (
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func Controller() error {
	engine := gin.Default()

	engine.GET("/", func(context *gin.Context) {
		context.Header("Content-Type", "text/html; charset=utf-8")
		context.String(200, indexHtmlString)
	})
	//engine.StaticFile("/","static/html/index.html")
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	engine.Static(config.FileUrl, config.FileBedPath)
	engine.POST(config.LoginUrl, login)

	engine.POST(config.UploadUrlUrl, validate, uploadUrl)
	engine.POST(config.UploadFileUrl, validate, uploadFile)
	engine.POST(config.RemoveFileUrl, validate, removeFile)
	engine.GET(config.GetFileCompleteInfoUrl, validate, getFileCompleteInfo)
	engine.GET(config.ListLastFileInfoUrl, validate, listLastFileInfo)
	engine.GET(config.ListFolderInfoUrl, validate, listFolderInfo)
	engine.GET(config.ListAllFileSimpleInfoUrl, validate, listAllFileSimpleInfo)
	engine.POST(config.ReceivePushSyncFileUrl, validate, receivePushSyncFile)
	engine.POST(config.PushSyncFileUrl, validate, pushSyncFile)
	engine.POST(config.PullSyncFileUrl, validate, pullSyncFile)

	engine.Run(config.ListenAddress)

	return nil
}
