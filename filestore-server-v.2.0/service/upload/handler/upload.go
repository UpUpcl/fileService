package handler

import (
	"context"
	Uploadconfig "upload/config"
	"upload/proto"
)

type Upload1 struct{}

func (u *Upload1) UploadEntry(cxt context.Context, req *Upload.ReqEntry, resp *Upload.RespEntry) error{
	resp.Entry = Uploadconfig.UploadEntry
	return nil
}