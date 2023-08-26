package favorite

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"github.com/pkg/errors"
)

// Dao is favorite dao
type Dao struct {
	// grpc
	rpcClient favgrpc.FavoriteClient
}
type BatchIsFavoredsResourcesReq struct {
	Mid  int64
	Aids []int64
	Sids []int64
}

// New initial favorite dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.rpcClient, err = favgrpc.NewClient(c.FavClient); err != nil {
		panic(errors.WithMessage(err, "panic by favgrpc"))
	}
	return
}

// IsFav is favorite
func (d *Dao) IsFav(c context.Context, mid, oid int64, favTp int32) (faved bool) {
	reply, err := d.rpcClient.IsFavored(c, &favgrpc.IsFavoredReq{Typ: favTp, Mid: mid, Oid: oid})
	if err != nil {
		log.Error("%+v", err)
		return false
	}
	if reply != nil {
		return reply.Faved
	}
	return false
}

// AddFav add fav video/season
func (d *Dao) AddFav(c context.Context, mid, oid, action int64, favTp int32, mobiApp, platform, device string) (err error) {
	// tp指收藏夹类型，默认收藏夹=2，otype指oid的类型
	if action == 1 {
		_, err = d.rpcClient.DelFav(c, &favgrpc.DelFavReq{Tp: model.FavTypeVideo, Mid: mid, Oid: oid, Otype: favTp, MobiApp: mobiApp, Platform: platform, Device: device})
	} else {
		_, err = d.rpcClient.AddFav(c, &favgrpc.AddFavReq{Tp: model.FavTypeVideo, Mid: mid, Oid: oid, Otype: favTp, MobiApp: mobiApp, Platform: platform, Device: device})
	}
	return
}

// IsFavoredsResources is fav multi type resource
func (d *Dao) IsFavoredsResources(c context.Context, mid, aid, sid int64) map[int32]bool {
	resourceMap := make(map[int32]*favgrpc.Oids, 2)
	if aid > 0 {
		resourceMap[model.FavTypeVideo] = &favgrpc.Oids{Oid: []int64{aid}}
	}
	if sid > 0 {
		resourceMap[model.FavTypeSeason] = &favgrpc.Oids{Oid: []int64{sid}}
	}
	req := &favgrpc.IsFavoredsResourcesReq{Mid: mid, ResourcesMap: resourceMap}
	reply, err := d.rpcClient.IsFavoredsResources(c, req)
	if err != nil {
		log.Error("d.rpcClient.IsFavoredsResources err(%+v) req(%+v)", err, req)
		return nil
	}
	res := make(map[int32]bool, 2)
	if s, ok := reply.GetFavored()[model.FavTypeVideo]; ok {
		if fv, ok := s.GetOidFavored()[aid]; ok {
			res[model.FavTypeVideo] = fv
		}
	}
	if s, ok := reply.GetFavored()[model.FavTypeSeason]; ok {
		if fs, ok := s.GetOidFavored()[sid]; ok {
			res[model.FavTypeSeason] = fs
		}
	}
	return res
}

// 批量获取收藏 map[int32]map[int64]bool = map[收藏类型]map[资源id]是否收藏
func (d *Dao) BatchIsFavoredsResources(c context.Context, req *BatchIsFavoredsResourcesReq) (map[int32]map[int64]bool, error) {
	resourceMap := make(map[int32]*favgrpc.Oids, 2)
	if len(req.Aids) > 0 {
		resourceMap[model.FavTypeVideo] = &favgrpc.Oids{Oid: req.Aids}
	}
	if len(req.Sids) > 0 {
		resourceMap[model.FavTypeSeason] = &favgrpc.Oids{Oid: req.Sids}
	}
	reqFavor := &favgrpc.IsFavoredsResourcesReq{Mid: req.Mid, ResourcesMap: resourceMap}
	reply, err := d.rpcClient.IsFavoredsResources(c, reqFavor)
	if err != nil {
		log.Error("d.rpcClient.BatchIsFavoredsResources err(%+v) req(%+v)", err, req)
		return nil, err
	}
	res := make(map[int32]map[int64]bool)
	//video favor
	if s, ok := reply.GetFavored()[model.FavTypeVideo]; ok {
		tmp := make(map[int64]bool)
		for _, aid := range req.Aids {
			if fv, ok := s.GetOidFavored()[aid]; ok {
				tmp[aid] = fv
			}
		}
		res[model.FavTypeVideo] = tmp
	}
	//season favor
	if s, ok := reply.GetFavored()[model.FavTypeSeason]; ok {
		tmp := make(map[int64]bool)
		for _, sid := range req.Sids {
			if fs, ok := s.GetOidFavored()[sid]; ok {
				tmp[sid] = fs
			}
		}
		res[model.FavTypeSeason] = tmp
	}
	return res, nil
}
