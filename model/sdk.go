package model

import "context"

type FileBedHandlerInter interface {
	GetAddress(ctx context.Context) string
	GetSecret(ctx context.Context) string
}
