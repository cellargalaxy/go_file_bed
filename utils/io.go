package utils

import (
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

const StaticFileBuildPath = "static/staticFileBuild.go"

func BuildStaticFile(fileOrFolderPaths ...string) error {
	for i := range fileOrFolderPaths {
		fileOrFolderPath := fileOrFolderPaths[i]

		existAndIsFile, _ := ExistAndIsFile(fileOrFolderPath)
		if existAndIsFile {
			bytes, err := ioutil.ReadFile(fileOrFolderPath)
			if err != nil {
				return err
			}
			base64String := base64.StdEncoding.EncodeToString(bytes)
			err = writeBuildStaticFile(map[string]string{fileOrFolderPath: base64String})
			if err != nil {
				return err
			}
		}

		existAndIsFolder, _ := ExistAndIsFolder(fileOrFolderPath)
		if !existAndIsFolder {
			return errors.New(fmt.Sprintf("路径不存在: %v", fileOrFolderPath))
		}

		fileInfoMap, err := GetAllFileInfoFromFolder(fileOrFolderPath)
		if err != nil {
			return err
		}
		base64Map := map[string]string{}
		for filePath := range fileInfoMap {
			bytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}
			base64String := base64.StdEncoding.EncodeToString(bytes)
			base64Map[filePath] = base64String
		}
		err = writeBuildStaticFile(base64Map)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeBuildStaticFile(base64Map map[string]string) error {
	mapStrings := ""
	for k, v := range base64Map {
		mapStrings = mapStrings + "\n	Base64Map[\"" + k + "\"] = \"" + v + "\""
	}
	code := `package static
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
		bytes, err := base64.StdEncoding.DecodeString(base64String)
		if err != nil {
			return err
		}
		writer, err := CreateFile(filePath)
		if err != nil {
			return err
		}
		defer writer.Close()
		_, err = writer.Write(bytes)
		if err != nil {
			return err
		}
	}
	return nil
}
