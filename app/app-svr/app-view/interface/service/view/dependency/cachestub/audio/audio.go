package audio

import (
	"context"

	"go-gateway/app/app-svr/app-view/interface/model/view"
	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"
)

var _ dependency.AudioDependency = &Impl{}

type Impl struct {
	Origin dependency.AudioDependency

	Reply struct {
		Audio map[int64]*view.Audio
	}
}

func (impl *Impl) AudioByCids(ctx context.Context, cids []int64) (map[int64]*view.Audio, error) {
	var res = make(map[int64]*view.Audio)
	for _, cid := range cids {
		if v, ok := impl.Reply.Audio[cid]; ok {
			res[cid] = v
		}
	}
	return res, nil
}
