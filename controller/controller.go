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
	"os"
)

var log = logrus.New()
var secretKey = "secret"
var secret = uuid.Must(uuid.NewV4()).String()
var fileBedPath = config.GetConfig().FileBedPath

func init() {
	os.MkdirAll(fileBedPath, 0666)
}

func Controller() {
	store := cookie.NewStore([]byte(secret))

	engine := gin.Default()
	engine.Use(sessions.Sessions("session_id", store))

	engine.StaticFile("/", "goFileBed.html")
	engine.Static(service.FileUrl, fileBedPath)
	engine.POST(service.LoginUrl, loginController)

	engine.POST(service.UploadFileUrl, validate, uploadFileController)
	engine.POST(service.UploadUrlUrl, validate, uploadUrlController)
	engine.POST(service.RemoveFileUrl, validate, removeFileController)
	engine.GET(service.ListFileOrFolderInfoUrl, validate, listFileOrFolderInfoController)
	engine.GET(service.ListAllFileInfoUrl, validate, listAllFileInfoController)

	engine.Run(config.GetConfig().ListeningAddress)
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

func listAllFileInfoController(context *gin.Context) {
	log.Info("查询所有文件")

	fileInfos, err := service.ListAllFileInfo()
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("", fileInfos))
	}
}

func listFileOrFolderInfoController(context *gin.Context) {
	fileOrFolderPath := context.Query("fileOrFolderPath")
	log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("查询文件")

	fileOrFolderInfos, err := service.ListFileOrFolderInfo(fileOrFolderPath)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("", fileOrFolderInfos))
	}
}

func uploadFileController(context *gin.Context) {
	sort := context.Request.FormValue("sort")
	file, header, err := context.Request.FormFile("file")
	filename := ""
	if header != nil {
		filename = header.Filename
	}
	log.WithFields(logrus.Fields{"sort": sort, "filename": filename}).Info("上传文件")

	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
		return
	}

	fileInfo, err := service.AddFile(sort, filename, file)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("upload file success", fileInfo))
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

func createSuccessResponse(massage string, data interface{}) map[string]interface{} {
	return createResponse(model.SuccessCode, massage, data)
}
func createFailResponse(massage string) map[string]interface{} {
	return createResponse(model.FailCode, massage, nil)
}
func createResponse(code int, massage string, data interface{}) map[string]interface{} {
	return gin.H{"code": code, "massage": massage, "data": data}
}
