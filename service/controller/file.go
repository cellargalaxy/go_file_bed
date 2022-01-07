package controller

import (
	"context"
	"github.com/cellargalaxy/go_file_bed/model"
	"github.com/cellargalaxy/go_file_bed/service"
	"io"
)

func AddUrl(ctx context.Context, request model.UrlAddRequest) (*model.UrlAddResponse, error) {
	object, err := service.AddUrl(ctx, request.Path, request.Url, request.Raw)
	if err != nil {
		return nil, err
	}
	var response model.UrlAddResponse
	response.Info = object
	return &response, nil
}

func AddFile(ctx context.Context, filePath string, reader io.Reader, raw bool) (*model.FileAddResponse, error) {
	object, err := service.AddFile(ctx, filePath, reader, raw)
	if err != nil {
		return nil, err
	}
	var response model.FileAddResponse
	response.Info = object
	return &response, nil
}

func RemoveFile(ctx context.Context, request model.FileRemoveRequest) (*model.FileRemoveResponse, error) {
	object, err := service.RemoveFile(ctx, request.Path)
	if err != nil {
		return nil, err
	}
	var response model.FileRemoveResponse
	response.Info = object
	return &response, nil
}

func GetFileCompleteInfo(ctx context.Context, request model.FileCompleteInfoGetRequest) (*model.FileCompleteInfoGetResponse, error) {
	object, err := service.GetFileCompleteInfo(ctx, request.Path)
	if err != nil {
		return nil, err
	}
	var response model.FileCompleteInfoGetResponse
	response.Info = object
	return &response, nil
}

func ListFileSimpleInfo(ctx context.Context, request model.FileSimpleInfoListRequest) (*model.FileSimpleInfoListResponse, error) {
	object, err := service.ListFileSimpleInfo(ctx, request.Path)
	if err != nil {
		return nil, err
	}
	var response model.FileSimpleInfoListResponse
	response.Infos = object
	return &response, nil
}

func ListLastFileInfo(ctx context.Context, request model.LastFileInfoListRequest) (*model.LastFileInfoListResponse, error) {
	object, err := service.ListLastFileInfo(ctx)
	if err != nil {
		return nil, err
	}
	var response model.LastFileInfoListResponse
	response.Infos = object
	return &response, nil
}
