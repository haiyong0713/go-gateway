package resource

import (
	"context"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	model "go-gateway/app/app-svr/app-feed/admin/model/resource"

	"github.com/pkg/errors"
)

const (
	_simpleArchiveURL = "/videoup/simplearchive"
)

// SimpleArchvie is
func (d *Dao) SimpleArchvie(ctx context.Context, aid int64) (*model.CCArchive, error) {
	params := url.Values{}
	params.Set("aid", strconv.FormatInt(aid, 10))
	out := &struct {
		Code int
		Data model.CCArchive
	}{}
	if err := d.client.Get(ctx, d.simpleArchiveURL, "", params, out); err != nil {
		return nil, err
	}
	if out.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(out.Code), d.simpleArchiveURL+"?"+params.Encode())
		return nil, err
	}
	return &out.Data, nil
}

// 获取稿件信息
func (d *Dao) GetArchiveInfo(c context.Context, avids []int64) (arcs map[int64]*archiveGRPC.Arc, err error) {
	if rep, err := d.archiveClient.Arcs(c, &archiveGRPC.ArcsRequest{Aids: avids}); err != nil {
		return nil, err
	} else {
		return rep.Arcs, nil
	}
}
