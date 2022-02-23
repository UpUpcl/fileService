package db

import (
	"filestore-server/db/mysql"
	"fmt"
	"log"
	"time"
)

type UserFile struct {
	UserName string
	FileHash string
	FileName string
	FileSize int64
	UploadAt string
	LastUpdated string
}

// OnUserFileUploadFinished 更新用户文件表
func OnUserFileUploadFinished(username, filehash, filename string,  filesize int64) bool {
	stmt, err := mysql.DBConn().Prepare("insert ignore into tbl_user_file " +
		"(`user_name`, `file_sha1`, `file_name`, `file_size`, `upload_at`) values(?,?,?,?,?)")
	if err != nil {
		return false
	}

	defer stmt.Close()
	_, err = stmt.Exec(username, filehash, filename, filesize, time.Now().Unix())
	if err != nil {
		return false
	}
	return true
}

// QueryUserFileMetas 批量获取用户存储文件
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mysql.DBConn().Prepare("select file_sha1, file_name, file_size, upload_at, last_update " +
		"from tbl_user_file where user_name = ? limit ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(username, limit)
	if err != nil {
		return nil, err
	}
	var userFile []UserFile
	for rows.Next(){
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize, &ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFile = append(userFile, ufile)
	}
	return userFile, nil
}

func UpdateUserFileName(username, newfilename, filesha1, filename string, filesize int64) bool {
	log.Println(username, newfilename, filesha1, filename, filesize)
	stmt, err := mysql.DBConn().Prepare("update tbl_user_file set file_name=? " +
		"where user_name=? and file_sha1=? and file_name=? and file_size=?")

	if err != nil {
		log.Println(err.Error())
		return false
	}
	defer stmt.Close()
	_, err = stmt.Exec(newfilename, username, filesha1, filename, filesize)

	if err != nil {
		return false
	}
	return true
}