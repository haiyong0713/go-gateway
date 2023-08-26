package tag

import (
	"context"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/conf"
	xcode "go-gateway/ecode"
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

// TagByName
func (d *Dao) TagByName(c context.Context, tname string) (*taggrpc.Tag, error) {
	tagRly, e := d.tagClient.TagByName(c, &taggrpc.TagByNameReq{Tname: tname})
	if e != nil {
		if ecode.EqualError(xcode.TagNotExist, e) {
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
