package utils

import (
	"crypto/md5"
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

func GetFileOrFolderInfo(fileOrFolderPath string) (int32, int64, os.FileInfo, error) {
	existAndIsFile, fileInfo := ExistAndIsFile(fileOrFolderPath)
	if existAndIsFile {
		return 1, fileInfo.Size(), fileInfo, nil
	}
	existAndIsFolder, fileInfo := ExistAndIsFolder(fileOrFolderPath)
	if !existAndIsFolder {
		return 0, -1, nil, errors.New(fmt.Sprintf("路径不存在: %v", fileOrFolderPath))
	}
	files, err := ioutil.ReadDir(fileOrFolderPath)
	if err != nil {
		return 0, -1, nil, err
	}
	var folderCount int32 = 0
	var folderSize int64 = 0
	for i := range files {
		count, size, _, err := GetFileOrFolderInfo(path.Join(fileOrFolderPath, files[i].Name()))
		if err != nil {
			continue
		}
		folderCount += count
		folderSize += size
	}
	return folderCount, folderSize, fileInfo, nil
}

func ClearPath(fileOrFolderPath string) string {
	fileOrFolderPath = strings.ReplaceAll(fileOrFolderPath, "\\", "/")
	return path.Clean(fileOrFolderPath)
}

func SumFileMd5(filePath string) (string, error) {
	reader, err := os.Open(filePath)
	defer reader.Close()
	if err != nil {
		return "", err
	}
	md5 := md5.New()
	io.Copy(md5, reader)
	return hex.EncodeToString(md5.Sum(nil)), nil
}
