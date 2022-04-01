package handler

import (
	"context"
	"download/config"
	download "download/proto"
)


type Download1 struct{}

func (u *Download1)DownloadEntry(ctx context.Context, req *download.ReqEntry, resp *download.RespEntry) error{
	resp.Entry = config.DownloadEntry
	return nil
}
