package thumbup

import (
	"context"

	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"

	"go-common/library/ecode"

	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"

	"github.com/pkg/errors"
)

var _ dependency.ThumbupDependency = &Impl{}

type Impl struct {
	Origin dependency.ThumbupDependency

	Reply struct {
		ArchiveHasLike map[int64]thumbup.State
	}
}

func (impl *Impl) HasLike(ctx context.Context, _ int64, _, _ string, aid int64) (thumbup.State, error) {
	state, ok := impl.Reply.ArchiveHasLike[aid]
	if !ok {
		return 0, errors.Wrapf(ecode.NothingFound, "aid: %d", aid)
	}
	return state, nil
}
