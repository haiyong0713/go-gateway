package dao

import (
	"context"

	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

// Dao dao interface
//
//go:generate kratos tool btsgen
type _bts interface {
	Close()
	Ping(ctx context.Context) (err error)
	// bts:-struct_name=Dao -nullcache=&model.Goods{Id:-1} -check_null_code=$!=nil&&$.Id==-1
	Goods(c context.Context, id int64) (*model.Goods, error)
	// bts:-struct_name=Dao -nullcache=[]*model.Goods{{Id:-1}} -check_null_code=len($)!=0&&$[0].Id==-1
	GoodsList(c context.Context) ([]*model.Goods, error)
	// bts:-struct_name=Dao -nullcache=&model.Task{Id:-1} -check_null_code=$!=nil&&$.Id==-1
	Task(c context.Context, uid int64) (*model.Task, error)
	// bts:-struct_name=Dao -nullcache=[]*model.Invite{{Uid:-1}} -check_null_code=len($)!=0&&$[0].Uid==-1
	InviteList(c context.Context, uid int64, id int64) ([]*model.Invite, error)
	// bts:-struct_name=Dao -nullcache=[]*model.Task{{Id:-1}} -check_null_code=len($)!=0&&$[0].Id==-1
	PrizeList(c context.Context) ([]*model.Task, error)
	// bts:-struct_name=Dao -nullcache="NULL" -check_null_code="NULL"
	Config(c context.Context, key string) (string, error)
}
