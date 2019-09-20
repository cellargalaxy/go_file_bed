package controller

import (
	"../config"
	"../model"
	"../service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"net/http"
)

var log = logrus.New()
var secretKey = "secret"
var secret = uuid.Must(uuid.NewV4()).String()
var fileBedPath = config.GetConfig().FileBedPath
var listenAddress = config.GetConfig().ListenAddress

func Controller() {
	log.WithFields(logrus.Fields{"fileBedPath": fileBedPath}).Info("文件床路径")
	log.WithFields(logrus.Fields{"listenAddress": listenAddress}).Info("监听地址")

	store := cookie.NewStore([]byte(secret))

	engine := gin.Default()
	engine.Use(sessions.Sessions("session_id", store))

	engine.StaticFile("/", "goFileBed.html")
	engine.Static(service.FileUrl, fileBedPath)
	engine.POST(service.LoginUrl, loginController)

	engine.POST(service.UploadUrlUrl, validate, uploadUrlController)
	engine.POST(service.UploadFileUrl, validate, uploadFileController)
	engine.POST(service.UploadFileByFilePathUrl, validate, uploadFileByFilePathController)
	engine.POST(service.RemoveFileUrl, validate, removeFileController)
	engine.GET(service.ListFolderInfoUrl, validate, listFolderInfoController)
	engine.GET(service.ListAllFileInfoUrl, validate, listAllFileInfoController)
	engine.POST(service.ReceivePushSynFileUrl, validate, receivePushSynFileController)
	engine.POST(service.PushSynFileUrl, validate, pushSynFileController)
	engine.POST(service.PullSynFileUrl, validate, pullSynFileController)
	engine.POST(service.ClearAllCacheUrl, validate, clearAllCacheController)

	engine.Run(listenAddress)
}

func validate(context *gin.Context) {
	if !isLogin(context) {
		context.Abort()
		context.JSON(http.StatusUnauthorized, createFailResponse("please login"))
	} else {
		context.Next()
	}
}

func setLogin(context *gin.Context) {
	session := sessions.Default(context)
	session.Set(secretKey, secret)
	session.Save()
}

func isLogin(context *gin.Context) bool {
	session := sessions.Default(context)
	sessionSecret := session.Get(secretKey)
	return sessionSecret == secret
}

func loginController(context *gin.Context) {
	token := context.Request.FormValue("token")
	log.Info("用户登录")

	if service.CheckToken(token) {
		setLogin(context)
		context.JSON(http.StatusOK, createSuccessResponse("login success", nil))
	} else {
		log.WithFields(logrus.Fields{"token": token}).Info("非法token")
		context.JSON(http.StatusOK, createFailResponse("illegal token"))
	}
}

func uploadUrlController(context *gin.Context) {
	sort := context.Request.FormValue("sort")
	url := context.Request.FormValue("url")
	log.WithFields(logrus.Fields{"sort": sort, "url": url}).Info("上传url文件")

	fileInfo, err := service.AddUrl(sort, url)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("upload file success", fileInfo))
	}
}

func uploadFileController(context *gin.Context) {
	sort := context.Request.FormValue("sort")
	file, header, err := context.Request.FormFile("file")
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
		return
	}

	filename := ""
	if header != nil {
		filename = header.Filename
	}
	if filename == "" {
		filename = uuid.Must(uuid.NewV4()).String()
	}
	log.WithFields(logrus.Fields{"sort": sort, "filename": filename}).Info("上传文件")

	fileInfo, err := service.AddFile(sort, filename, file)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("upload file success", fileInfo))
	}
}

func uploadFileByFilePathController(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	file, _, err := context.Request.FormFile("file")
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
		return
	}

	log.WithFields(logrus.Fields{"filePath": filePath}).Info("上传文件")

	fileInfo, err := service.AddFileByFilePath(filePath, file)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("upload file success", fileInfo))
	}
}

func removeFileController(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件")

	fileInfo, err := service.RemoveFile(filePath)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("remove file success", fileInfo))
	}
}

func listFolderInfoController(context *gin.Context) {
	folderPath := context.Query("folderPath")
	log.WithFields(logrus.Fields{"folderPath": folderPath}).Info("查询文件")

	fileOrFolderInfos, err := service.ListFolderInfo(folderPath)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("", fileOrFolderInfos))
	}
}

func listAllFileInfoController(context *gin.Context) {
	log.Info("查询所有文件")

	fileInfos, err := service.ListAllFileInfo()
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("", fileInfos))
	}
}

func receivePushSynFileController(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	md5 := context.Request.FormValue("md5")
	file, _, err := context.Request.FormFile("file")
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
		return
	}
	log.WithFields(logrus.Fields{"filePath": filePath, "md5": md5}).Info("接收推送同步文件")

	err = service.ReceivePushSynFile(filePath, md5, file)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("", nil))
	}
}

func pushSynFileController(context *gin.Context) {
	pushSynHost := context.Request.FormValue("pushSynHost")
	token := context.Request.FormValue("token")
	log.WithFields(logrus.Fields{"pushSynHost": pushSynHost}).Info("推送同步文件")

	failCount, err := service.PushSynFile(pushSynHost, token)
	if err != nil {
		context.JSON(http.StatusOK, createResponse(model.FailCode, "", map[string]interface{}{"failCount": failCount, "err": err}))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("", map[string]interface{}{"failCount": failCount, "err": err}))
	}
}

func pullSynFileController(context *gin.Context) {
	pullSynHost := context.Request.FormValue("pullSynHost")
	token := context.Request.FormValue("token")
	log.WithFields(logrus.Fields{"pullSynHost": pullSynHost}).Info("拉取同步文件")

	failCount, err := service.PullSynFile(pullSynHost, token)
	if err != nil {
		context.JSON(http.StatusOK, createResponse(model.FailCode, "", map[string]interface{}{"failCount": failCount, "err": err}))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("", map[string]interface{}{"failCount": failCount, "err": err}))
	}
}

func clearAllCacheController(context *gin.Context) {
	log.Info("清除全部缓存")

	service.ClearAllCache()
	context.JSON(http.StatusOK, createSuccessResponse("clear all cache success", nil))
}

func createSuccessResponse(massage string, data interface{}) map[string]interface{} {
	return createResponse(model.SuccessCode, massage, data)
}
func createFailResponse(massage string) map[string]interface{} {
	return createResponse(model.FailCode, massage, nil)
}
func createResponse(code int, massage string, data interface{}) map[string]interface{} {
	return gin.H{"code": code, "massage": massage, "data": data}
}
