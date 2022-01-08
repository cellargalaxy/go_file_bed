package model

import "context"

type FrpHandlerInter interface {
	GetAddress(ctx context.Context) string
	GetUsername(ctx context.Context) string
	GetPassword(ctx context.Context) string
}

type FrpInter interface {
	Pull(ctx context.Context, localPath, remotePath string) error
}
