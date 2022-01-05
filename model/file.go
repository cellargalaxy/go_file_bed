package model

import "github.com/cellargalaxy/go_common/util"

type FileSimpleInfo struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	IsFile bool   `json:"is_file"`
	Url    string `json:"url"`
}

type FileCompleteInfo struct {
	FileSimpleInfo
	Size  int64  `json:"size"`
	Count int32  `json:"count"`
	Md5   string `json:"md5"`
}

type UrlAddRequest struct {
	Path string `json:"path" form:"path" query:"path"`
	Url  string `json:"url" form:"url" query:"url"`
}

func (this UrlAddRequest) String() string {
	return util.ToJsonString(this)
}

type UrlAddResponse struct {
	Info *FileSimpleInfo `json:"info"`
}

func (this UrlAddResponse) String() string {
	return util.ToJsonString(this)
}

type FileAddResponse struct {
	Info *FileSimpleInfo `json:"info"`
}

func (this FileAddResponse) String() string {
	return util.ToJsonString(this)
}

type FileRemoveRequest struct {
	Path string `json:"path" form:"path" query:"path"`
}

func (this FileRemoveRequest) String() string {
	return util.ToJsonString(this)
}

type FileRemoveResponse struct {
	Info *FileSimpleInfo `json:"info"`
}

func (this FileRemoveResponse) String() string {
	return util.ToJsonString(this)
}

type FileCompleteInfoGetRequest struct {
	Path string `json:"path" form:"path" query:"path"`
}

func (this FileCompleteInfoGetRequest) String() string {
	return util.ToJsonString(this)
}

type FileCompleteInfoGetResponse struct {
	Info *FileCompleteInfo `json:"info"`
}

func (this FileCompleteInfoGetResponse) String() string {
	return util.ToJsonString(this)
}

type FileSimpleInfoListRequest struct {
	Path string `json:"path" form:"path" query:"path"`
}

func (this FileSimpleInfoListRequest) String() string {
	return util.ToJsonString(this)
}

type FileSimpleInfoListResponse struct {
	Infos []FileSimpleInfo `json:"infos"`
}

func (this FileSimpleInfoListResponse) String() string {
	return util.ToJsonString(this)
}

type LastFileInfoListRequest struct {
}

func (this LastFileInfoListRequest) String() string {
	return util.ToJsonString(this)
}

type LastFileInfoListResponse struct {
	Infos []FileSimpleInfo `json:"infos"`
}

func (this LastFileInfoListResponse) String() string {
	return util.ToJsonString(this)
}
