package dao

import (
	"../model"
	"../utils"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
)

var log = logrus.New()

//添加文件
func InsertFile(filePath string, reader io.Reader) error {
	writer, err := utils.CreateFile(filePath)
	defer writer.Close()
	if err != nil {
		return err
	}
	written, err := io.Copy(writer, reader)
	log.WithFields(logrus.Fields{"filePath": filePath, "written": written}).Info("写入文件字节数")
	return err
}

//删除文件，随便删除文件其上的空文件夹
func DeleteFile(filePath string) error {
	existAndIsFile, _ := utils.ExistAndIsFile(filePath)
	if !existAndIsFile {
		log.WithFields(logrus.Fields{"filePath": filePath}).Info("文件不存在或者不是文件")
		return errors.New(fmt.Sprintf("文件不存在或者不是文件: %v", filePath))
	}
	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	//将`/aaa/bbb/text.txt`变为`/aaa/bbb/`
	folder, _ := path.Split(filePath)
	//将`/aaa/bbb/`变为`/aaa/bbb`
	folder = path.Clean(folder)
	for {
		log.WithFields(logrus.Fields{"folder": folder}).Debug("检查文件夹是否为空后删除")
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return nil
		}
		err = os.Remove(folder)
		if err != nil {
			return err
		}
		//将`/aaa/bbb`变为`/aaa`
		//如果上面不将`/aaa/bbb/`变为`/aaa/bbb`
		//这里`/aaa/bbb/`依然会返回`/aaa/bbb/`
		folder = path.Dir(folder)
	}
}

//查询文件夹下的情况
func SelectFileOrFolder(fileOrFolderPath string) ([]model.FileOrFolderInfo, error) {
	existAndIsFile, fileInfo := utils.ExistAndIsFile(fileOrFolderPath)
	if existAndIsFile {
		return []model.FileOrFolderInfo{{fileOrFolderPath, fileInfo.Name(), 1, fileInfo.Size()}}, nil
	}

	existAndIsFolder, _ := utils.ExistAndIsFolder(fileOrFolderPath)
	if !existAndIsFolder {
		log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Info("所查询路径不存在")
		return nil, errors.New(fmt.Sprintf("所查询路径不存在: %v", fileOrFolderPath))
	}

	files, err := ioutil.ReadDir(fileOrFolderPath)
	if err != nil {
		return nil, err
	}

	folderInfos := make([]model.FileOrFolderInfo, len(files))
	for i := range files {
		childFileOrFolderPath := path.Join(fileOrFolderPath, files[i].Name())
		count, size, fileInfo, _ := utils.GetFileOrFolderInfo(childFileOrFolderPath)
		folderInfos[i] = model.FileOrFolderInfo{childFileOrFolderPath, fileInfo.Name(), count, size}
	}
	return folderInfos, nil
}
