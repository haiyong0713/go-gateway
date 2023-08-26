package favor

import (
	"context"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"
)

var _ dependency.FavDependency = &Impl{}

type Impl struct {
	Origin dependency.FavDependency

	Reply struct {
		ArchiveIsFavor map[int32]map[int64]bool
	}
}

func (impl *Impl) IsFavoredsResources(ctx context.Context, mid, aid, sid int64) map[int32]bool {
	res := make(map[int32]bool)
	if aid > 0 {
		if videoFv, ok := impl.Reply.ArchiveIsFavor[model.FavTypeVideo]; ok {
			if isFavor, ok := videoFv[aid]; ok {
				res[model.FavTypeVideo] = isFavor
			}
		}
	}
	if sid > 0 {
		if seasonFv, ok := impl.Reply.ArchiveIsFavor[model.FavTypeSeason]; ok {
			if isFavor, ok := seasonFv[sid]; ok {
				res[model.FavTypeSeason] = isFavor
			}
		}
	}
	return res
}
