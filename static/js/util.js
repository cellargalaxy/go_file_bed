function enSha256(text) {
    if (text === undefined || text == null) {
        text = ''
    }
    const hash = CryptoJS.SHA256(text)
    return hash
}

function enSha256Hex(text) {
    const hash = enSha256(text)
    return hash.toString(CryptoJS.enc.Hex)
}

function enMd5(text) {
    if (text === undefined || text == null) {
        text = ''
    }
    const hash = CryptoJS.MD5(text)
    return hash
}

function enMd5Hex(text) {
    const hash = enMd5(text)
    return hash.toString(CryptoJS.enc.Hex)
}

function enBase64(text) {
    const data = CryptoJS.enc.Utf8.parse(text)
    const base64 = CryptoJS.enc.Base64.stringify(data)
    return base64
}

function deBase64(base64) {
    let data = CryptoJS.enc.Base64.parse(base64)
    let text = data.toString(CryptoJS.enc.Utf8)
    return text
}

/**
 * AES CBC加密
 * @param secret
 * @param text
 * @returns {string}
 */
function enAESCBC(text, secret) {
    const data = CryptoJS.enc.Utf8.parse(text)
    const key = enSha256Hex(secret)
    const iv = enMd5(secret)
    const encrypt = CryptoJS.AES.encrypt(data, key, {
        iv: iv,
        padding: CryptoJS.pad.Pkcs7,
        mode: CryptoJS.mode.CBC,
    })
    return encrypt.toString()
}

/**
 * AES CBC解密
 * @param word
 * @returns {string}
 * @constructor
 */
function deAESCBC(text, secret) {
    const key = enSha256Hex(secret)
    const iv = enMd5(secret)
    const decrypt = CryptoJS.AES.decrypt(text, key, {
        iv: iv,
        padding: CryptoJS.pad.Pkcs7,
        mode: CryptoJS.mode.CBC,
    })
    return decrypt.toString(CryptoJS.enc.Utf8)
}

const secretKey = document.location.pathname + '@secret'

function getSecret() {
    return localStorage.getItem(secretKey)
}

function setSecret(secret) {
    localStorage.setItem(secretKey, secret)
}

const tokenExpKey = 'tokenExp'

function getTokenExp() {
    const exp = localStorage.getItem(tokenExpKey)
    if (isNum(exp)) {
        return parseInt(exp)
    }
    return 3
}

function setTokenExp(exp) {
    localStorage.setItem(tokenExpKey, exp)
}

function enJwt() {
    const timeStamp = getTimeStamp()
    const exp = getTokenExp()
    const header = {'typ': 'JWT', 'alg': 'HS256'}
    const headerJson = JSON.stringify(header)
    const payload = {'iat': timeStamp, 'exp': timeStamp + exp, 'allow_re_request': false, 'request_id': genId()}
    const payloadJson = JSON.stringify(payload)
    const secret = getSecret()
    const secretHex = enSha256Hex(secret)
    return KJUR.jws.JWS.sign("HS256", headerJson, payloadJson, {hex: secretHex})
}

function isNum(s) {
    if (s != null && s !== '') {
        return !isNaN(s)
    }
    return false
}

function getTimeStamp() {
    return Date.parse(new Date()) / 1000
}

function formatTimestamp(timestamp, fmt) {
    return formatDate(new Date(timestamp), fmt)
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

function genId() {
    return formatDate(new Date(), 'YYMMDDHHmmssS') + 'web' + randomWord()
}

function randomWord() {
    let words = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
        'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
        'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z']
    return words[Math.floor(Math.random() * words.length)]
}

function getQueryString(name) {
    let reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)", "i")
    let r = window.location.search.substr(1).match(reg)
    if (r != null) {
        return decodeURIComponent(r[2])
    }
    return null
}

function startWith(string, start) {
    return string.indexOf(start) === 0
}

function endWith(string, end) {
    return string.substring(string.length - end.length) === end
}

function containWith(string, sub) {
    return string.indexOf(sub) >= 0
}

function isNum(s) {
    if (s != null && s !== '') {
        return !isNaN(s)
    }
    return false
}

function toString(value) {
    if (value === null || typeof value === 'undefined') {
        return ''
    } else if (value instanceof Object) {
        return Object.keys(value)
            .sort()
            .map(key => toString(value[key]))
            .join(' ')
    } else {
        return String(value)
    }
}

function getAddYearDate(year) {
    let before = new Date()
    before.setFullYear(before.getFullYear() + year)
    return before
}

function parse2BeijingTimestamp(date) {
    let timestamp = 0
    if (date !== undefined && date != null && date !== '') {
        timestamp = Date.parse(date + ' GMT+8') / 1000
    }
    if (!isNum(timestamp)) {
        timestamp = 0
    }
    return timestamp
}

function formatFileSize(size) {
    if (size < 0) return '非法大小: ' + size
    if (size === 0) return '0B'
    const s = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
    const e = Math.floor(Math.log(size) / Math.log(1024))
    return (size / Math.pow(1024, Math.floor(e))).toFixed(2) + '' + s[e]
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

function log(...data) {
    console.log('log', data)
}

function sortCompare(aRow, bRow, key, sortDesc, formatter, compareOptions, compareLocale) {
    let a = aRow[key]
    let b = bRow[key]
    if (isNum(a) && isNum(b)) {
        a = parseFloat(a)
        b = parseFloat(b)
        return a < b ? -1 : a > b ? 1 : 0
    }
    return toString(a).localeCompare(toString(b), compareLocale, compareOptions)
}