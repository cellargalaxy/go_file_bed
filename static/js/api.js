const instance = axios.create({timeout: 60 * 1000})
instance.interceptors.request.use(
    config => {
        config.headers['Authorization'] = 'Bearer ' + enJwt()
        return config
    },
    error => Promise.reject(error))

async function ping() {
    let url = '../../ping'
    if (document.domain === 'localhost') {
        url += '.json'
    }
    try {
        let response = await instance.post(url, {
            params: {},
            paramsSerializer: params => {
                return Qs.stringify(params, {indices: false})
            }
        })
        return dealResponse(response)
    } catch (error) {
        dealErr(error)
    }
    return null
}

async function addUrl(path, link) {
    if (path === undefined || path == null || path === '') {
        dealErr('path为空')
        return null
    }
    if (link === undefined || link == null || link === '') {
        dealErr('link为空')
        return null
    }

    if (!window.confirm("确定创建？")) {
        return
    }

    let url = '../../api/addUrl'
    if (document.domain === 'localhost') {
        url += '.json'
    }
    try {
        let response = await instance.post(url, {
            path: path,
            url: link,
        })
        return dealResponse(response)
    } catch (error) {
        dealErr(error)
    }
    return null
}

async function addFile(path, file) {
    if (path === undefined || path == null || path === '') {
        dealErr('path为空')
        return null
    }
    if (file === undefined || file == null || file === '') {
        dealErr('file为空')
        return null
    }

    if (!window.confirm("确定创建？")) {
        return
    }

    const param = new FormData()
    param.append("path", path)
    param.append("file", file)

    let url = '../../api/addFile'
    if (document.domain === 'localhost') {
        url += '.json'
    }
    try {
        let response = await instance.post(url, param, {headers: {'Content-Type': 'multipart/form-data'}})
        return dealResponse(response)
    } catch (error) {
        dealErr(error)
    }
    return null
}

async function removeFile(path) {
    if (path === undefined || path == null || path === '') {
        dealErr('path为空')
        return null
    }

    if (!window.confirm("确定删除？")) {
        return
    }

    let url = '../../api/removeFile'
    if (document.domain === 'localhost') {
        url += '.json'
    }
    try {
        let response = await instance.post(url, {
            path: path,
        })
        return dealResponse(response)
    } catch (error) {
        dealErr(error)
    }
    return null
}

async function listServerConf(path) {
    if (path === undefined || path == null || path === '') {
        dealErr('path为空')
        return null
    }

    let url = '../../api/getFileCompleteInfo'
    if (document.domain === 'localhost') {
        url += '.json'
    }
    try {
        let response = await instance.get(url, {
            params: {path: path},
            paramsSerializer: params => {
                return Qs.stringify(params, {indices: false})
            }
        })
        return dealResponse(response)
    } catch (error) {
        dealErr(error)
    }
    return null
}

async function listFileSimpleInfo(path) {
    if (path === undefined || path == null || path === '') {
        dealErr('path为空')
        return null
    }

    let url = '../../api/listFileSimpleInfo'
    if (document.domain === 'localhost') {
        url += '.json'
    }
    try {
        let response = await instance.get(url, {
            params: {path: path},
            paramsSerializer: params => {
                return Qs.stringify(params, {indices: false})
            }
        })
        return dealResponse(response)
    } catch (error) {
        dealErr(error)
    }
    return null
}

async function listFileCompleteInfo(path) {
    if (path === undefined || path == null || path === '') {
        dealErr('path为空')
        return null
    }

    let url = '../../api/listFileCompleteInfo'
    if (document.domain === 'localhost') {
        url += '.json'
    }
    try {
        let response = await instance.get(url, {
            params: {path: path},
            paramsSerializer: params => {
                return Qs.stringify(params, {indices: false})
            }
        })
        return dealResponse(response)
    } catch (error) {
        dealErr(error)
    }
    return null
}

async function listLastFileInfo(path) {
    if (path === undefined || path == null || path === '') {
        dealErr('path为空')
        return null
    }

    let url = '../../api/listLastFileInfo'
    if (document.domain === 'localhost') {
        url += '.json'
    }
    try {
        let response = await instance.get(url, {
            params: {path: path},
            paramsSerializer: params => {
                return Qs.stringify(params, {indices: false})
            }
        })
        return dealResponse(response)
    } catch (error) {
        dealErr(error)
    }
    return null
}

function dealResponse(response) {
    let result = response.data
    if (result.code !== 1) {
        dealErr(result.msg)
        return null
    }
    return result.data
}

function dealErr(error) {
    let msg = JSON.stringify(error)
    if (msg === undefined || msg == null || msg === '' || msg === '{}' || msg === '[]') {
        msg = error
    }
    alert("error: " + msg)
    log(msg)
}