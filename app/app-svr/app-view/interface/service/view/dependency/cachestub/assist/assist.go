package assist

import (
	"context"

	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"

	"go-common/library/ecode"

	"github.com/pkg/errors"
)

var _ dependency.AssistDependency = &Impl{}

type Impl struct {
	Origin dependency.AssistDependency

	Reply struct {
		Assist map[int64][]int64
	}
}

func (impl *Impl) Assist(ctx context.Context, upMid int64) ([]int64, error) {
	v, ok := impl.Reply.Assist[upMid]
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "upMid: %d", upMid)
	}
	return v, nil
}
