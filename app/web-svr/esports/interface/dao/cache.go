package dao

import (
	"context"

	mdlEp "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -struct_name=Dao
	EpContests(c context.Context, ids []int64) (map[int64]*model.Contest, error)
	// bts: -struct_name=Dao
	EpSeasons(c context.Context, ids []int64) (map[int64]*model.Season, error)
	// bts: -struct_name=Dao
	EpTeams(c context.Context, ids []int64) (map[int64]*model.Team, error)
	// bts: -struct_name=Dao
	SearchMainIDs(c context.Context) ([]int64, error)
	// bts: -struct_name=Dao
	SearchMD(c context.Context, mainIDs []int64) (res map[int64]*model.SearchRes, err error)
	// bts: -struct_name=Dao
	EpGames(c context.Context, ids []int64) (map[int64]*mdlEp.Game, error)
	// bts: -struct_name=Dao
	EpGameMap(c context.Context, oids []int64, tp int64) (map[int64]int64, error)
	// bts: -struct_name=Dao
	LolGames(c context.Context, matchID int64) ([]*model.LolGame, error)
	// bts: -struct_name=Dao
	DotaGames(c context.Context, matchID int64) ([]*model.LolGame, error)
	// bts: -struct_name=Dao
	OwGames(c context.Context, matchID int64) ([]*model.OwGame, error)
	// bts: -struct_name=Dao
	SeasonGames(c context.Context) ([]int64, error)
	//bts: -nullcache=[]*model.MatchSeason{{SeasonID:-1}} -check_null_code=len($)==1&&$[0].SeasonID==-1 -struct_name=Dao
	FetchSeasonsByMatchId(c context.Context, matchID int64) ([]*model.MatchSeason, error)
	// bts: -nullcache=&model.MatchSeason{SeasonID:-1} -check_null_code=$.SeasonID==-1 -struct_name=Dao
	FetchSeasonsInfoMap(c context.Context, sids []int64) (map[int64]*model.MatchSeason, error)
	// bts: -nullcache=&model.VideoListInfo{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	VideoList(c context.Context, id int64) (*model.VideoListInfo, error)
}
