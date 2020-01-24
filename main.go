package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

//struct================================================================================================================

const SuccessCode = 1
const FailCode = 2

const FileUrl = "/file"
const LoginUrl = "/login"

const UploadUrlUrl = "/admin/uploadUrl"
const UploadFileUrl = "/admin/uploadFile"
const RemoveFileUrl = "/admin/removeFile"
const GetFileCompleteInfoUrl = "/admin/getFileCompleteInfo"
const ListLastFileInfoUrl = "/admin/listLastFileInfo"
const ListFolderInfoUrl = "/admin/listFolderInfo"
const ListAllFileInfoUrl = "/admin/listAllFileInfo"
const ReceivePushSynFileUrl = "/admin/receivePushSynFile"
const PushSynFileUrl = "/admin/pushSynFile"
const PullSynFileUrl = "/admin/pullSynFile"

type FileSimpleInfo struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Mime string `json:"mime"`
	Url  string `json:"url"`
}

type FileCompleteInfo struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Mime  string `json:"mime"`
	Url   string `json:"url"`
	Size  int64  `json:"size"`
	Count int32  `json:"count"`
	Md5   string `json:"md5"`
}

//config================================================================================================================

const defaultToken string = "token"
const defaultListenAddress string = ":8880"
const defaultFileBedPath string = "file_bed"
const defaultLastFileInfoCount = 10

var log = logrus.New()

var Token string
var ListenAddress string
var FileBedPath string

var LastFileInfoCount int
var timeout = 5 * time.Second
var pullOrPushTimeout = 60 * 60 * time.Second

func init() {
	log.Info("加载配置")

	token := os.Getenv("TOKEN")
	log.WithFields(logrus.Fields{"token": len(token)}).Info("环境变量读取token")
	if token == "" {
		token = defaultToken
	}
	Token = token
	log.WithFields(logrus.Fields{"token": len(Token)}).Info("配置token")

	listenAddress := os.Getenv("LISTEN_ADDRESS")
	log.WithFields(logrus.Fields{"listenAddress": listenAddress}).Info("环境变量读取listenAddress")
	if listenAddress == "" {
		listenAddress = defaultListenAddress
	}
	ListenAddress = listenAddress
	log.WithFields(logrus.Fields{"listenAddress": ListenAddress}).Info("配置token")

	fileBedPath := os.Getenv("FILE_BED_PATH")
	log.WithFields(logrus.Fields{"fileBedPath": fileBedPath}).Info("环境变量读取fileBedPath")
	if fileBedPath == "" {
		fileBedPath = defaultFileBedPath
	}
	FileBedPath = fileBedPath
	log.WithFields(logrus.Fields{"fileBedPath": FileBedPath}).Info("配置token")

	lastFileInfoCountString := os.Getenv("LAST_FILE_INFO_COUNT")
	lastFileInfoCount, err := strconv.Atoi(lastFileInfoCountString)
	log.WithFields(logrus.Fields{"lastFileInfoCountString": lastFileInfoCountString, "lastFileInfoCount": lastFileInfoCount, "err": err}).Info("环境变量读取lastFileInfoCount")
	if err != nil || lastFileInfoCount <= 0 {
		lastFileInfoCount = defaultLastFileInfoCount
	}
	LastFileInfoCount = lastFileInfoCount
	log.WithFields(logrus.Fields{"lastFileInfoCount": lastFileInfoCount}).Info("配置token")

	log.Info("加载配置成功")

	err = os.MkdirAll(FileBedPath, 0666)
	if err != nil {
		log.WithFields(logrus.Fields{"folderPath": FileBedPath, "err": err}).Error("创建文件夹失败")
	}
}

//main==================================================================================================================

func main() {
	Controller()
}

//controller============================================================================================================

var secretKey = "secret"
var secret = uuid.Must(uuid.NewV4()).String()

func Controller() {
	log.WithFields(logrus.Fields{"fileBedPath": FileBedPath}).Info("文件床路径")
	log.WithFields(logrus.Fields{"listenAddress": ListenAddress}).Info("监听地址")

	store := cookie.NewStore([]byte(secret))

	engine := gin.Default()
	engine.Use(sessions.Sessions("session_id", store))

	engine.GET("/", func(context *gin.Context) {
		context.Header("Content-Type", "text/html; charset=utf-8")
		context.String(200, indexHtmlString)
	})
	engine.Static(FileUrl, FileBedPath)
	engine.POST(LoginUrl, loginController)

	engine.POST(UploadUrlUrl, validate, uploadUrlController)
	engine.POST(UploadFileUrl, validate, uploadFileController)
	engine.POST(RemoveFileUrl, validate, removeFileController)
	engine.GET(GetFileCompleteInfoUrl, validate, getFileCompleteInfoController)
	engine.GET(ListLastFileInfoUrl, validate, listLastFileInfoController)
	engine.GET(ListFolderInfoUrl, validate, listFolderInfoController)
	engine.GET(ListAllFileInfoUrl, validate, listAllFileInfoController)
	engine.POST(ReceivePushSynFileUrl, validate, receivePushSynFileController)
	engine.POST(PushSynFileUrl, validate, pushSynFileController)
	engine.POST(PullSynFileUrl, validate, pullSynFileController)

	engine.Run(ListenAddress)
}

func validate(context *gin.Context) {
	if !isLogin(context) {
		context.Abort()
		context.JSON(http.StatusUnauthorized, createResponse(nil, errors.New("please login")))
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

	if CheckToken(token) {
		setLogin(context)
		context.JSON(http.StatusOK, createResponse("login success", nil))
	} else {
		log.WithFields(logrus.Fields{"token": token}).Info("非法token")
		context.JSON(http.StatusOK, createResponse(nil, errors.New("illegal token")))
	}
}

func uploadUrlController(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	url := context.Request.FormValue("url")
	log.WithFields(logrus.Fields{"filePath": filePath, "url": url}).Info("上传url文件")

	context.JSON(http.StatusOK, createResponse(AddUrl(filePath, url)))
}

func uploadFileController(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	file, header, err := context.Request.FormFile("file")
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("去读表单文件失败")
		context.JSON(http.StatusOK, createResponse(nil, err))
		return
	}
	defer file.Close()
	log.WithFields(logrus.Fields{"filePath": filePath, "filename": header.Filename}).Info("上传文件")

	context.JSON(http.StatusOK, createResponse(AddFile(filePath, file)))
}

func removeFileController(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("删除文件")

	context.JSON(http.StatusOK, createResponse(RemoveFile(filePath)))
}

func getFileCompleteInfoController(context *gin.Context) {
	fileOrFolderPath := context.Query("fileOrFolderPath")
	log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("查询完整文件")

	context.JSON(http.StatusOK, createResponse(GetFileCompleteInfo(fileOrFolderPath)))
}

func listLastFileInfoController(context *gin.Context) {
	context.JSON(http.StatusOK, createResponse(ListLastFileInfos()))
}

func listFolderInfoController(context *gin.Context) {
	folderPath := context.Query("folderPath")
	log.WithFields(logrus.Fields{"folderPath": folderPath}).Info("查询文件")

	context.JSON(http.StatusOK, createResponse(ListFolderInfo(folderPath)))
}

func listAllFileInfoController(context *gin.Context) {
	folderPath := context.Query("folderPath")
	log.WithFields(logrus.Fields{"folderPath": folderPath}).Info("查询文件夹下所有文件")

	context.JSON(http.StatusOK, createResponse(ListAllFileInfo(folderPath)))
}

func receivePushSynFileController(context *gin.Context) {
	filePath := context.Request.FormValue("filePath")
	md5 := context.Request.FormValue("md5")
	file, header, err := context.Request.FormFile("file")
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("去读表单文件失败")
		context.JSON(http.StatusOK, createResponse(nil, err))
		return
	}
	defer file.Close()
	log.WithFields(logrus.Fields{"filePath": filePath, "md5": md5, "filename": header.Filename}).Info("接收推送同步文件")

	context.JSON(http.StatusOK, createResponse(ReceivePushSynFile(filePath, md5, file)))
}

func pushSynFileController(context *gin.Context) {
	pushSynHost := context.Request.FormValue("pushSynHost")
	token := context.Request.FormValue("token")
	log.WithFields(logrus.Fields{"pushSynHost": pushSynHost}).Info("推送同步文件")

	context.JSON(http.StatusOK, createResponse(PushSynFile(pushSynHost, token)))
}

func pullSynFileController(context *gin.Context) {
	pullSynHost := context.Request.FormValue("pullSynHost")
	token := context.Request.FormValue("token")
	log.WithFields(logrus.Fields{"pullSynHost": pullSynHost}).Info("拉取同步文件")

	context.JSON(http.StatusOK, createResponse(PullSynFile(pullSynHost, token)))
}

func createResponse(data interface{}, err error) map[string]interface{} {
	if err == nil {
		return gin.H{"code": SuccessCode, "message": nil, "data": data}
	} else {
		return gin.H{"code": FailCode, "message": err.Error(), "data": data}
	}
}

var indexHtmlString = `<!DOCTYPE html>
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
            <b-form-file size="sm" v-model="file" @input="input"></b-form-file>
        </b-input-group>
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="filePath" v-model="filePath"></b-form-input>
            <b-button size="sm" variant="outline-primary" :disabled="loading" @click="upload">upload</b-button>
        </b-input-group>
    </form>
    <br/>
    <form id="uploadUrlForm">
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="sort" v-model="sort" @input="input"></b-form-input>
            <b-form-input size="sm" type="text" placeholder="url" v-model="url" @input="input"></b-form-input>
        </b-input-group>
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="filePath" v-model="filePath"></b-form-input>
            <b-button size="sm" variant="outline-primary" :disabled="loading" @click="upload">upload</b-button>
        </b-input-group>
    </form>
    <br/>
    <b-table id="lastFileInfoTable" stacked="xl" striped hover responsive small
             :fields="fields" :items="infos" :busy="loading">
        <template v-slot:cell(name)="data">
            <a href="javascript:;" @click="openFile(data.item.path)">{{data.item.name}}</a>
        </template>
        <template v-slot:cell(mime)="data">
            <code>{{data.item.mime}}</code>
        </template>
        <template v-slot:cell(md5)="data">
            <code>{{data.item.md5}}</code>
        </template>
        <template v-slot:cell(url)="data">
            <b-form-input size="sm" type="text" placeholder="url" :value="data.item.url"></b-form-input>
        </template>
        <template v-slot:cell(deal)="data">
            <b-button-group>
                <b-button size="sm" variant="outline-primary" :disabled="data.item.loading"
                          @click="info(data.item.path)">
                    info
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
            <b-form-input size="sm" type="text" placeholder="pullSynHost" v-model="pullSynHost"></b-form-input>
            <b-form-input size="sm" type="password" placeholder="token" v-model="token"></b-form-input>
            <b-button size="sm" variant="outline-primary" :disabled="loading" @click="pull">pull</b-button>
        </b-input-group>
    </form>
    <form id="pushSyncForm">
        <b-input-group>
            <b-form-input size="sm" type="text" placeholder="pushSynHost" v-model="pushSynHost"></b-form-input>
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
            <a href="javascript:;" @click="openFile(data.item.path)">{{data.item.name}}</a>
        </template>
        <template v-slot:cell(mime)="data">
            <code>{{data.item.mime}}</code>
        </template>
        <template v-slot:cell(md5)="data">
            <code>{{data.item.md5}}</code>
        </template>
        <template v-slot:cell(url)="data">
            <b-form-input size="sm" type="text" placeholder="url" v-if="data.item.isFile" :value="data.item.url">
            </b-form-input>
        </template>
        <template v-slot:cell(deal)="data">
            <b-button-group>
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
<script src="//cdn.bootcss.com/qs/6.8.0/qs.min.js"></script>
<script src="//cdn.bootcss.com/axios/0.19.0-beta.1/axios.min.js"></script>
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
                if (this.token == null || this.token == '') {
                    alert('token为空')
                    return
                }
                this.loading = true
                instance.post("login", Qs.stringify({token: this.token}))
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        alert(result.code == 1 ? '登录成功' : '登录失败')
                        if (result.code == 1) {
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
                return getCookie('login') == 'login'
            },
            init() {
                this.token = null
            }
        }
    })

    const uploadFileVue = new Vue({
        el: '#uploadFileForm',
        data: {
            file: null,
            sort: '',
            filePath: null,
            loading: false
        },
        methods: {
            upload() {
                if (this.file == null || this.filePath == null || this.filePath == '') {
                    alert('文件或者文件路径为空')
                    return
                }
                this.loading = true
                const param = new FormData()
                param.append("filePath", this.filePath)
                param.append("file", this.file)
                instance.post("admin/uploadFile", param, {headers: {'Content-Type': 'multipart/form-data'}})
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        if (result.code == 1) {
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
                if (this.file != null) {
                    this.filePath = createFilePath(this.sort, this.file.name)
                }
            },
            init() {
                this.file = null
                this.sort = ''
                this.filePath = null
            }
        }
    })

    const uploadUrlVue = new Vue({
        el: '#uploadUrlForm',
        data: {
            url: null,
            sort: '',
            filePath: null,
            loading: false
        },
        methods: {
            upload() {
                if (this.url == null || this.url == '' || this.filePath == null || this.filePath == '') {
                    alert('URL或者文件路径为空')
                    return
                }
                this.loading = true
                instance.post("admin/uploadUrl", Qs.stringify({filePath: this.filePath, url: this.url}))
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        if (result.code == 1) {
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
                if (this.url != null) {
                    let filename = this.url.split('//')
                    filename = filename[filename.length - 1]
                    filename = filename.split('?')
                    filename = filename[0].replace(/:/g, '_').replace(/\//g, '-').replace(/\\/g, '-')
                    this.filePath = createFilePath(this.sort, filename)
                }
            },
            init() {
                this.url = null
                this.sort = ''
                this.filePath = null
            }
        }
    })

    const lastFileInfoVue = new Vue({
        el: '#lastFileInfoTable',
        data: {
            fields: [
                {
                    key: 'name',
                    label: 'name',
                    sortable: true,
                },
                {
                    key: 'mime',
                    label: 'mime',
                    sortable: true,
                },
                {
                    key: 'size',
                    label: 'size',
                    sortable: true,
                },
                {
                    key: 'count',
                    label: 'count',
                    sortable: true,
                },
                {
                    key: 'md5',
                    label: 'md5',
                },
                {
                    key: 'url',
                    label: 'url',
                },
                {
                    key: 'deal',
                    label: 'deal',
                },
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
                        if (result.code == 1) {
                            if (result.data == null || result.data.length == 0) {
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
                    if (this.infos[index].path == path) {
                        break
                    }
                }
                if (index == this.infos.length) {
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
                    if (this.infos[index].path == path) {
                        break
                    }
                }
                if (index == this.infos.length) {
                    alert('非法查询文件的路径: ' + path)
                    return
                }
                getFileCompleteInfo(this.infos, index)
            },
            deleteFile(path) {
                let index = 0
                for (; index < this.infos.length; index++) {
                    if (this.infos[index].path == path) {
                        break
                    }
                }
                if (index == this.infos.length) {
                    alert('非法删除文件的路径: ' + path)
                    return
                }
                removeFile(this.infos, index)
            },
        }
    })

    const pullSyncVue = new Vue({
        el: '#pullSyncForm',
        data: {
            pullSynHost: null,
            token: null,
            loading: false
        },
        methods: {
            pull() {
                if (this.pullSynHost == null || this.pullSynHost == '' || this.token == null || this.token == '') {
                    alert('pull同步URL或者token为空')
                    return
                }
                this.loading = true
                instance.post("admin/pullSynFile", Qs.stringify({pullSynHost: this.pullSynHost, token: this.token}))
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        alert('失败数量: ' + result.data + ', 失败原因: ' + result.message)
                        if (result.code == 1) {
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
                this.pullSynHost = null
                this.token = null
            }
        }
    })

    const pushSyncVue = new Vue({
        el: '#pushSyncForm',
        data: {
            pushSynHost: null,
            token: null,
            loading: false
        },
        methods: {
            push() {
                if (this.pushSynHost == null || this.pushSynHost == '' || this.token == null || this.token == '') {
                    alert('push同步URL或者token为空')
                    return
                }
                this.loading = true
                instance.post("admin/pushSynFile", Qs.stringify({pushSynHost: this.pushSynHost, token: this.token}))
                    .then(response => {
                        this.loading = false
                        const result = response.data
                        alert('失败数量: ' + result.data + ', 失败原因: ' + result.message)
                        if (result.code == 1) {
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
                this.pushSynHost = null
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
                if (index == -1) {
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
                    if (names[i] == '') {
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
                {
                    key: 'name',
                    label: 'name',
                    sortable: true,
                },
                {
                    key: 'mime',
                    label: 'mime',
                    sortable: true,
                },
                {
                    key: 'size',
                    label: 'size',
                    sortable: true,
                },
                {
                    key: 'count',
                    label: 'count',
                    sortable: true,
                },
                {
                    key: 'md5',
                    label: 'md5',
                },
                {
                    key: 'url',
                    label: 'url',
                },
                {
                    key: 'deal',
                    label: 'deal',
                },
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
                        if (result.code == 1) {
                            if (result.data == null || result.data.length == 0) {
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
                    if (this.infos[index].path == path) {
                        break
                    }
                }
                if (index == this.infos.length) {
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
                    if (this.infos[index].path == path) {
                        break
                    }
                }
                if (index == this.infos.length) {
                    alert('非法查询文件的路径: ' + path)
                    return
                }
                getFileCompleteInfo(this.infos, index)
            },
            deleteFile(path) {
                let index = 0
                for (; index < this.infos.length; index++) {
                    if (this.infos[index].path == path) {
                        break
                    }
                }
                if (index == this.infos.length) {
                    alert('非法删除文件的路径: ' + path)
                    return
                }
                removeFile(this.infos, index)
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
                alert(result.code == 1 ? '删除成功' : '删除失败')
                if (result.code == 1) {
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
                if (result.code == 1) {
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
        info.isFile = info.mime != null && info.mime != ''
        info.url = encodeURI(info.url)
        if (info.size != null) {
            info.size = formatFileSize(info.size)
        }
        return info
    }

    function formatFileSize(size) {
        if (size < 0) return '非法大小: ' + size
        if (size == 0) return '0 B'
        var s = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
        var e = Math.floor(Math.log(size) / Math.log(1024));
        return (size / Math.pow(1024, Math.floor(e))).toFixed(2) + "" + s[e];
    }

    function createFilePath(sort, filename) {
        if (sort == null) {
            sort = ''
        }
        return sort + '/' + formatDate(new Date(), 'yyyyMMdd') + '/' + filename
    }

    function formatDate(date, fmt) {
        let o = {
            'M+': date.getMonth() + 1, //月份
            'd+': date.getDate(), //日
            'h+': date.getHours(), //小时
            'm+': date.getMinutes(), //分
            's+': date.getSeconds(), //秒
            'q+': Math.floor((date.getMonth() + 3) / 3), //季度
            'S': date.getMilliseconds() //毫秒
        }
        if (/(y+)/.test(fmt)) {
            fmt = fmt.replace(RegExp.$1, (date.getFullYear() + '').substr(4 - RegExp.$1.length))
        }
        for (let k in o) {
            if (new RegExp('(' + k + ')').test(fmt)) {
                fmt = fmt.replace(RegExp.$1, (RegExp.$1.length == 1)
                    ? (o[k]) : (('00' + o[k]).substr(('' + o[k]).length)))
            }
        }
        return fmt
    }

    function getCookieFromString(cookieString, name) {
        if (cookieString == null) {
            return null
        }
        let nameEQ = name + '='
        let ca = cookieString.split(';')
        for (let i = 0; i < ca.length; i++) {
            let c = ca[i]
            while (c.charAt(0) == ' ') c = c.substring(1, c.length)
            if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length)
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

//service===============================================================================================================

var lastFileInfos []FileSimpleInfo

func CheckToken(token string) bool {
	return Token == token
}

func AddUrl(filePath string, url string) ([]FileSimpleInfo, error) {
	request := gorequest.New()
	response, _, errs := request.Get(url).
		Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36").
		Timeout(timeout).End()

	log.WithFields(logrus.Fields{"url": url, "errs": errs}).Info("url下载请求")
	if errs != nil && len(errs) > 0 {
		return nil, errors.New(fmt.Sprintf("url下载请求失败: %v", errs))
	}
	defer response.Body.Close()

	log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Info("url下载请求")
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("url下载请求响应码异常: %v", response.StatusCode))
	}

	return AddFile(filePath, response.Body)
}

func AddFile(filePath string, reader io.Reader) ([]FileSimpleInfo, error) {
	fileSimpleInfo, err := addFileNotCompressImage(filePath, reader)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(fileSimpleInfo.Mime, "image") || strings.Contains(fileSimpleInfo.Mime, "gif") {
		log.WithFields(logrus.Fields{"filePath": fileSimpleInfo.Path}).Info("文件不是图片")
		return []FileSimpleInfo{fileSimpleInfo}, nil
	}

	log.WithFields(logrus.Fields{"filePath": fileSimpleInfo.Path}).Info("文件是图片")
	fileSimpleInfos, err := compressImage(fileSimpleInfo.Path)
	if err == nil {
		for _, fileSimpleInfo := range fileSimpleInfos {
			lastFileInfos = append(lastFileInfos, fileSimpleInfo)
		}
		if len(lastFileInfos) > LastFileInfoCount {
			lastFileInfos = lastFileInfos[len(lastFileInfos)-LastFileInfoCount:]
		}
	}

	return fileSimpleInfos, err
}

func addFileNotCompressImage(filePath string, reader io.Reader) (FileSimpleInfo, error) {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return FileSimpleInfo{}, err
	}

	err = InsertFile(bedFilePath, reader)
	if err != nil {
		return FileSimpleInfo{}, err
	}
	return getFileSimpleInfo(bedFilePath)
}

func compressImage(filePath string) ([]FileSimpleInfo, error) {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return nil, err
	}
	defer DeleteFile(bedFilePath)

	img, err := imaging.Open(bedFilePath)
	if err != nil {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath, "err": err}).Error("打开原始图片文件失败")
		return nil, err
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	var bedFilePaths []string
	times := 1
	for width > 128 && height > 128 {
		newImg := imaging.Resize(img, width, 0, imaging.Lanczos)
		newBedFilePath := fmt.Sprintf("%s.%d%s", bedFilePath, times, path.Ext(bedFilePath))
		log.WithFields(logrus.Fields{"newBedFilePath": newBedFilePath}).Info("创建缩略图片路径")
		width = width / 2
		height = height / 2
		times = times * 2

		err = imaging.Save(newImg, newBedFilePath)
		if err != nil {
			log.WithFields(logrus.Fields{"newBedFilePath": newBedFilePath, "err": err}).Error("保存缩略图片文件失败")
			DeleteFile(newBedFilePath)
			continue
		}
		bedFilePaths = append(bedFilePaths, newBedFilePath)
	}

	var fileSimpleInfos []FileSimpleInfo
	for i := range bedFilePaths {
		fileSimpleInfo, err := getFileSimpleInfo(bedFilePaths[i])
		if err == nil {
			fileSimpleInfos = append(fileSimpleInfos, fileSimpleInfo)
		}
	}
	return fileSimpleInfos, nil
}

func RemoveFile(filePath string) (FileSimpleInfo, error) {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return FileSimpleInfo{}, err
	}

	existAndIsFile, _ := ExistAndIsFile(bedFilePath)
	if !existAndIsFile {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("删除文件不存在或者不是文件")
		return FileSimpleInfo{}, errors.New("删除文件不存在或者不是文件")
	}
	fileSimpleInfo, _ := getFileSimpleInfo(bedFilePath)

	err = DeleteFile(bedFilePath)
	return fileSimpleInfo, err
}

func getFileSimpleInfo(bedFileOrFolderPath string) (FileSimpleInfo, error) {
	filePath := ClearPath(strings.Replace(bedFileOrFolderPath, FileBedPath, "", 1))
	log.WithFields(logrus.Fields{"filePath": filePath}).Info("文件路径")

	existAndIsFile, fileInfo := ExistAndIsFile(bedFileOrFolderPath)
	fileSimpleInfo := FileSimpleInfo{
		Path: filePath,
		Name: fileInfo.Name(),
	}
	if existAndIsFile {
		fileSimpleInfo.Mime = mime.TypeByExtension(path.Ext(fileInfo.Name()))
		fileSimpleInfo.Url = createUrl(filePath)
	}

	return fileSimpleInfo, nil
}

func GetFileCompleteInfo(fileOrFolderPath string) (FileCompleteInfo, error) {
	bedFileOrFolderPath, err := createBedPath(fileOrFolderPath)
	if err != nil {
		return FileCompleteInfo{}, err
	}

	existAndIsFile, fileInfo := ExistAndIsFile(bedFileOrFolderPath)
	size, count, _ := GetFileOrFolderSizeAndCount(bedFileOrFolderPath)
	fileCompleteInfo := FileCompleteInfo{
		Path:  fileOrFolderPath,
		Name:  fileInfo.Name(),
		Size:  size,
		Count: count,
	}
	if existAndIsFile {
		fileCompleteInfo.Mime = mime.TypeByExtension(path.Ext(fileInfo.Name()))
		fileCompleteInfo.Url = createUrl(fileOrFolderPath)
		md5, err := GetFileMd5(bedFileOrFolderPath)
		if err == nil {
			fileCompleteInfo.Md5 = md5
		}
	}
	return fileCompleteInfo, nil
}

func createUrl(filePath string) string {
	return ClearPath(path.Join(FileUrl, filePath))
}

func ListLastFileInfos() ([]FileSimpleInfo, error) {
	return lastFileInfos, nil
}

func ListFolderInfo(folderPath string) ([]FileSimpleInfo, error) {
	bedFolderPath, err := createBedPath(folderPath)
	if err != nil {
		return nil, err
	}

	isFile, _ := ExistAndIsFile(bedFolderPath)
	if isFile {
		fileSimpleInfo, err := getFileSimpleInfo(bedFolderPath)
		return []FileSimpleInfo{fileSimpleInfo}, err
	}

	files, err := ioutil.ReadDir(bedFolderPath)
	if err != nil {
		log.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath, "err": err}).Error("读取文件夹失败")
		return nil, err
	}

	var fileSimpleInfos []FileSimpleInfo
	for i := range files {
		childFilePath := path.Join(bedFolderPath, files[i].Name())
		info, err := getFileSimpleInfo(childFilePath)
		if err == nil {
			fileSimpleInfos = append(fileSimpleInfos, info)
		}
	}
	return fileSimpleInfos, nil
}

func ListAllFileInfo(folderPath string) ([]FileCompleteInfo, error) {
	bedFolderPath, err := createBedPath(folderPath)
	if err != nil {
		return nil, err
	}

	isFile, _ := ExistAndIsFile(bedFolderPath)
	if isFile {
		fileCompleteInfo, err := GetFileCompleteInfo(folderPath)
		return []FileCompleteInfo{fileCompleteInfo}, err
	}

	files, err := ioutil.ReadDir(bedFolderPath)
	if err != nil {
		log.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath, "err": err}).Error("读取文件夹失败")
		return nil, err
	}

	var fileCompleteInfos []FileCompleteInfo
	for i := range files {
		childFilePath := path.Join(folderPath, files[i].Name())
		infos, err := ListAllFileInfo(childFilePath)
		if err != nil {
			continue
		}
		for _, info := range infos {
			fileCompleteInfos = append(fileCompleteInfos, info)
		}
	}
	return fileCompleteInfos, nil
}

func ReceivePushSynFile(filePath string, md5 string, reader io.Reader) (FileSimpleInfo, error) {
	existAndSame, err := checkFile(filePath, md5)
	if err != nil {
		return FileSimpleInfo{}, err
	}
	if existAndSame {
		log.WithFields(logrus.Fields{"filePath": filePath}).Info("文件存在且md5相同")
		return FileSimpleInfo{}, nil
	}
	return addFileNotCompressImage(filePath, reader)
}

//检查文件是否存在且md5相同，true:存在且md5相同
func checkFile(filePath string, md5 string) (bool, error) {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return false, err
	}

	existAndIsFile, _ := ExistAndIsFile(bedFilePath)
	if !existAndIsFile {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Info("所检查文件不存在")
		return false, nil
	}

	fileMd5, err := GetFileMd5(bedFilePath)
	if err != nil {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath, "err": err}).Error("所检查的文件MD5计算失败")
		return false, errors.New("所检查的文件MD5计算失败")
	}

	return fileMd5 == md5, nil
}

func PushSynFile(pushSynHost string, token string) (int, error) {
	fileInfos, err := ListAllFileInfo("")
	if err != nil {
		return 0, err
	}

	client := &http.Client{Timeout: pullOrPushTimeout}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("http client jar创建失败")
		return 0, err
	}
	client.Jar = jar

	err = synLogin(client, pushSynHost, token)
	if err != nil {
		return 0, err
	}

	receivePushSynFileUrl := pushSynHost + ReceivePushSynFileUrl
	log.WithFields(logrus.Fields{"receivePushSynFileUrl": receivePushSynFileUrl}).Info("远端接收push文件的url")
	var failFileInfos []FileCompleteInfo
	var failFileErrors []error
	for i := range fileInfos {
		err := pushFile(client, receivePushSynFileUrl, fileInfos[i])
		if err != nil {
			failFileInfos = append(failFileInfos, fileInfos[i])
			failFileErrors = append(failFileErrors, err)
		}
	}
	for i := range failFileInfos {
		log.WithFields(logrus.Fields{"path": failFileInfos[i].Path, "err": failFileErrors[i]}).Error("push文件失败")
	}
	return len(failFileInfos), nil
}

func pushFile(client *http.Client, receivePushSynFileUrl string, fileOrFolderInfo FileCompleteInfo) error {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	go func() {
		defer pipeWriter.Close()
		defer writer.Close()

		err := writer.WriteField("filePath", fileOrFolderInfo.Path)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("写入表单参数filePath失败")
			return
		}
		err = writer.WriteField("md5", fileOrFolderInfo.Md5)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("写入表单参数md5失败")
			return
		}

		formFile, err := writer.CreateFormFile("file", fileOrFolderInfo.Name)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("创建文件表单失败")
			return
		}

		bedFilePath, err := createBedPath(fileOrFolderInfo.Path)
		if err != nil {
			return
		}
		file, err := os.Open(bedFilePath)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("打开push文件失败")
			return
		}
		defer file.Close()

		_, err = io.Copy(formFile, file)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("将文件写入表单失败")
			return
		}
	}()

	contentType := writer.FormDataContentType()
	response, err := client.Post(receivePushSynFileUrl, contentType, pipeReader)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("创建push文件请求失败")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("push文件http状态码异常")
		return errors.New(fmt.Sprintf("push文件http状态码异常: %d", response.StatusCode))
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("push文件http读取响应异常")
		return err
	}
	log.WithFields(logrus.Fields{"data": string(data)}).Info("push文件请求结果")

	var pushResult struct {
		Code    int32       `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(data, &pushResult)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("push文件http响应反序列化失败")
		return err
	}
	if pushResult.Code != SuccessCode {
		log.WithFields(logrus.Fields{"pushResult": pushResult, "fileOrFolderInfo": fileOrFolderInfo}).Error("push文件失败")
		return errors.New(fmt.Sprintf("push文件失败: %v", pushResult))
	}
	log.WithFields(logrus.Fields{"pushResult": pushResult, "fileOrFolderInfo": fileOrFolderInfo}).Info("push文件成功")
	return nil
}

func PullSynFile(pullSynHost string, token string) (int, error) {
	pullSynHost = strings.ReplaceAll(pullSynHost, "\\", "/")
	pullSynHost = strings.TrimRight(pullSynHost, "/")

	client := &http.Client{Timeout: pullOrPushTimeout}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("http client jar创建失败")
		return 0, err
	}
	client.Jar = jar

	err = synLogin(client, pullSynHost, token)
	if err != nil {
		return 0, err
	}

	allFileInfoUrl := pullSynHost + ListAllFileInfoUrl
	log.WithFields(logrus.Fields{"allFileInfoUrl": allFileInfoUrl}).Info("获取全部文件信息Url")
	response, err := client.Get(allFileInfoUrl)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("获取全部文件信息Url失败")
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("获取全部文件信息http状态码异常")
		return 0, errors.New("获取全部文件信息http状态码异常")
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("读取全部文件信息http状态码异常")
		return 0, err
	}
	log.WithFields(logrus.Fields{"data": string(data)}).Info("获取全部文件信息请求结果")
	var allFileInfoResult struct {
		Code    int                `json:"code"`
		Message string             `json:"message"`
		Data    []FileCompleteInfo `json:"data"`
	}
	err = json.Unmarshal(data, &allFileInfoResult)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("反序列化全部文件信息失败")
		return 0, err
	}
	if allFileInfoResult.Code != SuccessCode {
		log.WithFields(logrus.Fields{"allFileInfoResult": allFileInfoResult}).Error("获取全部文件信息失败")
		return 0, errors.New("获取全部文件信息失败")
	}
	log.WithFields(logrus.Fields{"allFileInfoResult": allFileInfoResult}).Info("获取全部文件信息成功")
	var failFileInfos []FileCompleteInfo
	var errs []error
	for i := range allFileInfoResult.Data {
		fileOrFolderInfo := allFileInfoResult.Data[i]
		pullFileUrl := pullSynHost + fileOrFolderInfo.Url
		log.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl}).Info("文件下载url")
		err := pullFile(pullFileUrl, fileOrFolderInfo)
		if err != nil {
			failFileInfos = append(failFileInfos, fileOrFolderInfo)
			errs = append(errs, err)
		} else {
			log.WithFields(logrus.Fields{"path": fileOrFolderInfo.Path}).Info("文件下载成功")
		}
	}
	for i := range failFileInfos {
		log.WithFields(logrus.Fields{"path": failFileInfos[i].Path, "err": errs[i]}).Error("下载文件失败")
	}
	return len(failFileInfos), nil
}

func pullFile(pullFileUrl string, fileOrFolderInfo FileCompleteInfo) error {
	filePath := fileOrFolderInfo.Path
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return err
	}

	existAndIsFolder, _ := ExistAndIsFolder(bedFilePath)
	if existAndIsFolder {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath}).Error("下载文件的路径是文件夹")
		return errors.New("下载文件的路径是文件夹")
	}

	existAndSame, err := checkFile(filePath, fileOrFolderInfo.Md5)
	if err != nil {
		return err
	}
	if existAndSame {
		log.WithFields(logrus.Fields{"filePath": filePath}).Info("文件存在且md5相同")
		return nil
	}

	response, err := http.Get(pullFileUrl)
	if err != nil {
		log.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl, "err": err}).Error("文件下载请求失败")
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"pullFileUrl": pullFileUrl, "StatusCode": response.StatusCode}).Error("文件下载请求状态码异常")
		return errors.New("文件下载请求状态码异常")
	}

	_, err = addFileNotCompressImage(filePath, response.Body)
	if err != nil {
		return err
	}

	fileMd5, err := GetFileMd5(bedFilePath)
	if err != nil {
		return err
	}
	if fileMd5 != fileOrFolderInfo.Md5 {
		log.WithFields(logrus.Fields{"bedFilePath": bedFilePath, "localMd5": fileMd5, "remoteMd5": fileOrFolderInfo.Md5}).Info("文件下载了，但MD5不匹配")
		return errors.New("文件下载了，但MD5不匹配")
	}
	return nil
}

func synLogin(client *http.Client, synUrl string, token string) error {
	loginUrl := synUrl + LoginUrl
	log.WithFields(logrus.Fields{"loginUrl": loginUrl}).Info("登录远程端Url")
	postValues := url.Values{}
	postValues.Add("token", token)
	response, err := client.PostForm(loginUrl, postValues)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("登录http请求异常")
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.WithFields(logrus.Fields{"StatusCode": response.StatusCode}).Error("登录http状态码异常")
		return errors.New(fmt.Sprintf("登录http状态码异常: %v", response.StatusCode))
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("登录http请求读取异常")
		return err
	}
	log.WithFields(logrus.Fields{"loginData": string(data)}).Info("登录请求结果")
	var loginResult struct {
		Code    int32       `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(data, &loginResult)
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Error("登录http请求反序列化异常")
		return err
	}
	if loginResult.Code != SuccessCode {
		log.WithFields(logrus.Fields{"loginResult": loginResult}).Error("登录失败")
		return errors.New(fmt.Sprintf("登录失败: %v", loginResult))
	}
	log.WithFields(logrus.Fields{"loginResult": loginResult}).Info("登录成功")
	return nil
}

func createBedPath(fileOrFolderPath string) (string, error) {
	bedFileOrFolderPath := ClearPath(path.Join(FileBedPath, fileOrFolderPath))
	log.WithFields(logrus.Fields{"bedFileOrFolderPath": bedFileOrFolderPath}).Info("创建床文件路径")

	if !strings.HasPrefix(bedFileOrFolderPath, FileBedPath) {
		log.WithFields(logrus.Fields{"bedFileOrFolderPath": bedFileOrFolderPath}).Error("床文件路径不在指定路径下")
		return "", errors.New("床文件路径不在指定路径下")
	}

	return bedFileOrFolderPath, nil
}

//dao===================================================================================================================

func InsertFile(filePath string, reader io.Reader) error {
	return WriteFileWithReaderOrCreateIfNotExist(filePath, reader)
}

func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("删除文件失败")
		return err
	}

	//将`/aaa/bbb/text.txt`变为`/aaa/bbb/`
	folderPath, _ := path.Split(filePath)
	//将`/aaa/bbb/`变为`/aaa/bbb`
	folderPath = path.Clean(folderPath)
	for {
		log.WithFields(logrus.Fields{"folderPath": folderPath}).Info("创建父文件夹检查是否为空后删除")
		files, err := ioutil.ReadDir(folderPath)
		if err != nil {
			log.WithFields(logrus.Fields{"folderPath": folderPath, "err": err}).Error("读取父文件夹失败")
			return err
		}
		if len(files) > 0 {
			log.WithFields(logrus.Fields{"folderPath": folderPath}).Info("父文件夹不为空")
			return nil
		}
		err = os.Remove(folderPath)
		if err != nil {
			log.WithFields(logrus.Fields{"folderPath": folderPath, "err": err}).Error("删除父文件夹失败")
			return err
		}
		//将`/aaa/bbb`变为`/aaa`
		//如果上面不将`/aaa/bbb/`变为`/aaa/bbb`
		//这里`/aaa/bbb/`依然会返回`/aaa/bbb/`
		folderPath = path.Dir(folderPath)
	}
}

func GetFileMd5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("计算MD5打开文件失败")
		return "", err
	}
	defer file.Close()
	md5 := md5.New()
	_, err = io.Copy(md5, file)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("计算MD5读取文件失败")
		return "", err
	}
	return hex.EncodeToString(md5.Sum(nil)), nil
}

func GetFileOrFolderSizeAndCount(fileOrFolderPath string) (int64, int32, error) {
	isFile, fileInfo := ExistAndIsFile(fileOrFolderPath)
	if isFile {
		return fileInfo.Size(), 1, nil
	}
	return getFolderSizeAndCount(fileOrFolderPath)
}

func getFolderSizeAndCount(folderPath string) (int64, int32, error) {
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		log.WithFields(logrus.Fields{"folderPath": folderPath, "err": err}).Error("读取文件夹失败")
		return 0, 0, err
	}
	size := int64(0)
	count := int32(0)
	for i := range files {
		childFilePath := path.Join(folderPath, files[i].Name())
		isFile, fileInfo := ExistAndIsFile(childFilePath)
		if isFile {
			size = size + fileInfo.Size()
			count = count + 1
			continue
		}
		childSize, childCount, err := getFolderSizeAndCount(childFilePath)
		if err != nil {
			continue
		}
		size = size + childSize
		count = count + childCount
	}
	return size, count, nil
}

//utils=================================================================================================================

func ExistPath(path string) (bool, os.FileInfo) {
	fileInfo, err := os.Stat(path)
	return err == nil || os.IsExist(err), fileInfo
}

func ExistAndIsFolder(folderPath string) (bool, os.FileInfo) {
	exist, fileInfo := ExistPath(folderPath)
	return exist && fileInfo.IsDir(), fileInfo
}

func ExistAndIsFile(filePath string) (bool, os.FileInfo) {
	exist, fileInfo := ExistPath(filePath)
	return exist && !fileInfo.IsDir(), fileInfo
}

func WriteFileWithBytesOrCreateIfNotExist(filePath string, bytes []byte) error {
	exist, _ := ExistPath(filePath)
	if exist {
		err := ioutil.WriteFile(filePath, bytes, 0644)
		if err != nil {
			log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("写入文件失败")
		}
		return err
	}
	return CreateFileWithBytes(filePath, bytes)
}

func WriteFileWithReaderOrCreateIfNotExist(filePath string, reader io.Reader) error {
	exist, _ := ExistPath(filePath)
	if exist {
		file, err := os.Open(filePath)
		if err != nil {
			log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("打开文件失败")
			return err
		}
		defer file.Close()
		written, err := io.Copy(file, reader)
		if err != nil {
			log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("写入文件数据失败")
		} else {
			log.WithFields(logrus.Fields{"filePath": filePath, "written": written}).Error("写入文件数据成功")
		}
		return err
	}
	return CreateFileWithReader(filePath, reader)
}

func ReadFileOrCreateIfNotExist(filePath string, defaultText string) (string, error) {
	exist, _ := ExistPath(filePath)
	if exist {
		bytes, err := readFile(filePath)
		if err != nil {
			return "", err
		}
		text := string(bytes)
		log.WithFields(logrus.Fields{"filePath": filePath, "text": text}).Info("读取文件文本")
		return text, err
	}
	err := CreateFileWithBytes(filePath, []byte(defaultText))
	return defaultText, err
}

func CreateFileWithBytes(filePath string, bytes []byte) error {
	file, err := createFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	written, err := file.Write(bytes)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("写入文件初始数据失败")
	} else {
		log.WithFields(logrus.Fields{"filePath": filePath, "written": written}).Error("写入文件初始数据成功")
	}
	return err
}

func CreateFileWithReader(filePath string, reader io.Reader) error {
	file, err := createFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	written, err := io.Copy(file, reader)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("写入文件初始数据失败")
	} else {
		log.WithFields(logrus.Fields{"filePath": filePath, "written": written}).Error("写入文件初始数据成功")
	}
	return err
}

func createFile(filePath string) (*os.File, error) {
	folderPath, _ := path.Split(filePath)
	log.WithFields(logrus.Fields{"folderPath": folderPath}).Info("文件父文件夹")
	if folderPath != "" {
		err := os.MkdirAll(folderPath, 0666)
		if err != nil {
			log.WithFields(logrus.Fields{"folderPath": folderPath, "err": err}).Error("创建父文件夹失败")
			return nil, err
		}
	}
	file, err := os.Create(filePath)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("创建文件失败")
	} else {
		log.WithFields(logrus.Fields{"filePath": filePath}).Info("创建文件成功")
	}
	return file, err
}

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("打开文件失败")
		return nil, err
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.WithFields(logrus.Fields{"filePath": filePath, "err": err}).Error("读取文件失败")
		return nil, err
	}
	return bytes, err
}

func ClearPath(fileOrFolderPath string) string {
	fileOrFolderPath = strings.ReplaceAll(fileOrFolderPath, "\\", "/")
	return path.Clean(fileOrFolderPath)
}
