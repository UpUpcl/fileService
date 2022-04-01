package meta

import (
	"filestore-server/db"
	"fmt"
	"log"
	"sort"
)

// FileMeta : 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UploadFileMeta 新增/更新文件元信息
func UploadFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// UpdateFileMetaDB 新增/更新元数据到mysql中
func UpdateFileMetaDB(fmeat FileMeta) bool {
	return db.OnFileUploadFinished(fmeat.FileSha1, fmeat.FileName, fmeat.FileSize, fmeat.Location)
}

func UpdateUserFileNameDB(username, newfilename string, fmeta FileMeta) bool {
	return db.UpdateUserFileName(username, newfilename, fmeta.FileSha1, fmeta.FileName, fmeta.FileSize)
}

// GetFileMeta 根据filehash返回一个FileMeta结构体
func GetFileMeta(fileShal string) FileMeta {
	return fileMetas[fileShal]
}

func GetLastFileMetas(count int) []FileMeta {
	fMetaArry := make([]FileMeta, len(fileMetas))

	for _, v := range fileMetas{
		fMetaArry = append(fMetaArry, v)
	}

	sort.Sort(ByUploadTime(fMetaArry))
	return fMetaArry[0:count]
}

// GetFileMetaDB 从mysql获取文件元信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	tfile, err := db.GetFileMeta(fileSha1)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	log.Println("fmeta 数据：")
	log.Println(fmeta)
	return &fmeta, nil
}

// RemoveFileMeta 删除元信息
func RemoveFileMeta(fileSha1 string)  {
	delete(fileMetas, fileSha1)
}
