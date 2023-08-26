package archive

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/time"
	xtime "time"

	"go-gateway/app/app-svr/archive/job/model/archive"
	"go-gateway/app/app-svr/archive/service/model/videoshot"
)

// RawStaff get Staff by aid.
func (d *Dao) RawStaff(c context.Context, aid int64) (res []*archive.Staff, err error) {
	apiRes, err := d.RawStaff4API(c, aid)
	if err != nil {
		return nil, err
	}
	return apiRes, nil
}

func (d *Dao) RawStaff4API(c context.Context, aid int64) ([]*archive.Staff, error) {
	apiRes := []*archive.Staff{}
	reply, err := d.GetArchiveStaff(c, aid)
	if err != nil {
		log.Error("RawStaff4API rpc is err %+v %+v", aid, err)
		return nil, err
	}
	for _, v := range reply.GetArchiveStaff() {
		apiRes = append(apiRes, &archive.Staff{
			Aid:        v.GetAid(),
			Mid:        v.GetStaffMid(),
			Title:      v.GetStaffTitle(),
			Ctime:      xtime.Unix(v.GetCtime(), 0).Format(xtime.RFC3339),
			IndexOrder: v.GetIndexOrder(),
			Attribute:  v.GetAttribute(),
		})
	}
	return apiRes, nil
}

// RawArchive get a archive by avid.
func (d *Dao) RawArchive(c context.Context, aid int64) (a *archive.Archive, err error) {
	apiRes, err := d.RawArchive4API(c, aid)
	if err != nil {
		if ecode.EqualError(ecode.NothingFound, err) {
			log.Error("d.RawArchive4API res is nil %+v %+v", aid, err)
			return nil, nil
		}
		return nil, err
	}
	return apiRes, nil
}

func (d *Dao) RawArchive4API(c context.Context, aid int64) (*archive.Archive, error) {
	res, err := d.GetArchive(c, aid)
	if err != nil {
		log.Error("d.GetArchive rpc is err %+v %+v", aid, err)
		return nil, err
	}
	a := &archive.Archive{
		ID:        res.GetAid(),
		Mid:       res.GetMid(),
		TypeID:    int16(res.GetTypeid()),
		Duration:  int(res.GetDuration()),
		Title:     res.GetTitle(),
		Cover:     res.GetCover(),
		Content:   res.GetContent(),
		Tag:       res.GetTag(),
		Attribute: int32(res.GetAttribute()),
		Copyright: int8(res.GetCopyright()),
		State:     int(res.GetState()),
		Author:    res.GetAuthor(),
		Access:    int(res.GetAccess()),
		Forward:   int(res.GetForward()),
		PubTime:   xtime.Unix(res.GetPubtime(), 0).Format(xtime.RFC3339),
		Round:     int8(res.GetRound()),
		CTime:     xtime.Unix(res.GetCtime(), 0).Format(xtime.RFC3339),
		MTime:     time.Time(res.GetMtime()),
		Reason:    res.RejectReason,
	}
	return a, nil
}

// RawAddit get archive addit.
func (d *Dao) RawAddit(c context.Context, aid int64) (addit *archive.Addit, err error) {
	apiRes, err := d.RawAddit4API(c, aid)
	if err != nil || apiRes == nil {
		return nil, nil
	}
	return apiRes, nil
}

func (d *Dao) RawAddit4API(c context.Context, aid int64) (addit *archive.Addit, err error) {
	res, err := d.GetArchiveAddit(c, aid)
	if err != nil {
		log.Error("d.GetArchiveAddit rpc is err %+v %+v", aid, err)
		return nil, err
	}
	addit = &archive.Addit{
		ID:          res.GetId(),
		Aid:         res.GetAid(),
		Desc:        res.GetDescription(),
		Source:      res.GetSource(),
		RedirectURL: res.GetRedirectUrl(),
		MissionID:   res.GetMissionId(),
		UpFrom:      int32(res.GetUpFrom()),
		OrderID:     int(res.GetOrderId()),
		Dynamic:     res.GetDynamic(),
		InnerAttr:   res.GetInnerAttr(),
		Ipv6:        res.GetIpv6(),
	}
	return addit, nil
}

// RawBiz get archive biz.
func (d *Dao) RawBiz(c context.Context, aid int64, state int, bizType int) (*archive.Biz, error) {
	apiRes, err := d.RawBiz4API(c, aid, state, bizType)
	if err != nil {
		if ecode.EqualError(ecode.NothingFound, err) {
			return nil, nil
		}
		log.Error("d.GetArchiveBiz res is err %+v %+v", aid, err)
		return nil, err
	}
	return apiRes, nil
}

func (d *Dao) RawBiz4API(c context.Context, aid int64, state int, bizType int) (*archive.Biz, error) {
	res, err := d.GetArchiveBiz(c, aid, state, bizType)
	if err != nil {
		return nil, err
	}
	reply := &archive.Biz{
		Aid:     res.GetAid(),
		Data:    res.GetData(),
		SubType: int(res.GetSubType()),
	}
	return reply, nil

}

// RawVideos get videos by 2 table em.......
func (d *Dao) RawVideos(c context.Context, aid int64) (vs []*archive.Video, err error) {
	apiRes, err := d.RawVideos4API(c, aid)
	if err != nil {
		return nil, err
	}
	return apiRes, nil
}

func (d *Dao) RawVideos4API(c context.Context, aid int64) (vs []*archive.Video, err error) {
	res, err := d.GetArchiveVideoRelation(c, aid)
	if err != nil {
		log.Error("d.GetArchiveVideoRelation is err %+v %+v", err, aid)
		return nil, err
	}
	for _, v := range res {
		vs = append(vs, &archive.Video{
			ID:          v.GetId(),
			Aid:         v.GetAid(),
			Title:       v.GetTitle(),
			Desc:        v.GetDescription(),
			Filename:    v.GetFilename(),
			SrcType:     v.GetSrcType(),
			Cid:         v.GetCid(),
			Duration:    v.GetDuration(),
			Filesize:    v.GetFilesize(),
			Resolutions: v.GetResolutions(),
			Index:       int(v.GetIndexOrder()),
			CTime:       xtime.Unix(v.GetCtime(), 0).Format(xtime.RFC3339),
			MTime:       xtime.Unix(v.GetMtime(), 0).Format(xtime.RFC3339),
			Status:      int16(v.GetStatus()),
			State:       int16(v.GetState()),
			Playurl:     v.GetPlayurl(),
			Attribute:   int32(v.GetAttribute()),
			FailCode:    int8(v.GetFailcode()),
			XcodeState:  int8(v.GetXcodeState()),
			Dimensions:  v.GetDimensions(),
		})
	}
	return vs, nil
}

// RawVideoShots is
func (d *Dao) RawVideoShots(c context.Context, cids []int64) (vs []*videoshot.Videoshot, err error) {
	apiRes, err := d.RawVideoShots4API(c, cids)
	if err != nil {
		log.Error("d.GetArchiveVideoShot is err %+v %+v", err, cids)
		return nil, err
	}
	return apiRes, nil

}

func (d *Dao) RawVideoShots4API(c context.Context, cids []int64) (vs []*videoshot.Videoshot, err error) {
	res, err := d.GetArchiveVideoShot(c, cids)
	if err != nil {
		log.Error("d.GetArchiveVideoShot is err %+v %+v", err, cids)
		return nil, err
	}
	for _, v := range res {
		tmp := &videoshot.Videoshot{
			Cid:     v.Id,
			Count:   v.Count,
			HDImg:   v.HdImage,
			HDCount: v.HdCount,
			SdCount: v.SdCount,
			SdImg:   v.SdImage,
		}
		vs = append(vs, tmp)
	}
	return vs, nil
}

// RawGetFirstPassByAID is
func (d *Dao) RawGetFirstPassByAID(c context.Context, aid int64) (id int64, err error) {
	apiRes, err := d.RawGetFirstPassByAID4API(c, aid)
	if err != nil {
		if ecode.EqualError(ecode.NothingFound, err) {
			log.Error("d.GetArchiveFirstPass res is nil %+v %+v", aid, err)
			return id, nil
		}
		log.Error("d.GetArchiveFirstPass res is err %+v %+v", aid, err)
		return id, err
	}
	return apiRes, nil
}

func (d *Dao) RawGetFirstPassByAID4API(c context.Context, aid int64) (id int64, err error) {
	res, err := d.GetArchiveFirstPass(c, aid)
	if err != nil {
		return 0, err
	}
	return res.GetId(), nil
}

// RawTypes is second types opposite first types.
func (d *Dao) RawTypes(c context.Context) (types []*archive.ArcType, err error) {
	res, err := d.GetArchiveType(c)
	if err != nil {
		log.Error("d.GetArchiveType res is err %+v ", err)
		return nil, err
	}
	for _, v := range res {
		types = append(types, &archive.ArcType{
			ID:   v.GetId(),
			PID:  v.GetPid(),
			Name: v.GetName(),
		})
	}
	return types, nil
}
