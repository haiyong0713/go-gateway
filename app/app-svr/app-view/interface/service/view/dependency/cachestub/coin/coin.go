package coin

import (
	"context"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/app-view/interface/model/coin"
	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"

	"github.com/pkg/errors"
)

var _ dependency.CoinDependency = &Impl{}

type Impl struct {
	Origin dependency.CoinDependency

	Reply struct {
		ArchiveUserCoins map[int64]int64
	}
}

func (impl *Impl) ArchiveUserCoins(ctx context.Context, aid, _, _ int64) (*coin.ArchiveUserCoins, error) {
	v, ok := impl.Reply.ArchiveUserCoins[aid]
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "aid: %d", aid)
	}
	return &coin.ArchiveUserCoins{Multiply: v}, nil
}
