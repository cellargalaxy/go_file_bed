package service

import (
	"context"
)

//检查文件是否存在且md5相同，true:存在且md5相同
func checkFile(ctx context.Context, filePath string, md5 string) (bool, error) {
	info, err := GetFileCompleteInfo(ctx, filePath)
	if err != nil {
		return false, err
	}
	return info != nil && info.Md5 == md5, nil
}
