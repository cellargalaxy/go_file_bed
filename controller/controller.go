package controller

import (
	"fmt"
	"github.com/cellargalaxy/go-file-bed/config"
	_ "github.com/cellargalaxy/go-file-bed/docs"
	"github.com/cellargalaxy/go-file-bed/resources"
	"github.com/cellargalaxy/go-file-bed/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"net/http"
)

var secretKey = "secret"
var secret = uuid.Must(uuid.NewV4()).String()

func Controller() error {
	indexHtmlString, err := resources.StaticBox.String("html/index.html")
	if err != nil {
		return err
	}

	store := cookie.NewStore([]byte(secret))

	engine := gin.Default()
	engine.Use(sessions.Sessions("session_id", store))

	engine.GET("/", func(context *gin.Context) {
		context.Header("Content-Type", "text/html; charset=utf-8")
		context.String(200, indexHtmlString)
	})
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

func validate(context *gin.Context) {
	if !isLogin(context) {
		context.Abort()
		context.JSON(http.StatusUnauthorized, createResponse(nil, fmt.Errorf("please login")))
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

// @Summary login
// @Param token formData string true "token"
// @Router /login [post]
// @Accept multipart/form-data
// @Success 200 {string} json
func login(context *gin.Context) {
	token := context.Request.FormValue("token")
	logrus.Info("用户登录")

	if service.CheckToken(token) {
		setLogin(context)
		context.JSON(http.StatusOK, createResponse("login success", nil))
	} else {
		logrus.WithFields(logrus.Fields{"token": token}).Info("非法token")
		context.JSON(http.StatusOK, createResponse(nil, fmt.Errorf("illegal token")))
	}
}

// @Summary uploadUrl
// @Param filePath formData string true "filePath"
// @Param url formData string true "url"
// @Router /admin/uploadUrl [post]
// @Accept multipart/form-data
// @Success 200 {string} json
func uploadUrl(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	url := context.Request.FormValue("url")
	logrus.WithFields(logrus.Fields{"filePath": filePath, "url": url}).Info("上传url文件")

	context.JSON(http.StatusOK, createResponse(service.AddUrl(filePath, url)))
}

// @Summary uploadFile
// @Param filePath formData string true "filePath"
// @Param file formData file true "file"
// @Router /admin/uploadFile [post]
// @Accept multipart/form-data
// @Success 200 {string} json
func uploadFile(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	file, header, err := context.Request.FormFile("file")
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("读取表单文件失败")
		context.JSON(http.StatusOK, createResponse(nil, err))
		return
	}
	defer file.Close()
	logrus.WithFields(logrus.Fields{"filePath": filePath, "filename": header.Filename}).Info("上传文件")

	context.JSON(http.StatusOK, createResponse(service.AddFile(filePath, file)))
}

// @Summary removeFile
// @Param filePath formData string true "filePath"
// @Router /admin/removeFile [post]
// @Accept multipart/form-data
// @Success 200 {string} json
func removeFile(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	logrus.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件")

	context.JSON(http.StatusOK, createResponse(service.RemoveFile(filePath)))
}

// @Summary getFileCompleteInfo
// @Param fileOrFolderPath query string true "fileOrFolderPath"
// @Router /admin/getFileCompleteInfo [get]
// @Accept multipart/form-data
// @Success 200 {string} json
func getFileCompleteInfo(context *gin.Context) {
	fileOrFolderPath := context.Query("fileOrFolderPath")
	logrus.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("查询完整文件信息")

	context.JSON(http.StatusOK, createResponse(service.GetFileCompleteInfo(fileOrFolderPath)))
}

// @Summary listLastFileInfo
// @Router /admin/listLastFileInfo [get]
// @Accept multipart/form-data
// @Success 200 {string} json
func listLastFileInfo(context *gin.Context) {
	context.JSON(http.StatusOK, createResponse(service.ListLastFileInfos()))
}

// @Summary listFolderInfo
// @Param folderPath query string false "folderPath"
// @Router /admin/listFolderInfo [get]
// @Accept multipart/form-data
// @Success 200 {string} json
func listFolderInfo(context *gin.Context) {
	folderPath := context.Query("folderPath")
	logrus.WithFields(logrus.Fields{"folderPath": folderPath}).Info("查询文件")

	context.JSON(http.StatusOK, createResponse(service.ListFolderInfo(folderPath)))
}

// @Summary listAllFileSimpleInfo
// @Router /admin/listAllFileSimpleInfo [get]
// @Accept multipart/form-data
// @Success 200 {string} json
func listAllFileSimpleInfo(context *gin.Context) {
	logrus.WithFields(logrus.Fields{}).Info("查询所有文件")

	context.JSON(http.StatusOK, createResponse(service.ListAllFileSimpleInfo()))
}

// @Summary receivePushSyncFile
// @Param filePath formData string true "filePath"
// @Param md5 formData string true "md5"
// @Param file formData file  true "file"
// @Router /admin/receivePushSyncFile [post]
// @Accept multipart/form-data
// @Success 200 {string} json
func receivePushSyncFile(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	md5 := context.Request.FormValue("md5")
	file, header, err := context.Request.FormFile("file")
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("读取表单文件失败")
		context.JSON(http.StatusOK, createResponse(nil, err))
		return
	}
	defer file.Close()
	logrus.WithFields(logrus.Fields{"filePath": filePath, "md5": md5, "filename": header.Filename}).Info("接收推送同步文件")

	context.JSON(http.StatusOK, createResponse(service.ReceivePushSyncFile(filePath, md5, file)))
}

// @Summary pushSyncFile
// @Param pushSyncHost formData string true "pushSyncHost"
// @Param token formData string true "token"
// @Router /admin/pushSyncFile [post]
// @Accept multipart/form-data
// @Success 200 {string} json
func pushSyncFile(context *gin.Context) {
	pushSyncHost := context.Request.FormValue("pushSyncHost")
	token := context.Request.FormValue("token")
	logrus.WithFields(logrus.Fields{"pushSyncHost": pushSyncHost}).Info("推送同步文件")

	context.JSON(http.StatusOK, createResponse(service.PushSyncFile(pushSyncHost, token)))
}

// @Summary pullSyncFile
// @Param pullSyncHost formData string true "pullSyncHost"
// @Param token formData string true "token"
// @Router /admin/pullSyncFile [post]
// @Accept multipart/form-data
// @Success 200 {string} json
func pullSyncFile(context *gin.Context) {
	pullSyncHost := context.Request.FormValue("pullSyncHost")
	token := context.Request.FormValue("token")
	logrus.WithFields(logrus.Fields{"pullSyncHost": pullSyncHost}).Info("拉取同步文件")

	context.JSON(http.StatusOK, createResponse(service.PullSyncFile(pullSyncHost, token)))
}

func createResponse(data interface{}, err error) map[string]interface{} {
	if err == nil {
		return gin.H{"code": config.SuccessCode, "message": nil, "data": data}
	} else {
		return gin.H{"code": config.FailCode, "message": err.Error(), "data": data}
	}
}
