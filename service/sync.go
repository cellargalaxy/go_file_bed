package service

import (
	"context"
	"fmt"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/dao"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/cellargalaxy/go_file_bed/sdk"
	"github.com/sirupsen/logrus"
	"path"
)

func PushSyncFile(ctx context.Context, address, secret, path string) error {
	client, err := NewFileSyncClient(address, secret)
	if err != nil {
		return err
	}
	return client.Push(ctx, path, path)
}

func PullSyncFile(ctx context.Context, address, secret, path string) error {
	client, err := NewFileSyncClient(address, secret)
	if err != nil {
		return err
	}
	return client.Pull(ctx, path, path)
}

func NewFileSyncClient(address, secret string) (model.FileSyncInter, error) {
	client, err := sdk.NewDefaultFileBedClient(address, secret)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("创建FileBedClient为空")
	}
	return FileSyncClient{client: *client}, nil
}

type FileSyncClient struct {
	client sdk.FileBedClient
}

func (this FileSyncClient) Push(ctx context.Context, localPath, remotePath string) error {
	localPath = util.ClearPath(ctx, localPath)
	remotePath = util.ClearPath(ctx, remotePath)

	localInfos, err := ListFileSimpleInfo(ctx, localPath)
	if err != nil {
		return err
	}

	for i := range localInfos {
		local := path.Join(localPath, localInfos[i].Name)
		remote := path.Join(remotePath, localInfos[i].Name)

		if !localInfos[i].IsFile {
			this.Push(ctx, local, remote)
			continue
		}

		localInfo, err := GetFileCompleteInfo(ctx, local)
		if localInfo == nil || err != nil {
			continue
		}

		var request model.FileCompleteInfoGetRequest
		request.Path = remote
		logrus.WithContext(ctx).WithFields(logrus.Fields{"FileCompleteInfoGetRequest": request}).Info("Push文件，创建请求体")
		remoteInfo, err := this.client.GetFileCompleteInfo(ctx, request)
		if err != nil {
			continue
		}
		if remoteInfo != nil && localInfo.Md5 == remoteInfo.Md5 {
			continue
		}

		file, err := dao.GetReadFile(ctx, local)
		if file == nil || err != nil {
			continue
		}
		this.client.AddFile(ctx, remote, file, true)
	}
	return nil
}

func (this FileSyncClient) Pull(ctx context.Context, localPath, remotePath string) error {
	localPath = util.ClearPath(ctx, localPath)
	remotePath = util.ClearPath(ctx, remotePath)

	var request model.FileSimpleInfoListRequest
	request.Path = remotePath
	logrus.WithContext(ctx).WithFields(logrus.Fields{"FileSimpleInfoListRequest": request}).Info("Pull文件，创建请求体")
	remoteInfos, err := this.client.ListFileSimpleInfo(ctx, request)
	if err != nil {
		return err
	}

	for i := range remoteInfos {
		local := path.Join(localPath, remoteInfos[i].Name)
		remote := path.Join(remotePath, remoteInfos[i].Name)

		if !remoteInfos[i].IsFile {
			this.Pull(ctx, local, remote)
			continue
		}

		var request model.FileCompleteInfoGetRequest
		request.Path = remote
		logrus.WithContext(ctx).WithFields(logrus.Fields{"FileCompleteInfoGetRequest": request}).Info("Pull文件，创建请求体")
		remoteInfo, err := this.client.GetFileCompleteInfo(ctx, request)
		if remoteInfo == nil || err != nil {
			continue
		}
		localInfo, err := GetFileCompleteInfo(ctx, local)
		if err != nil {
			continue
		}
		if localInfo != nil && localInfo.Md5 == remoteInfo.Md5 {
			continue
		}

		url, err := this.client.GetFileDownloadUrl(ctx, remote)
		if err != nil {
			continue
		}
		AddUrl(ctx, local, url, true)
	}
	return nil
}
