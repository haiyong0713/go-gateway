package archive

import (
	"context"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/ecode"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/job/model/archive"
	"go-gateway/app/app-svr/archive/service/api"
)

func (d *Dao) GrpcRawVideos(c context.Context, aid int64) ([]*archive.Video, error) {
	rly, err := d.creativeGRPC.GetArchiveVideoRelation(c, &creativeAPI.GetArchiveVideoRelationReq{Aid: aid})
	if err != nil {
		return nil, err
	}
	var as []*archive.Video
	for _, v := range rly.GetRelation() {
		if v == nil {
			continue
		}
		as = append(as, &archive.Video{
			Cid:        v.GetCid(),
			SrcType:    v.GetSrcType(),
			Index:      int(v.GetIndexOrder()),
			Title:      v.GetTitle(),
			Duration:   v.GetDuration(),
			Filename:   v.GetFilename(),
			Status:     int16(v.GetStatus()),
			State:      int16(v.GetState()),
			Dimensions: v.GetDimensions(),
		})
	}
	return as, nil
}

func (d *Dao) GrpcRawArchive(c context.Context, aid int64) (*api.Arc, error) {
	arcRly, err := d.creativeGRPC.GetArchive(c, &creativeAPI.GetArchiveReq{Aid: aid})
	if err != nil {
		return nil, err
	}
	if arcRly == nil {
		return nil, ecode.NothingFound
	}
	return &api.Arc{
		Aid:       arcRly.GetAid(),
		Author:    api.Author{Mid: arcRly.GetMid()},
		TypeID:    arcRly.GetTypeid(),
		Copyright: arcRly.GetCopyright(),
		Title:     arcRly.GetTitle(),
		Pic:       arcRly.GetCover(),
		Desc:      arcRly.GetContent(),
		Tag:       arcRly.GetTag(),
		Duration:  arcRly.GetDuration(),
		Attribute: int32(arcRly.GetAttribute()),
		Access:    int32(arcRly.GetAccess()),
		State:     arcRly.GetState(),
		PubDate:   xtime.Time(arcRly.GetPubtime()),
		Forward:   arcRly.GetForward(),
	}, nil
}
