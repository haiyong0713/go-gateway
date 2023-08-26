package history

import (
	"context"
	"go-common/library/ecode"
	viewapi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"

	"github.com/pkg/errors"

	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

var _ dependency.HistoryDependency = &Impl{}

type Impl struct {
	Origin dependency.HistoryDependency

	Reply struct {
		ArchiveProgress map[int64]*hisgrpc.ModelHistory
	}
}

func (impl *Impl) Progress(ctx context.Context, aid, _ int64, _ string) (*viewapi.History, error) {
	v, ok := impl.Reply.ArchiveProgress[aid]
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "aid: %d", aid)
	}
	return &viewapi.History{Cid: v.Cid, Progress: v.Pro}, nil
}
