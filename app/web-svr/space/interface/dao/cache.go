package dao

import (
	"context"

	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -nullcache=&model.Notice{Notice:"ff2364a0be3d20e46cc69efb36afe9a5"} -check_null_code=$.Notice=="ff2364a0be3d20e46cc69efb36afe9a5" -struct_name=Dao
	Notice(c context.Context, mid int64) (*model.Notice, error)
	// bts: -nullcache=&model.AidReason{Aid:-1} -check_null_code=$!=nil&&$.Aid==-1 -struct_name=Dao
	TopArc(c context.Context, mid int64) (*model.AidReason, error)
	// bts: -nullcache=&model.AidReasons{List:[]*model.AidReason{{Aid:-1}}} -check_null_code=len($.List)==1&&$.List[0].Aid==-1 -struct_name=Dao
	Masterpiece(c context.Context, mid int64) (*model.AidReasons, error)
	// bts: -nullcache=&model.ThemeDetails{List:[]*model.ThemeDetail{{ID:-1}}} -check_null_code=len($.List)==1&&$.List[0].ID==-1 -struct_name=Dao
	Theme(c context.Context, mid int64) (*model.ThemeDetails, error)
	// bts: -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	TopDynamic(c context.Context, mid int64) (int64, error)
	// bts: -nullcache=&pb.OfficialReply{Uid:req.Mid,Id:-1} -check_null_code=$==nil||$.Id==-1 -struct_name=Dao -singleflight=true
	Official(c context.Context, req *pb.OfficialRequest) (*pb.OfficialReply, error)
	// bts: -nullcache=&model.UserTab{TabType:-1} -check_null_code=$==nil -struct_name=Dao -singleflight=true
	UserTab(c context.Context, req *pb.UserTabReq) (*model.UserTab, error)
	// bts: -nullcache=&model.TopPhotoArc{Aid:-1} -check_null_code=$==nil||$.Aid==-1 -struct_name=Dao -singleflight=true
	TopPhotoArc(c context.Context, mid int64) (*model.TopPhotoArc, error)
	// bts: -nullcache=&pb.WhitelistReply{IsWhite:false} -check_null_code=$==nil -struct_name=Dao -singleflight=true
	Whitelist(c context.Context, req *pb.WhitelistReq) (*pb.WhitelistReply, error)
	// bts: -nullcache=&pb.WhitelistValidTimeReply{IsWhite:false} -check_null_code=$==nil -struct_name=Dao -singleflight=true
	QueryWhitelistValid(c context.Context, req *pb.WhitelistReq) (res *pb.WhitelistValidTimeReply, err error)
}
