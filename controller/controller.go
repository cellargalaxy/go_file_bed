package controller

import (
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/cellargalaxy/go_file_bed/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Controller() error {
	engine := gin.Default()
	engine.Use(claims)
	engine.Use(util.GinLog)
	engine.GET("/ping", util.Ping)
	engine.POST("/ping", validate, util.Ping)

	engine.Use(staticCache)
	engine.StaticFS("/static", http.FS(static.StaticFile))

	engine.Static(model.FileUrl, model.FileBedPath)

	engine.POST(model.AddUrlUrl, validate, addUrl)
	engine.POST(model.AddFileUrl, validate, addFile)
	engine.POST(model.RemoveFileUrl, validate, removeFile)
	engine.GET(model.GetFileCompleteInfoUrl, validate, getFileCompleteInfo)
	engine.GET(model.ListFileSimpleInfoUrl, validate, listFileSimpleInfo)
	engine.GET(model.ListLastFileInfoUrl, validate, listLastFileInfo)

	engine.POST(model.PushSyncFileUrl, validate, pushSyncFile)
	engine.POST(model.PullSyncFileUrl, validate, pullSyncFile)

	err := engine.Run(model.ListenAddress)
	if err != nil {
		panic(fmt.Errorf("web服务启动，异常: %+v", err))
	}
	return nil
}

func staticCache(c *gin.Context) {
	if strings.HasPrefix(c.Request.RequestURI, "/static") {
		c.Header("Cache-Control", "max-age=86400")
	} else if strings.HasPrefix(c.Request.RequestURI, model.FileUrl) {
		c.Header("Cache-Control", "max-age=31536000")
	}
}
