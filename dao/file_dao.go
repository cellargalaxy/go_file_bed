package dao

import (
	"fmt"
	"github.com/cellargalaxy/go-file-bed/config"
	"github.com/cellargalaxy/go-file-bed/model"
	"github.com/cellargalaxy/go-file-bed/utils"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func InsertFile(filePath string, reader io.Reader) (model.FileSimpleInfo, error) {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return model.FileSimpleInfo{}, err
	}

	err = utils.WriteFileWithReaderOrCreateIfNotExist(bedFilePath, reader)
	if err != nil {
		return model.FileSimpleInfo{}, err
	}

	return GetFileSimpleInfo(filePath)
}

func DeleteFile(filePath string) (model.FileSimpleInfo, error) {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return model.FileSimpleInfo{}, err
	}

	info, err := GetFileSimpleInfo(filePath)
	if err != nil {
		return info, err
	}

	err = os.Remove(bedFilePath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"bedFilePath": bedFilePath, "err": err}).Error("删除文件失败")
		return info, err
	}

	//将`/aaa/bbb/text.txt`变为`/aaa/bbb/`
	folderPath, _ := path.Split(bedFilePath)
	//将`/aaa/bbb/`变为`/aaa/bbb`
	folderPath = path.Clean(folderPath)
	for {
		logrus.WithFields(logrus.Fields{"folderPath": folderPath}).Info("创建父文件夹检查是否为空后删除")
		files, err := ioutil.ReadDir(folderPath)
		if err != nil {
			logrus.WithFields(logrus.Fields{"folderPath": folderPath, "err": err}).Error("读取父文件夹失败")
			return info, err
		}
		if len(files) > 0 {
			logrus.WithFields(logrus.Fields{"folderPath": folderPath}).Info("父文件夹不为空")
			return info, nil
		}
		err = os.Remove(folderPath)
		if err != nil {
			logrus.WithFields(logrus.Fields{"folderPath": folderPath, "err": err}).Error("删除父文件夹失败")
			return info, err
		}
		//将`/aaa/bbb`变为`/aaa`
		//如果上面不将`/aaa/bbb/`变为`/aaa/bbb`
		//这里`/aaa/bbb/`依然会返回`/aaa/bbb/`
		folderPath = path.Dir(folderPath)
	}
}

func CopyFile(filePath string, writer io.Writer) error {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return err
	}

	isFile, _ := utils.ExistAndIsFile(bedFilePath)
	if !isFile {
		return fmt.Errorf("文件不存在或者不是文件")
	}

	file, err := os.Open(bedFilePath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("打开文件失败")
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("拷贝文件数据失败")
		return err
	}
	return nil
}

func ExistAndIsFile(filePath string) (bool, error) {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return false, err
	}
	exist, _ := utils.ExistAndIsFile(bedFilePath)
	return exist, nil
}

func ExistAndIsFolder(filePath string) (bool, error) {
	bedFilePath, err := createBedPath(filePath)
	if err != nil {
		return false, err
	}
	exist, _ := utils.ExistAndIsFolder(bedFilePath)
	return exist, nil
}

func GetFileSimpleInfo(fileOrFolderPath string) (model.FileSimpleInfo, error) {
	bedFileOrFolderPath, err := createBedPath(fileOrFolderPath)
	if err != nil {
		return model.FileSimpleInfo{}, err
	}

	isFile, fileInfo := utils.ExistAndIsFile(bedFileOrFolderPath)
	return model.FileSimpleInfo{
		Path:   fileOrFolderPath,
		Name:   fileInfo.Name(),
		IsFile: isFile,
	}, nil
}

func GetFileCompleteInfo(fileOrFolderPath string) (model.FileCompleteInfo, error) {
	bedFileOrFolderPath, err := createBedPath(fileOrFolderPath)
	if err != nil {
		return model.FileCompleteInfo{}, err
	}

	fileSimpleInfo, err := GetFileSimpleInfo(fileOrFolderPath)
	if err != nil {
		return model.FileCompleteInfo{}, err
	}

	isFile, fileInfo := utils.ExistAndIsFile(bedFileOrFolderPath)
	if isFile {
		md5, err := utils.GetFileMd5(bedFileOrFolderPath)
		if err != nil {
			return model.FileCompleteInfo{}, err
		}
		return model.FileCompleteInfo{
			FileSimpleInfo: fileSimpleInfo,
			Size:           fileInfo.Size(),
			Count:          1,
			Md5:            md5,
		}, nil
	}

	size, count, err := getFolderSizeAndCount(bedFileOrFolderPath)
	if err != nil {
		return model.FileCompleteInfo{}, err
	}
	return model.FileCompleteInfo{
		FileSimpleInfo: fileSimpleInfo,
		Size:           size,
		Count:          count,
	}, nil
}

func ListFileSimpleInfo(folderPath string) ([]model.FileSimpleInfo, error) {
	bedFolderPath, err := createBedPath(folderPath)
	if err != nil {
		return nil, err
	}

	isFile, _ := utils.ExistAndIsFile(bedFolderPath)
	if isFile {
		info, err := GetFileSimpleInfo(folderPath)
		if err != nil {
			return nil, err
		}
		return []model.FileSimpleInfo{info}, nil
	}

	files, err := ioutil.ReadDir(bedFolderPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath, "err": err}).Error("读取文件夹失败")
		return nil, err
	}
	var infos []model.FileSimpleInfo
	for _, childFile := range files {
		childFilePath := path.Join(folderPath, childFile.Name())
		info, err := GetFileSimpleInfo(childFilePath)
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func ListFileCompleteInfo(folderPath string) ([]model.FileCompleteInfo, error) {
	bedFolderPath, err := createBedPath(folderPath)
	if err != nil {
		return nil, err
	}

	isFile, _ := utils.ExistAndIsFile(bedFolderPath)
	if isFile {
		info, err := GetFileCompleteInfo(folderPath)
		if err != nil {
			return nil, err
		}
		return []model.FileCompleteInfo{info}, nil
	}

	files, err := ioutil.ReadDir(bedFolderPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath, "err": err}).Error("读取文件夹失败")
		return nil, err
	}
	var infos []model.FileCompleteInfo
	for _, childFile := range files {
		childFilePath := path.Join(folderPath, childFile.Name())
		info, err := GetFileCompleteInfo(childFilePath)
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func ListAllFileSimpleInfo(folderPath string) ([]model.FileSimpleInfo, error) {
	bedFolderPath, err := createBedPath(folderPath)
	if err != nil {
		return nil, err
	}

	isFile, _ := utils.ExistAndIsFile(bedFolderPath)
	if isFile {
		info, err := GetFileSimpleInfo(folderPath)
		if err != nil {
			return nil, err
		}
		return []model.FileSimpleInfo{info}, nil
	}

	files, err := ioutil.ReadDir(bedFolderPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath, "err": err}).Error("读取文件夹失败")
		return nil, err
	}
	var infos []model.FileSimpleInfo
	for _, childFile := range files {
		childFilePath := path.Join(folderPath, childFile.Name())
		childInfos, err := ListAllFileSimpleInfo(childFilePath)
		if err != nil {
			continue
		}
		for _, info := range childInfos {
			infos = append(infos, info)
		}
	}
	return infos, nil
}

func ListAllFileCompleteInfo(folderPath string) ([]model.FileCompleteInfo, error) {
	bedFolderPath, err := createBedPath(folderPath)
	if err != nil {
		return nil, err
	}

	isFile, _ := utils.ExistAndIsFile(bedFolderPath)
	if isFile {
		info, err := GetFileCompleteInfo(folderPath)
		if err != nil {
			return nil, err
		}
		return []model.FileCompleteInfo{info}, nil
	}

	files, err := ioutil.ReadDir(bedFolderPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"bedFolderPath": bedFolderPath, "err": err}).Error("读取文件夹失败")
		return nil, err
	}
	var infos []model.FileCompleteInfo
	for _, childFile := range files {
		childFilePath := path.Join(folderPath, childFile.Name())
		childInfos, err := ListAllFileCompleteInfo(childFilePath)
		if err != nil {
			continue
		}
		for _, info := range childInfos {
			infos = append(infos, info)
		}
	}
	return infos, nil
}

func getFolderSizeAndCount(folderPath string) (int64, int32, error) {
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"folderPath": folderPath, "err": err}).Error("读取文件夹失败")
		return 0, 0, err
	}
	size := int64(0)
	count := int32(0)
	for _, childFile := range files {
		childFilePath := path.Join(folderPath, childFile.Name())
		isFile, fileInfo := utils.ExistAndIsFile(childFilePath)
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

func createBedPath(fileOrFolderPath string) (string, error) {
	bedFileOrFolderPath := utils.ClearPath(path.Join(config.FileBedPath, fileOrFolderPath))
	logrus.WithFields(logrus.Fields{"bedFileOrFolderPath": bedFileOrFolderPath}).Info("创建床文件路径")

	if !strings.HasPrefix(bedFileOrFolderPath, config.FileBedPath) {
		logrus.WithFields(logrus.Fields{"bedFileOrFolderPath": bedFileOrFolderPath}).Error("文件路径不在床路径下")
		return "", fmt.Errorf("文件路径不在床路径下")
	}

	return bedFileOrFolderPath, nil
}
