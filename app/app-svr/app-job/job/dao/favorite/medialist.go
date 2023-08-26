package favorite

import (
	"context"
	"time"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	"git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"
)

// GenMedialist generates a new media list for the given serie
func (d *Dao) GenMedialist(ctx context.Context, serie *show.Serie, cover string) (fid int64, err error) {
	var reply *api.AddFolderReply
	for i := 0; i < 5; i++ {
		if reply, err = d.favClient.AddFolder(ctx, &api.AddFolderReq{
			Name:        serie.MedialistTitle(),
			Description: serie.ShareSubtitle,
			Typ:         int32(favmdl.TypeVideo),
			Mid:         d.conf.WeeklySel.PlaylistMid,
			Public:      favmdl.AttrDefaultPublic,
			Cover:       cover,
		}); err == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err != nil {
		log.Error("GenMedialist SerieType %s, Num %d, Err %v", serie.Type, serie.Number, err)
		return
	}
	fid = reply.Fid
	return
}

// AddMedias add a list of archives into the generated folder
func (d *Dao) AddMedias(ctx context.Context, fid int64, rsc []*show.SerieRes) (err error) {
	var aids []int64
	for _, v := range rsc {
		aids = append(aids, v.RID)
	}
	for i := 0; i < 5; i++ {
		if _, err = d.favClient.MultiAdd(ctx, &api.MultiAddReq{
			Mid:  d.conf.WeeklySel.PlaylistMid,
			Fid:  fid,
			Oids: aids,
			Typ:  int32(favmdl.TypeVideo),
		}); err == nil {
			break
		}
	}
	if err != nil {
		log.Error("AddMedias Fid %d, Err %v", fid, err)
	}
	return
}

// ReplaceMedias replaces a list of archives into the generated folder
func (d *Dao) ReplaceMedias(ctx context.Context, mid, fid int64, aids []int64) (err error) {
	if _, err = d.favClient.MultiReplace(ctx, &api.MultiReplaceReq{
		Mid:  mid,
		Fid:  fid,
		Oids: aids,
		Typ:  int32(favmdl.TypeVideo),
	}); err != nil {
		log.Error("AddMedias Fid %d, Err %v", fid, err)
	}
	return
}
