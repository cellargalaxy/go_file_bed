package model

type FileSimpleInfo struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	IsFile bool   `json:"isFile"`
	Url    string `json:"url"`
}

type FileCompleteInfo struct {
	FileSimpleInfo
	Size  int64  `json:"size"`
	Count int32  `json:"count"`
	Md5   string `json:"md5"`
}
