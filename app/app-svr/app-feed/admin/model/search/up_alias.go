package search

import (
	"fmt"
	"go-common/library/time"
	pb "go-gateway/app/app-svr/app-feed/admin/api/search"
	"strings"
)

type UpAlias struct {
	Id          int64     `gorm:"id" json:"id"`                     // 记录id
	Mid         int64     `gorm:"mid" json:"mid"`                   // 用户mid
	Nickname    string    `gorm:"nickname" json:"nickname"`         // 用户昵称
	SearchWords string    `gorm:"search_words" json:"search_words"` // 搜索词
	Stime       time.Time `gorm:"stime" json:"stime"`               // 生效开始时间
	Etime       time.Time `gorm:"etime" json:"etime"`               // 生效结束时间
	IsForever   int32     `gorm:"is_forever" json:"is_forever"`     // 是否永久
	Applier     string    `gorm:"applier" json:"applier"`           // 申请人
	State       int32     `gorm:"state" json:"state"`               // 是否在线
	Ctime       time.Time `gorm:"ctime" json:"ctime"`               // 创建时间
	Mtime       time.Time `gorm:"mtime" json:"mtime"`               // 创建时间
}

func (entity *UpAlias) TableName() string {
	return "search_up_alias"
}

func (entity *UpAlias) GetEntityForPB() *pb.UpAlias {
	return &pb.UpAlias{
		Id:          entity.Id,
		Mid:         entity.Mid,
		Nickname:    entity.Nickname,
		SearchWords: entity.SearchWords,
		Stime:       entity.Stime.Time().Unix(),
		Etime:       entity.Etime.Time().Unix(),
		IsForever:   entity.IsForever,
		Applier:     entity.Applier,
		State:       entity.State,
		Ctime:       entity.Ctime.Time().Unix(),
	}
}

func (entity *UpAlias) GetEntityForSyncPB() *pb.SyncUpAlias {
	return &pb.SyncUpAlias{
		Id:          entity.Id,
		Mid:         entity.Mid,
		Nickname:    entity.Nickname,
		SearchWords: strings.Split(entity.SearchWords, ","),
		Mtime:       entity.Mtime.Time().Unix(),
	}
}

func (entity *UpAlias) GetEntityForExport() string {
	return fmt.Sprintf(
		"%d\t%d\t%s\t%s\t%d\t%d\t%d\t%s\t%d\t%d",
		entity.Id,
		entity.Mid,
		entity.Nickname,
		entity.SearchWords,
		entity.Stime.Time().Unix(),
		entity.Etime.Time().Unix(),
		entity.IsForever,
		entity.Applier,
		entity.State,
		entity.Ctime.Time().Unix(),
	)
}
