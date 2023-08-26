package reply

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"

	"go-common/library/ecode"

	replyapi "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	"github.com/pkg/errors"
)

var _ dependency.ReplyDependency = &Impl{}

type Impl struct {
	Origin dependency.ReplyDependency

	Reply struct {
		ReplyPreface map[string]*replyapi.ReplyListPrefaceReply
	}
}

func (impl *Impl) GetReplyListPreface(ctx context.Context, _, aid int64, _ string) (*replyapi.ReplyListPrefaceReply, error) {
	key := fmt.Sprintf("%d,%d", aid, model.ReplyTypeAv)
	v, ok := impl.Reply.ReplyPreface[key]
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "key: %s", key)
	}
	return v, nil
}

func (impl *Impl) GetArchiveHonor(c context.Context, aid int64) (*replyapi.ArchiveHonorResp, error) {
	return nil, nil
}
