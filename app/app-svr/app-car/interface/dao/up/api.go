package up

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	ml "go-gateway/app/app-svr/app-car/interface/model/medialist"

	"github.com/pkg/errors"
)

const (
	_typeUpArcs = 1
	_oTypeUgc   = 2
)

// MediaList http://bapi.bilibili.co/project/3492/interface/api/157694
func (d *Dao) MediaList(ctx context.Context, req *ml.MediaListReq) (*ml.MediaListRes, error) {
	params := req.ToUrlValues()
	res := &ml.MediaListRes{}
	err := d.client.Get(ctx, d.mediaListUrl, "", params, res)
	if err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		log.Error("MediaList req:%+v res:%+v", req, res)
		return nil, errors.Wrapf(ecode.Int(res.Code), res.Message)
	}
	return res, nil
}
