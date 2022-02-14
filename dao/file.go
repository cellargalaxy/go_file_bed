package dao

import (
	"context"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strings"
)

func InsertFile(ctx context.Context, filePath string, reader io.Reader) (*model.FileSimpleInfo, error) {
	bedPath, err := createBedPath(ctx, filePath)
	if err != nil {
		return nil, err
	}
	err = util.WriteFileWithReader(ctx, bedPath, reader)
	if err != nil {
		return nil, err
	}
	return SelectFileSimpleInfo(ctx, filePath)
}

func DeleteFile(ctx context.Context, filePath string) (*model.FileSimpleInfo, error) {
	bedPath, err := createBedPath(ctx, filePath)
	if err != nil {
		return nil, err
	}

	info, err := SelectFileSimpleInfo(ctx, filePath)
	if info == nil || err != nil {
		return info, err
	}

	err = util.RemoveFile(ctx, bedPath)
	if err != nil {
		return info, err
	}

	//将`/aaa/bbb/text.txt`变为`/aaa/bbb/`
	folderPath, _ := path.Split(bedPath)
	//将`/aaa/bbb/`变为`/aaa/bbb`
	folderPath = util.ClearPath(ctx, folderPath)
	for i := 0; i < 1024; i++ {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"folderPath": folderPath}).Info("删除文件，删除父文件夹")
		files, err := util.ListFile(ctx, folderPath)
		if err != nil {
			return info, err
		}
		if len(files) > 0 {
			logrus.WithContext(ctx).WithFields(logrus.Fields{"folderPath": folderPath}).Info("删除文件，父文件夹不为空")
			return info, nil
		}
		err = util.RemoveFile(ctx, folderPath)
		if err != nil {
			return info, err
		}
		//将`/aaa/bbb`变为`/aaa`
		//如果上面不将`/aaa/bbb/`变为`/aaa/bbb`
		//这里`/aaa/bbb/`依然会返回`/aaa/bbb/`
		folderPath = path.Dir(folderPath)
	}
	return info, err
}

func SelectFileSimpleInfo(ctx context.Context, fileOrFolderPath string) (*model.FileSimpleInfo, error) {
	bedPath, err := createBedPath(ctx, fileOrFolderPath)
	if err != nil {
		return nil, err
	}
	pathInfo := util.GetPathInfo(ctx, bedPath)
	if pathInfo == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Warn("查询文件简单信息，路径不存在")
		return nil, nil
	}

	var info model.FileSimpleInfo
	info.Path = fileOrFolderPath
	info.Name = pathInfo.Name()
	info.IsFile = !pathInfo.IsDir()
	return &info, nil
}

func SelectFileCompleteInfo(ctx context.Context, fileOrFolderPath string) (*model.FileCompleteInfo, error) {
	bedPath, err := createBedPath(ctx, fileOrFolderPath)
	if err != nil {
		return nil, err
	}
	pathInfo := util.GetPathInfo(ctx, bedPath)
	if pathInfo == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Warn("查询文件完整信息，路径不存在")
		return nil, nil
	}

	var info model.FileCompleteInfo
	info.Path = fileOrFolderPath
	info.Name = pathInfo.Name()
	info.IsFile = !pathInfo.IsDir()

	if !pathInfo.IsDir() {
		info.Size = pathInfo.Size()
		info.Count = 1

		info.Md5 = "out_max_hash_limit"
		if info.Size <= config.Config.MaxHashLimit {
			md5, err := util.GetFileMd5(ctx, bedPath)
			if err != nil {
				return nil, err
			}
			info.Md5 = md5
		}

		return &info, nil
	}

	size, count, err := selectFolderSizeAndCount(ctx, bedPath)
	if err != nil {
		return nil, err
	}
	info.Size = size
	info.Count = count
	return &info, nil
}

func SelectFolderSimpleInfo(ctx context.Context, folderPath string) ([]model.FileSimpleInfo, error) {
	bedPath, err := createBedPath(ctx, folderPath)
	if err != nil {
		return nil, err
	}
	pathInfo := util.GetPathInfo(ctx, bedPath)
	if pathInfo == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Warn("查询文件简单信息，路径不存在")
		return nil, nil
	}

	var infos []model.FileSimpleInfo
	if !pathInfo.IsDir() {
		info, err := SelectFileSimpleInfo(ctx, folderPath)
		if info == nil || err != nil {
			return nil, err
		}
		infos = append(infos, *info)
		return infos, nil
	}

	files, err := util.ListFile(ctx, bedPath)
	if err != nil {
		return nil, err
	}
	for _, childFile := range files {
		childFilePath := path.Join(folderPath, childFile.Name())
		info, err := SelectFileSimpleInfo(ctx, childFilePath)
		if info == nil || err != nil {
			continue
		}
		infos = append(infos, *info)
	}
	return infos, nil
}

func SelectFolderCompleteInfo(ctx context.Context, folderPath string) ([]model.FileCompleteInfo, error) {
	bedPath, err := createBedPath(ctx, folderPath)
	if err != nil {
		return nil, err
	}
	pathInfo := util.GetPathInfo(ctx, bedPath)
	if pathInfo == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Warn("查询文件完整信息，路径不存在")
		return nil, nil
	}

	var infos []model.FileCompleteInfo
	if !pathInfo.IsDir() {
		info, err := SelectFileCompleteInfo(ctx, folderPath)
		if info == nil || err != nil {
			return nil, err
		}
		infos = append(infos, *info)
		return infos, nil
	}

	files, err := util.ListFile(ctx, bedPath)
	if err != nil {
		return nil, err
	}
	for _, childFile := range files {
		childFilePath := path.Join(folderPath, childFile.Name())
		info, err := SelectFileCompleteInfo(ctx, childFilePath)
		if info == nil || err != nil {
			continue
		}
		infos = append(infos, *info)
	}
	return infos, nil
}

func selectFolderSizeAndCount(ctx context.Context, folderPath string) (int64, int32, error) {
	files, err := util.ListFile(ctx, folderPath)
	if err != nil {
		return 0, 0, err
	}
	size := int64(0)
	count := int32(0)
	for _, childFile := range files {
		childFilePath := path.Join(folderPath, childFile.Name())
		pathInfo := util.GetPathInfo(ctx, childFilePath)
		if pathInfo == nil {
			continue
		}
		if !pathInfo.IsDir() {
			size += pathInfo.Size()
			count += 1
			continue
		}
		childSize, childCount, err := selectFolderSizeAndCount(ctx, childFilePath)
		if err != nil {
			continue
		}
		size += childSize
		count += childCount
	}
	return size, count, nil
}

func GetFileData(ctx context.Context, filePath string) ([]byte, error) {
	bedPath, err := createBedPath(ctx, filePath)
	if err != nil {
		return nil, err
	}
	return util.ReadFileWithData(ctx, bedPath, nil)
}

func GetReadFile(ctx context.Context, filePath string) (*os.File, error) {
	bedPath, err := createBedPath(ctx, filePath)
	if err != nil {
		return nil, err
	}
	return util.GetReadFile(ctx, bedPath)
}

func MoveFile(ctx context.Context, formPath, toPath string) error {
	formBedPath, err := createBedPath(ctx, formPath)
	if err != nil {
		return err
	}
	toBedPath, err := createBedPath(ctx, toPath)
	if err != nil {
		return err
	}
	_, err = util.GetReadFile(ctx, toBedPath)
	if err != nil {
		return err
	}
	err = os.Rename(formBedPath, toBedPath)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("移动文件，异常")
		return fmt.Errorf("移动文件，异常: %+v", err)
	}
	return nil
}

func createBedPath(ctx context.Context, fileOrFolderPath string) (string, error) {
	bedPath := util.ClearPath(ctx, path.Join(model.FileBedPath, fileOrFolderPath))
	logrus.WithContext(ctx).WithFields(logrus.Fields{"bedPath": bedPath}).Info("创建床文件路径")

	if !strings.HasPrefix(bedPath, model.FileBedPath) {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"bedPath": bedPath}).Error("文件路径不在床路径下")
		return "", fmt.Errorf("文件路径不在床路径下")
	}

	return bedPath, nil
}
