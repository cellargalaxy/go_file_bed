package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

//文件或者文件夹是否存在
func ExistPath(path string) (bool, os.FileInfo) {
	fileInfo, err := os.Stat(path)
	return err == nil || os.IsExist(err), fileInfo
}

//是否是文件夹
func ExistAndIsFolder(folderPath string) (bool, os.FileInfo) {
	exist, fileInfo := ExistPath(folderPath)
	return exist && fileInfo.IsDir(), fileInfo
}

//是否是文件
func ExistAndIsFile(filePath string) (bool, os.FileInfo) {
	exist, fileInfo := ExistPath(filePath)
	return exist && !fileInfo.IsDir(), fileInfo
}

//创建并打开文件
func CreateFile(filePath string) (*os.File, error) {
	folder, _ := path.Split(filePath)
	if folder != "" {
		err := os.MkdirAll(folder, 0666)
		if err != nil {
			return nil, err
		}
	}
	return os.Create(filePath)
}

func ClearPath(fileOrFolderPath string) string {
	fileOrFolderPath = strings.ReplaceAll(fileOrFolderPath, "\\", "/")
	return path.Clean(fileOrFolderPath)
}

func SumFileMd5(filePath string) (string, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	md5 := md5.New()
	io.Copy(md5, reader)
	return hex.EncodeToString(md5.Sum(nil)), nil
}

func GetAllFileInfoFromFolder(folderPath string) (map[string]os.FileInfo, error) {
	existAndIsFile, fileInfo := ExistAndIsFile(folderPath)
	if existAndIsFile {
		return map[string]os.FileInfo{folderPath: fileInfo}, nil
	}
	existAndIsFolder, _ := ExistAndIsFolder(folderPath)
	if !existAndIsFolder {
		return nil, errors.New(fmt.Sprintf("路径不存在: %v", folderPath))
	}
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}
	fileInfoMap := map[string]os.FileInfo{}
	for i := range files {
		childFolderPath := path.Join(folderPath, files[i].Name())
		childFileInfoMap, err := GetAllFileInfoFromFolder(childFolderPath)
		if err != nil {
			return nil, err
		}
		for k, v := range childFileInfoMap {
			fileInfoMap[k] = v
		}
	}
	return fileInfoMap, nil
}

const StaticFileBuildPackageName = "staticBuild"
const StaticFileBuildPath = StaticFileBuildPackageName + "/staticFileBuild.go"

func BuildStaticFile(fileOrFolderPaths ...string) error {
	for i := range fileOrFolderPaths {
		err := BuildOneStaticFile(fileOrFolderPaths[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func BuildOneStaticFile(fileOrFolderPath string) error {
	fileInfoMap, err := GetAllFileInfoFromFolder(fileOrFolderPath)
	if err != nil {
		return err
	}
	base64Map := map[string]string{}
	for filePath := range fileInfoMap {
		byteArray, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		base64String, err := ByteArrayToGzipToBase64(byteArray)
		if err != nil {
			return err
		}
		base64Map[filePath] = base64String
	}
	return writeBuildStaticFile(base64Map)
}

func ByteArrayToGzipToBase64(byteArray []byte) (string, error) {
	var byteBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&byteBuffer)
	if _, err := gzipWriter.Write(byteArray); err != nil {
		return "", err
	}
	if err := gzipWriter.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(byteBuffer.Bytes()), nil
}

func Base64ToGzipByteArray(base64String string) ([]byte, error) {
	byteArray, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(byteArray)
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()
	return ioutil.ReadAll(gzipReader)
}

func writeBuildStaticFile(base64Map map[string]string) error {
	mapStrings := ""
	for k, v := range base64Map {
		mapStrings = mapStrings + "\n	Base64Map[\"" + k + "\"] = \"" + v + "\""
	}
	code := `package ` + StaticFileBuildPackageName + `
var Base64Map = map[string]string{}
func init() {` + mapStrings + `
}`
	writer, err := CreateFile(StaticFileBuildPath)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = writer.Write([]byte(code))
	return err
}

func OutputStaticFile(base64Map map[string]string) error {
	for filePath, base64String := range base64Map {
		existAndIsFile, _ := ExistAndIsFile(filePath)
		if existAndIsFile {
			continue
		}
		byteArray, err := Base64ToGzipByteArray(base64String)
		if err != nil {
			return err
		}

		writer, err := CreateFile(filePath)
		if err != nil {
			return err
		}
		_, err = writer.Write(byteArray)
		writer.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

var Url2PathMaps = []map[string]string{
	{"key": "-", "value": "--"},
	{"key": ":", "value": "-c"},
	{"key": "/", "value": "-s"},
	{"key": "\\", "value": "-b"},
	{"key": "?", "value": "-q"},
	{"key": "*", "value": "-a"},
	{"key": ">", "value": "-g"},
	{"key": "<", "value": "-l"},
	{"key": "|", "value": "-v"},
}
var Path2UrlMaps = []map[string]string{
	{"key": "-c", "value": ":"},
	{"key": "-s", "value": "/"},
	{"key": "-b", "value": "\\"},
	{"key": "-q", "value": "?"},
	{"key": "-a", "value": "*"},
	{"key": "-g", "value": ">"},
	{"key": "-l", "value": "<"},
	{"key": "-v", "value": "|"},
	{"key": "--", "value": "-"},
}

func Url2Path(url string) string {
	for i := range Url2PathMaps {
		url = strings.ReplaceAll(url, Url2PathMaps[i]["key"], Url2PathMaps[i]["value"])
	}
	return url
}

func Path2Url(pathString string) string {
	for i := range Path2UrlMaps {
		pathString = strings.ReplaceAll(pathString, Path2UrlMaps[i]["key"], Path2UrlMaps[i]["value"])
	}
	return pathString
}
