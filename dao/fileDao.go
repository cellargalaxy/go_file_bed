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
	if err != nil {
		return err
	}
	defer writer.Close()
	written, err := io.Copy(writer, reader)
	log.WithFields(logrus.Fields{"filePath": filePath, "written": written}).Info("写入文件字节数")
	return err
}

//删除文件，随便删除文件其上的空文件夹
func DeleteFile(filePath string) error {
	existAndIsFile, _ := utils.ExistAndIsFile(filePath)
	if !existAndIsFile {
		log.WithFields(logrus.Fields{"filePath": filePath}).Error("文件不存在或者不是文件")
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
		md5, err := utils.SumFileMd5(fileOrFolderPath)
		if err != nil {
			return nil, err
		}
		return []model.FileOrFolderInfo{{fileOrFolderPath, fileInfo.Name(), 1, fileInfo.Size(), md5}}, nil
	}

	existAndIsFolder, _ := utils.ExistAndIsFolder(fileOrFolderPath)
	if !existAndIsFolder {
		log.WithFields(logrus.Fields{"fileOrFolderPath": fileOrFolderPath}).Error("所查询路径不存在")
		return nil, errors.New(fmt.Sprintf("所查询路径不存在: %v", fileOrFolderPath))
	}

	files, err := ioutil.ReadDir(fileOrFolderPath)
	if err != nil {
		return nil, err
	}

	var folderInfos []model.FileOrFolderInfo
	for i := range files {
		childFileOrFolderPath := path.Join(fileOrFolderPath, files[i].Name())
		count, size, fileInfo, err := utils.GetFileOrFolderInfo(childFileOrFolderPath)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("查询文件或者文件夹下的信息失败")
			continue
		}
		folderInfos = append(folderInfos, model.FileOrFolderInfo{childFileOrFolderPath, fileInfo.Name(), count, size, ""})
	}
	return folderInfos, nil
}

//查询某个文件夹下的全部文件的信息
func SelectAllFile(folderPath string) ([]model.FileOrFolderInfo, error) {
	existAndIsFile, fileInfo := utils.ExistAndIsFile(folderPath)
	if existAndIsFile {
		md5, err := utils.SumFileMd5(folderPath)
		if err != nil {
			return nil, err
		}
		return []model.FileOrFolderInfo{{folderPath, fileInfo.Name(), 1, fileInfo.Size(), md5}}, nil
	}

	existAndIsFolder, _ := utils.ExistAndIsFolder(folderPath)
	if !existAndIsFolder {
		log.WithFields(logrus.Fields{"folderPath": folderPath}).Error("所查询路径不存在")
		return nil, errors.New(fmt.Sprintf("所查询路径不存在: %v", folderPath))
	}

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var fileInfos []model.FileOrFolderInfo
	for i := range files {
		childFolderPath := path.Join(folderPath, files[i].Name())
		childFolderInfos, err := SelectAllFile(childFolderPath)
		if err != nil {
			log.WithFields(logrus.Fields{"err": err}).Error("查询文件的信息失败")
			continue
		}
		fileInfos = append(fileInfos, childFolderInfos...)
	}
	return fileInfos, nil
}
