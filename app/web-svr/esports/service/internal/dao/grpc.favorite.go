package dao

import (
	"context"
	favpb "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	api "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/esports/service/api/v1"
	favmdl "go-main/app/community/favorite/service/model"
)

const (
	_defaultFid = 0
)

func (d *dao) AddFav(ctx context.Context, contestId int64, mid int64) (err error) {
	arg := &api.AddFavReq{Tp: favpb.TypeEsports, Mid: mid, Oid: contestId, Fid: _defaultFid}
	if _, err = d.favoriteClient.AddFav(ctx, arg); err != nil {
		log.Errorc(ctx, "[Dao][FavoriteClient][AddFav]s.favClient.AddFav(%+v) error(%v)", arg, err)
		return
	}
	return
}

func (d *dao) DelFav(ctx context.Context, contestId int64, mid int64) (err error) {
	arg := &api.DelFavReq{Tp: int32(favpb.TypeEsports), Mid: mid, Oid: contestId, Fid: 0}
	if _, err = d.favoriteClient.DelFav(ctx, arg); err != nil {
		log.Error("[Dao][FavoriteClient][DelFav]s.favClient.DelFav(%+v) error(%v)", arg, err)
		return
	}
	return
}

func (d *dao) GetSubscriberByContestId(ctx context.Context, contestId int64, cursor int64, cursorSize int32) (res *v1.ContestSubscribers, err error) {
	var favRs *api.SubscribersReply
	res = new(v1.ContestSubscribers)
	arg := &api.SubscribersReq{Type: int32(favmdl.TypeEsports), Oid: contestId, Cursor: cursor, Size_: cursorSize}
	if favRs, err = d.favoriteClient.Subscribers(ctx, arg); err != nil {
		log.Errorc(ctx, "SubContestUserV2 s.favClient.Subscribers, contestId:%d, cursor:%d, size:%d, error(%+v)", contestId, cursor, cursorSize, err)
		return
	}
	res.Cursor = favRs.Cursor
	res.User = make([]*v1.User, 0)
	if len(favRs.User) == 0 {
		return
	}
	for _, users := range favRs.User {
		res.User = append(res.User, &v1.User{
			Id:    users.Id,
			Oid:   users.Oid,
			Mid:   users.Mid,
			Typ:   users.Typ,
			State: users.State,
			Ctime: users.Ctime,
			Mtime: users.Mtime,
		})
	}
	return
}

func (d *dao) GetSubscribeRelationByContests(ctx context.Context, contestIds []int64, mid int64) (relations map[int64]bool, err error) {
	relations = make(map[int64]bool)
	var favRes *api.IsFavoredsReply
	if mid > 0 {
		if favRes, err = d.favoriteClient.IsFavoreds(ctx, &api.IsFavoredsReq{Typ: int32(favmdl.TypeEsports), Mid: mid, Oids: contestIds}); err != nil {
			log.Error("[Dao][favoriteClient][IsFavoreds]s.favClient.IsFavoreds(%d,%+v) error(%+v)", mid, contestIds, err)
			err = nil
			return
		}
		if favRes != nil {
			relations = favRes.Faveds
		}
	}
	return
}
