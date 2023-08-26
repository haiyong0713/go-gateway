package notes

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/app-view/interface/conf"

	"go-gateway/app/app-svr/hkt-note/service/api"
)

type Dao struct {
	notesClient api.HktNoteClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.notesClient, err = api.NewClient(c.NotesClient); err != nil {
		panic(fmt.Sprintf("reply NewClient not found err(%v)", err))
	}
	return
}

// up主是否有笔记
func (d *Dao) IsUpNotes(c context.Context, aid, upId int64) (*api.UpArcReply, error) {
	req := &api.UpArcReq{
		Oid:     aid,
		UpperId: upId,
	}
	reply, err := d.notesClient.UpArc(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// 稿件播放页笔记
func (d *Dao) ArcNote(c context.Context, aid, upId, mid int64, subType int32) (*api.ArcTagReply, error) {
	req := &api.ArcTagReq{
		Oid:       aid,
		UpperId:   upId,
		LoginMid:  mid,
		SubTypeId: subType,
	}
	reply, err := d.notesClient.ArcTag(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
