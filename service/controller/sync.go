package controller

import (
	"context"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/cellargalaxy/go_file_bed/service"
)

func PushSyncFile(ctx context.Context, request model.PushSyncFileRequest) (*model.PushSyncFileResponse, error) {
	err := service.PushSyncFile(ctx, request.Address, request.Secret, request.Path)
	if err != nil {
		return nil, err
	}
	var response model.PushSyncFileResponse
	return &response, nil
}

func PullSyncFile(ctx context.Context, request model.PullSyncFileRequest) (*model.PullSyncFileResponse, error) {
	err := service.PullSyncFile(ctx, request.Address, request.Secret, request.Path)
	if err != nil {
		return nil, err
	}
	var response model.PullSyncFileResponse
	return &response, nil
}
