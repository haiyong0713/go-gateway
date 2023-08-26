package reply

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model"

	api "git.bilibili.co/bapis/bapis-go/community/interface/reply"
)

type Dao struct {
	replyClient api.ReplyInterfaceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.replyClient, err = api.NewClient(c.ReplyClient); err != nil {
		panic(fmt.Sprintf("reply NewClient not found err(%v)", err))
	}
	return
}

// 评论是否变形
func (d *Dao) GetReplyListPreface(c context.Context, mid int64, aid int64, buvid string) (*api.ReplyListPrefaceReply, error) {
	req := api.ReplyListPrefaceReq{
		Mid:   mid,
		Oid:   aid,
		Type:  model.ReplyTypeAv,
		Buvid: buvid,
	}
	res, err := d.replyClient.ReplyListPreface(c, &req)
	if err != nil {
		log.Error("ReplyListPreface fail: mid:%d,aid:%d,err:%+v", mid, aid, err)
		return nil, err
	}
	if res == nil {
		log.Error("ReplyListPreface resp is nil: mid:%d,aid:%d", mid, aid)
		return nil, ecode.NothingFound
	}
	return res, nil
}

func (d *Dao) GetReplyListsPreface(c context.Context, req *api.ReplyListsPrefaceReq) (*api.ReplyListsPrefaceReply, error) {
	return d.replyClient.ReplyListsPreface(c, req)
}

func (d *Dao) GetArchiveHonor(c context.Context, aid int64) (*api.ArchiveHonorResp, error) {
	req := &api.ArchiveHonorReq{
		Kind: api.ArchiveHonorReq_SUPERB_REPLY,
		Aid:  aid,
	}
	res, err := d.replyClient.ArchiveHonor(c, req)
	if err != nil {
		log.Error("ArchiveHonor fail: aid:%d,err:%+v", aid, err)
		return nil, err
	}
	if res == nil {
		log.Error("ArchiveHonor resp is nil: aid:%d", aid)
		return nil, ecode.NothingFound
	}

	return res, nil
}
