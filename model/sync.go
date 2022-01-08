package model

import (
	"context"
	"github.com/cellargalaxy/go_common/util"
)

type FileSyncInter interface {
	Pull(ctx context.Context, localPath, remotePath string) error
	Push(ctx context.Context, localPath, remotePath string) error
}

type PushSyncFileRequest struct {
	Address string `json:"address" form:"address" query:"address"`
	Secret  string `json:"secret" form:"secret" query:"secret"`
	Path    string `json:"path" form:"path" query:"path"`
}

func (this PushSyncFileRequest) String() string {
	return util.ToJsonString(this)
}

type PushSyncFileResponse struct {
}

func (this PushSyncFileResponse) String() string {
	return util.ToJsonString(this)
}

type PullSyncFileRequest struct {
	Address string `json:"address" form:"address" query:"address"`
	Secret  string `json:"secret" form:"secret" query:"secret"`
	Path    string `json:"path" form:"path" query:"path"`
}

func (this PullSyncFileRequest) String() string {
	return util.ToJsonString(this)
}

type PullSyncFileResponse struct {
}

func (this PullSyncFileResponse) String() string {
	return util.ToJsonString(this)
}
