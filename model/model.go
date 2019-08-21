package model

const SuccessCode = 1
const FailCode = 2

type FileOrFolderInfo struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	IsFile    bool   `json:"isFile"`
	FileCount int32  `json:"fileCount"`
	FileSize  int64  `json:"fileSize"`
	Md5       string `json:"md5"`
	Url       string `json:"url"`
}
