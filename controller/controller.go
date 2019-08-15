package controller

import (
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

func Controller() {
	store := cookie.NewStore([]byte(secret))

	engine := gin.Default()
	engine.Use(sessions.Sessions("session_id", store))
	engine.LoadHTMLGlob("templates/*")

	engine.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", gin.H{})
	})
	engine.POST("/login", loginController)
	engine.Static("/file", service.FileBedPath)

	adminGroup := engine.Group("/admin", validate)
	adminGroup.POST("/uploadFile", uploadFileController)
	adminGroup.POST("/removeFile", removeFileController)
	adminGroup.GET("/listFileOrFolderInfo", listFileOrFolderInfoController)

	engine.Run("0.0.0.0:8880")
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
	return true
	session := sessions.Default(context)
	sessionSecret := session.Get(secretKey)
	return sessionSecret == secret
}

func loginController(context *gin.Context) {
	token := context.Param("token")
	log.Info("用户登录")

	if service.CheckToken(token) {
		setLogin(context)
		context.JSON(http.StatusOK, createSuccessResponse("login success", nil))
	} else {
		context.JSON(http.StatusOK, createFailResponse("illegal token"))
	}
}

func removeFileController(context *gin.Context) {
	filePath := context.Param("filePath")
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件")

	err := service.RemoveFile(filePath)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("remove file success", nil))
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

	err = service.AddFile(sort, filename, file)
	if err != nil {
		context.JSON(http.StatusOK, createFailResponse(err.Error()))
	} else {
		context.JSON(http.StatusOK, createSuccessResponse("upload file success", nil))
	}
}

func createSuccessResponse(massage string, data interface{}) map[string]interface{} {
	return createResponse(0, massage, data)
}
func createFailResponse(massage string) map[string]interface{} {
	return createResponse(1, massage, nil)
}
func createResponse(code int, massage string, data interface{}) map[string]interface{} {
	return gin.H{"code": code, "massage": massage, "data": data}
}
