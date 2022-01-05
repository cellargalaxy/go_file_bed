package model

import "context"

type FileSyncInter interface {
	Pull(ctx context.Context, localPath, remotePath string) error
	Push(ctx context.Context, localPath, remotePath string) error
}
