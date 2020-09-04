package controller

import (
	"fmt"
	"github.com/cellargalaxy/go-file-bed/config"
	_ "github.com/cellargalaxy/go-file-bed/docs"
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
	store := cookie.NewStore([]byte(secret))

	engine := gin.Default()
	engine.Use(sessions.Sessions("session_id", store))

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

const indexHtmlString = `<!DOCTYPE html>
<html lang="en" xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
<head>
    <meta charset="UTF-8">
    <link type="text/css" rel="stylesheet" href="//unpkg.com/bootstrap/dist/css/bootstrap.min.css"/>
    <link type="text/css" rel="stylesheet" href="//unpkg.com/bootstrap-vue@latest/dist/bootstrap-vue.min.css"/>
    <title>file bed</title>
</head>
<body>
<div class="container">

    <b-input-group id="loginForm">
        <b-form-input size="sm" type="password" placeholder="token" v-model="token"></b-form-input>
        <b-button size="sm" variant="outline-primary" :disabled="loading" @click="login">login</b-button>
    </b-input-group>
    <br/>
    <form id="uploadFileForm">
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="sort" v-model="sort" @input="input"></b-form-input>
            <b-form-input size="sm" type="text" placeholder="date" v-model="date" @input="input"></b-form-input>
        </b-input-group>
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="filePath" v-model="filePath"></b-form-input>
            <b-button size="sm" variant="outline-primary" :disabled="loading" @click="uploads">upload</b-button>
        </b-input-group>
        <b-input-group>
            <b-form-file multiple size="sm" v-model="files" @input="input"></b-form-file>
        </b-input-group>
    </form>
    <br/>
    <form id="uploadUrlForm">
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="sort" v-model="sort" @input="input"></b-form-input>
            <b-form-input size="sm" type="text" placeholder="date" v-model="date" @input="input"></b-form-input>
        </b-input-group>
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="filePath" v-model="filePath"></b-form-input>
            <b-button size="sm" variant="outline-primary" :disabled="loading" @click="uploads">upload</b-button>
        </b-input-group>
        <b-input-group>
            <b-form-textarea size="sm" placeholder="url" v-model="urls" :rows="row" @input="input"></b-form-textarea>
        </b-input-group>
    </form>
    <br/>
    <b-table id="lastFileInfoTable" stacked="xl" striped hover responsive small
             :fields="fields" :items="infos" :busy="loading">
        <template v-slot:cell(name)="data">
            <code>{{data.item.name}}</code>
        </template>
        <template v-slot:cell(md5)="data">
            <code>{{data.item.md5}}</code>
        </template>
        <template v-slot:cell(url)="data">
            <b-button-group>
                <b-button size="sm" variant="outline-success" @click="copyShort(data.item)">short</b-button>
                <b-button size="sm" variant="outline-primary" @click="copyLong(data.item)">long</b-button>
                <b-button size="sm" variant="outline-warning" @click="copyMarkdown(data.item)">markdown</b-button>
            </b-button-group>
        </template>
        <template v-slot:cell(deal)="data">
            <b-button-group>
                <b-button size="sm" variant="outline-success" @click="openFile(data.item.path)">open</b-button>
                <b-button size="sm" variant="outline-primary" :disabled="data.item.loading"
                          @click="info(data.item.path)">info
                </b-button>
                <b-button size="sm" variant="outline-danger" :disabled="data.item.loading"
                          @click="deleteFile(data.item.path)">delete
                </b-button>
            </b-button-group>
        </template>

        <template v-slot:table-busy>
            <div class="text-center text-primary">
                <b-spinner class="align-middle"></b-spinner>
                <strong>Loading...</strong>
            </div>
        </template>
    </b-table>
    <br/>
    <form id="pullSyncForm">
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="pullSyncHost" v-model="pullSyncHost"></b-form-input>
            <b-form-input size="sm" type="password" placeholder="token" v-model="token"></b-form-input>
            <b-button size="sm" variant="outline-primary" :disabled="loading" @click="pull">pull</b-button>
        </b-input-group>
    </form>
    <form id="pushSyncForm">
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="pushSyncHost" v-model="pushSyncHost"></b-form-input>
            <b-form-input size="sm" type="password" placeholder="token" v-model="token"></b-form-input>
            <b-button size="sm" variant="outline-primary" :disabled="loading" @click="push">push</b-button>
        </b-input-group>
    </form>
    <br/>
    <b-breadcrumb id="filePathBreadcrumb">
        <b-breadcrumb-item @click="go(-1)">
            <b-icon icon="house-fill"></b-icon>
        </b-breadcrumb-item>
        <b-breadcrumb-item v-for="(breadcrumb, index) in breadcrumbs" @click="go(index)">
            {{breadcrumb.name}}
        </b-breadcrumb-item>
    </b-breadcrumb>
    <b-table id="fileInfoTable" stacked="xl" striped hover responsive small
             :fields="fields" :items="infos" :busy="loading">
        <template v-slot:cell(name)="data">
            <code>{{data.item.name}}</code>
        </template>
        <template v-slot:cell(md5)="data">
            <code>{{data.item.md5}}</code>
        </template>
        <template v-slot:cell(url)="data">
            <b-button-group v-if="data.item.isFile">
                <b-button size="sm" variant="outline-success" @click="copyShort(data.item)">short</b-button>
                <b-button size="sm" variant="outline-primary" @click="copyLong(data.item)">long</b-button>
                <b-button size="sm" variant="outline-warning" @click="copyMarkdown(data.item)">markdown</b-button>
            </b-button-group>
        </template>
        <template v-slot:cell(deal)="data">
            <b-button-group>
                <b-button size="sm" variant="outline-success" @click="openFile(data.item.path)">open</b-button>
                <b-button size="sm" variant="outline-primary" :disabled="data.item.loading"
                          @click="info(data.item.path)">info
                </b-button>
                <b-button size="sm" variant="outline-danger" :disabled="data.item.loading"
                          @click="deleteFile(data.item.path)">delete
                </b-button>
            </b-button-group>
        </template>

        <template v-slot:table-busy>
            <div class="text-center text-primary">
                <b-spinner class="align-middle"></b-spinner>
                <strong>Loading...</strong>
            </div>
        </template>
    </b-table>


</div>
</body>
<script src="//vuejs.org/js/vue.min.js"></script>
<script src="//unpkg.com/bootstrap-vue@latest/dist/bootstrap-vue.min.js"></script>
<script src="//unpkg.com/bootstrap-vue@latest/dist/bootstrap-vue-icons.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/qs/6.9.3/qs.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/axios/0.19.2/axios.min.js"></script>
<script>
    window.onbeforeunload = (event) => 'maybe some data not save'
    var instance = axios.create({timeout: 60 * 60 * 1000})

    const loginFormVue = new Vue({
        el: '#loginForm',
        data: {
            token: null,
            loading: false
        },
        methods: {
            login() {
                if (this.token == null || this.token === '') {
                    alert('token为空')
                    return
                }
                this.loading = true
                instance.post("login", Qs.stringify({token: this.token}))
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        alert(result.code === 1 ? '登录成功' : '登录失败')
                        if (result.code === 1) {
                            this.init()
                            this.setLogin()
                            flush()
                        }
                    })
                    .catch(error => {
                        this.loading = false
                        alert("error: " + JSON.stringify(error))
                    })
            },
            setLogin() {
                setCookie('login', 'login')
            },
            getLogin() {
                return getCookie('login') === 'login'
            },
            init() {
                this.token = null
            }
        }
    })

    const uploadFileVue = new Vue({
        el: '#uploadFileForm',
        data: {
            files: [],
            sort: '',
            date: formatDate(new Date(), 'YYYYMMDD'),
            filePath: null,
            loading: false
        },
        methods: {
            uploads() {
                if (this.files === undefined || this.files == null || this.files.length === 0) {
                    alert('还没有选择文件')
                    return
                }
                for (let i = 0; i < this.files.length; i++) {
                    if (this.files[i] === undefined || this.files[i] == null) {
                        continue
                    }
                    this.upload(this.files[i])
                }
            },
            upload(file) {
                if (file === undefined || file == null) {
                    alert('文件或为空')
                    return
                }
                const filePath = createFilePath(this.sort, this.date, file.name)
                if (filePath === undefined || filePath == null || filePath === '') {
                    alert('文件路径为空')
                    return
                }
                this.loading = true
                const param = new FormData()
                param.append("filePath", filePath)
                param.append("file", file)
                instance.post("admin/uploadFile", param, {headers: {'Content-Type': 'multipart/form-data'}})
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        if (result.code === 1) {
                            this.init()
                            flush()
                        } else {
                            alert(result.message)
                        }
                    })
                    .catch(error => {
                        this.loading = false
                        alert("error: " + JSON.stringify(error))
                    })
            },
            input() {
                if (this.sort != null) {
                    this.sort = this.sort.replace(/\s/g, '');
                }
                if (this.files !== undefined && this.files != null && this.files.length > 0) {
                    this.filePath = createFilePath(this.sort, this.date, this.files[0].name)
                }
            },
            init() {
                this.files = []
                this.filePath = null
            }
        }
    })

    const uploadUrlVue = new Vue({
        el: '#uploadUrlForm',
        data: {
            urls: '',
            sort: '',
            date: formatDate(new Date(), 'YYYYMMDD'),
            filePath: null,
            loading: false
        },
        computed: {
            row() {
                let row = this.urls.split('\n').length
                row = row <= 0 ? 3 : row
                return row
            }
        },
        methods: {
            uploads() {
                if (this.urls === undefined || this.urls == null || this.urls === '') {
                    alert('URL为空')
                    return
                }
                const urls = this.urls.split('\n')
                for (let i = 0; i < urls.length; i++) {
                    if (urls[i] === undefined || urls[i] == null || urls[i] === '') {
                        continue
                    }
                    this.upload(urls[i])
                }
            },
            upload(url) {
                if (url === undefined || url == null || url === '') {
                    alert('URL为空')
                    return
                }
                let filename = url.split('//')
                filename = filename[filename.length - 1]
                filename = filename.split('?')
                filename = filename[0].replace(/:/g, '_').replace(/\//g, '-').replace(/\\/g, '-')
                const filePath = createFilePath(this.sort, this.date, filename)
                if (filePath === undefined || filePath == null || filePath === '') {
                    alert('文件路径为空')
                    return
                }
                this.loading = true
                instance.post("admin/uploadUrl", Qs.stringify({filePath: filePath, url: url}))
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        if (result.code === 1) {
                            this.init()
                            flush()
                        } else {
                            alert(result.message)
                        }
                    })
                    .catch(error => {
                        this.loading = false
                        alert("error: " + JSON.stringify(error))
                    })
            },
            input() {
                if (this.sort != null) {
                    this.sort = this.sort.replace(/\s/g, '');
                }
                if (this.urls !== undefined && this.urls != null && this.urls !== '') {
                    this.urls = this.urls + '\n'

                    const url = this.urls.split('\n')[0]
                    let filename = url.split('//')
                    filename = filename[filename.length - 1]
                    filename = filename.split('?')
                    filename = filename[0].replace(/:/g, '_').replace(/\//g, '-').replace(/\\/g, '-')
                    this.filePath = createFilePath(this.sort, this.date, filename)
                }
            },
            init() {
                this.urls = ''
                this.filePath = null
            }
        }
    })

    const lastFileInfoVue = new Vue({
        el: '#lastFileInfoTable',
        data: {
            fields: [
                {key: 'name', label: 'name', sortable: true,},
                {key: 'size', label: 'size', sortable: true,},
                {key: 'count', label: 'count', sortable: true,},
                {key: 'md5', label: 'md5',},
                {key: 'url', label: 'url',},
                {key: 'deal', label: 'deal',},
            ],
            infos: [],
            loading: false,
        },
        methods: {
            listLastFileInfo() {
                this.loading = true
                instance.get("admin/listLastFileInfo", {params: {}})
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        if (result.code === 1) {
                            if (result.data == null || result.data.length === 0) {
                                alert('没有最新文件')
                            }
                            this.infos = initFileInfos(result.data)
                        } else {
                            alert(result.message)
                        }
                    })
                    .catch(error => {
                        this.loading = false
                        alert("error: " + JSON.stringify(error))
                    })
            },
            openFile(path) {
                let index = 0
                for (; index < this.infos.length; index++) {
                    if (this.infos[index].path === path) {
                        break
                    }
                }
                if (index === this.infos.length) {
                    alert('非法打开文件的路径: ' + path)
                    return
                }
                if (this.infos[index].isFile) {
                    window.open(this.infos[index].url)
                } else {
                    alert('明明是最近文件，但居然是文件夹？！？')
                }
            },
            info(path) {
                let index = 0
                for (; index < this.infos.length; index++) {
                    if (this.infos[index].path === path) {
                        break
                    }
                }
                if (index === this.infos.length) {
                    alert('非法查询文件的路径: ' + path)
                    return
                }
                getFileCompleteInfo(this.infos, index)
            },
            deleteFile(path) {
                let index = 0
                for (; index < this.infos.length; index++) {
                    if (this.infos[index].path === path) {
                        break
                    }
                }
                if (index === this.infos.length) {
                    alert('非法删除文件的路径: ' + path)
                    return
                }
                removeFile(this.infos, index)
            },
            copyShort(file) {
                writeClipboard('/' + file.url)
            },
            copyLong(file) {
                writeClipboard(document.location.origin + '/' + file.url)
            },
            copyMarkdown(file) {
                writeClipboard('![' + file.name + '](' + document.location.origin + '/' + file.url + ')')
            },
        }
    })

    const pullSyncVue = new Vue({
        el: '#pullSyncForm',
        data: {
            pullSyncHost: null,
            token: null,
            loading: false
        },
        methods: {
            pull() {
                if (this.pullSyncHost == null || this.pullSyncHost === '' || this.token == null || this.token === '') {
                    alert('pull同步URL或者token为空')
                    return
                }
                this.loading = true
                instance.post("admin/pullSyncFile", Qs.stringify({pullSyncHost: this.pullSyncHost, token: this.token}))
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        alert('失败数量: ' + result.data + ', 失败原因: ' + result.message)
                        if (result.code === 1) {
                            this.init()
                            flush()
                        }
                    })
                    .catch(error => {
                        this.loading = false
                        alert("error: " + JSON.stringify(error))
                    })
            },
            init() {
                this.pullSyncHost = null
                this.token = null
            }
        }
    })

    const pushSyncVue = new Vue({
        el: '#pushSyncForm',
        data: {
            pushSyncHost: null,
            token: null,
            loading: false
        },
        methods: {
            push() {
                if (this.pushSyncHost == null || this.pushSyncHost === '' || this.token == null || this.token === '') {
                    alert('push同步URL或者token为空')
                    return
                }
                this.loading = true
                instance.post("admin/pushSyncFile", Qs.stringify({pushSyncHost: this.pushSyncHost, token: this.token}))
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        alert('失败数量: ' + result.data + ', 失败原因: ' + result.message)
                        if (result.code === 1) {
                            this.init()
                            flush()
                        }
                    })
                    .catch(error => {
                        this.loading = false
                        alert("error: " + JSON.stringify(error))
                    })
            },
            init() {
                this.pushSyncHost = null
                this.token = null
            }
        }
    })

    const filePathVue = new Vue({
        el: '#filePathBreadcrumb',
        data: {
            breadcrumbs: [],
        },
        methods: {
            go(index) {
                if (index === -1) {
                    fileInfoVue.init('/')
                    return
                }
                fileInfoVue.init(this.breadcrumbs[index].path)
            },
            init(path) {
                this.breadcrumbs = []
                const names = path.split('/')
                let filePath = ''
                for (let i = 0; i < names.length; i++) {
                    if (names[i] === '') {
                        continue
                    }
                    filePath = filePath + '/' + names[i]
                    this.breadcrumbs.push({name: names[i], path: filePath})
                }
            }
        }
    })

    const fileInfoVue = new Vue({
        el: '#fileInfoTable',
        data: {
            fields: [
                {key: 'name', label: 'name', sortable: true,},
                {key: 'size', label: 'size', sortable: true,},
                {key: 'count', label: 'count', sortable: true,},
                {key: 'md5', label: 'md5',},
                {key: 'url', label: 'url',},
                {key: 'deal', label: 'deal',},
            ],
            folderPath: '',
            infos: [],
            loading: false,
        },
        methods: {
            listFolderInfo() {
                this.loading = true
                instance.get("admin/listFolderInfo", {params: {folderPath: this.folderPath}})
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        if (result.code === 1) {
                            if (result.data == null || result.data.length === 0) {
                                alert('此目录下没有文件')
                            }
                            this.infos = initFileInfos(result.data)
                            filePathVue.init(this.folderPath)
                        } else {
                            alert(result.message)
                        }
                    })
                    .catch(error => {
                        this.loading = false
                        alert("error: " + JSON.stringify(error))
                    })
            },
            init(path) {
                this.folderPath = path
                this.listFolderInfo()
            },
            openFile(path) {
                let index = 0
                for (; index < this.infos.length; index++) {
                    if (this.infos[index].path === path) {
                        break
                    }
                }
                if (index === this.infos.length) {
                    alert('非法打开文件的路径: ' + path)
                    return
                }
                if (this.infos[index].isFile) {
                    window.open(this.infos[index].url)
                } else {
                    this.init(this.infos[index].path)
                }
            },
            info(path) {
                let index = 0
                for (; index < this.infos.length; index++) {
                    if (this.infos[index].path === path) {
                        break
                    }
                }
                if (index === this.infos.length) {
                    alert('非法查询文件的路径: ' + path)
                    return
                }
                getFileCompleteInfo(this.infos, index)
            },
            deleteFile(path) {
                let index = 0
                for (; index < this.infos.length; index++) {
                    if (this.infos[index].path === path) {
                        break
                    }
                }
                if (index === this.infos.length) {
                    alert('非法删除文件的路径: ' + path)
                    return
                }
                removeFile(this.infos, index)
            },
            copyShort(file) {
                writeClipboard('/' + file.url)
            },
            copyLong(file) {
                writeClipboard(document.location.origin + '/' + file.url)
            },
            copyMarkdown(file) {
                writeClipboard('![' + file.name + '](' + document.location.origin + '/' + file.url + ')')
            },
        }
    })

    function removeFile(infos, index) {
        if (!infos[index].isFile) {
            alert("所删除不是文件")
            return
        }
        if (!confirm("确实删除文件？！？: " + infos[index].path)) {
            return
        }
        infos[index].loading = true
        instance.post("admin/removeFile", Qs.stringify({filePath: infos[index].path}))
            .then(response => {
                infos[index].loading = false
                const result = response.data
                alert(result.code === 1 ? '删除成功' : '删除失败')
                if (result.code === 1) {
                    flush()
                }
            })
            .catch(error => {
                infos[index].loading = false
                alert("error: " + JSON.stringify(error))
            })
    }

    function getFileCompleteInfo(infos, index) {
        infos[index].loading = true
        instance.get("admin/getFileCompleteInfo", {params: {fileOrFolderPath: infos[index].path}})
            .then(response => {
                infos[index].loading = false
                const result = response.data
                if (result.code === 1) {
                    const info = initFileInfo(result.data)
                    infos[index].size = info.size
                    infos[index].count = info.count
                    infos[index].md5 = info.md5
                } else {
                    alert(result.message)
                }
            })
            .catch(error => {
                infos[index].loading = false
                alert("error: " + JSON.stringify(error))
            })
    }

    function flush() {
        lastFileInfoVue.listLastFileInfo()
        fileInfoVue.listFolderInfo()
    }

    function initFileInfos(infos) {
        if (infos == null) {
            return []
        }
        for (let i = 0; i < infos.length; i++) {
            infos[i] = initFileInfo(infos[i])
        }
        return infos
    }

    function initFileInfo(info) {
        if (info == null) {
            return null
        }
        info.loading = false
        info.url = encodeURI(info.url)
        if (info.size !== undefined) {
            info.size = formatFileSize(info.size)
        }
        if (info.url.startsWith('/')) {
            info.url = info.url.substring(1)
        }
        return info
    }

    function writeClipboard(text) {
        const textarea = document.createElement('textarea')
        textarea.style.opacity = 0
        textarea.style.position = 'absolute'
        textarea.style.left = '-100000px'
        document.body.appendChild(textarea)

        textarea.value = text
        textarea.select()
        textarea.setSelectionRange(0, text.length)
        document.execCommand('copy')
        document.body.removeChild(textarea)
    }

    function formatFileSize(size) {
        if (size < 0) return '非法大小: ' + size
        if (size === 0) return '0B'
        var s = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
        var e = Math.floor(Math.log(size) / Math.log(1024))
        return (size / Math.pow(1024, Math.floor(e))).toFixed(2) + '' + s[e]
    }

    function createFilePath(sort, date, filename) {
        let filePath = filename
        if (date != null && date !== '') {
            filePath = date + '/' + filePath
        }
        if (sort != null && sort !== '') {
            filePath = sort + '/' + filePath
        }
        return filePath
    }

    function formatDate(date, fmt) {
        let o = {
            'M+': date.getMonth() + 1, //月份
            'D+': date.getDate(), //日
            'H+': date.getHours(), //小时
            'm+': date.getMinutes(), //分
            's+': date.getSeconds(), //秒
            'q+': Math.floor((date.getMonth() + 3) / 3), //季度
            'S': date.getMilliseconds() //毫秒
        }
        if (/(Y+)/.test(fmt)) {
            fmt = fmt.replace(RegExp.$1, (date.getFullYear() + '').substr(4 - RegExp.$1.length))
        }
        for (let k in o) {
            if (new RegExp('(' + k + ')').test(fmt)) {
                fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1) ? (o[k]) : (('00' + o[k]).substr(('' + o[k]).length)))
            }
        }
        return fmt
    }

    function getCookieFromString(cookieString, name) {
        if (cookieString === null) {
            return null
        }
        let nameEQ = name + '='
        let ca = cookieString.split(';')
        for (let i = 0; i < ca.length; i++) {
            let c = ca[i]
            while (c.charAt(0) === ' ') c = c.substring(1, c.length)
            if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length)
        }
        return null
    }

    function getCookie(name) {
        return getCookieFromString(document.cookie, name)
    }

    function setCookie(key, value) {
        let date = new Date()
        date.setTime(date.getTime() + (1000 * 60 * 60))
        document.cookie = key + '=' + value + '; expires=' + date.toGMTString()
    }

    if (loginFormVue.getLogin()) {
        flush()
    }
</script>
</html>`
