package resource

import (
	"context"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -nullcache=&pb.FrontPage{Id:-1} -check_null_code=$==nil -struct_name=Dao
	DefaultPage(ctx context.Context, request *pb.FrontPageReq) (*pb.FrontPage, error)
	// bts: -check_null_code=$==nil -struct_name=Dao
	OnlinePage(ctx context.Context, request *pb.FrontPageReq) ([]*pb.FrontPage, error)
	// bts:  -check_null_code=$==nil -struct_name=Dao
	HiddenPage(ctx context.Context, request *pb.FrontPageReq) ([]*pb.FrontPage, error)
}
