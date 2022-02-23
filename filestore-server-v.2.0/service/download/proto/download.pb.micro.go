// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: download.proto

package download

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

import (
	context "context"
	client "github.com/asim/go-micro/v3/client"
	server "github.com/asim/go-micro/v3/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for Download service

type DownloadService interface {
	DownloadEntry(ctx context.Context, in *ReqEntry, opts ...client.CallOption) (*RespEntry, error)
}

type downloadService struct {
	c    client.Client
	name string
}

func NewDownloadService(name string, c client.Client) DownloadService {
	if c == nil {
		c = client.NewClient()
	}
	if len(name) == 0 {
		name = "download"
	}
	return &downloadService{
		c:    c,
		name: name,
	}
}

func (c *downloadService) DownloadEntry(ctx context.Context, in *ReqEntry, opts ...client.CallOption) (*RespEntry, error) {
	req := c.c.NewRequest(c.name, "Download.DownloadEntry", in)
	out := new(RespEntry)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Download service

type DownloadHandler interface {
	DownloadEntry(context.Context, *ReqEntry, *RespEntry) error
}

func RegisterDownloadHandler(s server.Server, hdlr DownloadHandler, opts ...server.HandlerOption) error {
	type download interface {
		DownloadEntry(ctx context.Context, in *ReqEntry, out *RespEntry) error
	}
	type Download struct {
		download
	}
	h := &downloadHandler{hdlr}
	return s.Handle(s.NewHandler(&Download{h}, opts...))
}

type downloadHandler struct {
	DownloadHandler
}

func (h *downloadHandler) DownloadEntry(ctx context.Context, in *ReqEntry, out *RespEntry) error {
	return h.DownloadHandler.DownloadEntry(ctx, in, out)
}
