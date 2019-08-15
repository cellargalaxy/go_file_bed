package model

type FileOrFolderInfo struct {
	Path      string
	Name      string
	FileCount int32
	FileSize  int64
	Md5       string
}
