package tag

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/native-page/interface/conf"
	tagEcode "go-main/app/community/tag/ecode"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

type Dao struct {
	tagClient taggrpc.TagRPCClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.tagClient, err = taggrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	return
}

// UpdateExtraAttr .
func (d *Dao) UpdateExtraAttr(c context.Context, tid int64, state int32) error {
	// state 0-取消活动属性 1-设置活动属性
	_, e := d.tagClient.UpdateExtraAttr(c, &taggrpc.UpdateExtraAttrReq{Tid: tid, Type: 1, State: state})
	if e != nil {
		log.Error("d.tagClient.UpdateExtraAttr(%d,%d) error(%v)", tid, state, e)
	}
	return e
}

// TagByName
func (d *Dao) TagByName(c context.Context, tname string) (*taggrpc.Tag, error) {
	tagRly, e := d.tagClient.TagByName(c, &taggrpc.TagByNameReq{Tname: tname})
	if e != nil {
		if ecode.EqualError(tagEcode.TagNotExist, e) {
			return nil, nil
		}
		log.Error(" d.tagClient.TagByName(%s) error(%v)", tname, e)
		return nil, e
	}
	if tagRly != nil && tagRly.Tag != nil {
		return tagRly.Tag, nil
	}
	return nil, nil
}

// AddTag
func (d *Dao) AddTag(c context.Context, tagName string, mid int64) (*taggrpc.TagReply, error) {
	rly, err := d.tagClient.AddTag(c, &taggrpc.AddTagReq{Mid: mid, Name: tagName})
	if err != nil {
		log.Error("d.tagClient.AddTag(%d,%s) error(%v)", mid, tagName, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) AddSub(c context.Context, arg *taggrpc.AddSubReq) error {
	_, err := d.tagClient.AddSub(c, arg)
	return err
}
