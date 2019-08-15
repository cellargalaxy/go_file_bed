package model

const SuccessCode = 1
const FailCode = 2

type FileOrFolderInfo struct {
	Path      string
	Name      string
	FileCount int32
	FileSize  int64
	Md5       string
}
