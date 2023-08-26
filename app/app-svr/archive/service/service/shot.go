package service

import (
	"context"
	"database/sql"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/model"

	"go-gateway/app/app-svr/archive/ecode"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model/videoshot"
)

// Videoshot get video shot.
func (s *Service) Videoshot(c context.Context, aid, cid int64, commonParams *api.CommonParam) (*api.VideoShotReply, error) {
	// check archive&video state
	a, err := s.arc.Arc(c, aid)
	if err != nil {
		return nil, err
	}
	if !a.IsNormal() {
		return nil, ecode.VideoshotNotExist
	}
	attr := a.Attribute
	vs, err := s.arc.NewVideoShotCache(c, cid)
	if err != nil {
		if err != redis.ErrNil {
			return nil, err
		}
		if vs, err = s.arc.RawVideoShot(c, cid); err != nil {
			if err != sql.ErrNoRows {
				return nil, err
			}
			vs = &videoshot.Videoshot{}
		}
		s.addCache(func() {
			_ = s.arc.AddNewVideoShotCache(context.Background(), cid, vs)
		})
	}
	// 视频云保证有高清缩略图时一定有普通缩略图
	if vs == nil || vs.Count <= 0 {
		return nil, ecode.VideoshotNotExist
	}
	res := &api.VideoShotReply{}
	plat := model.Plat(commonParams.GetMobiApp(), commonParams.GetDevice())
	//android 673版本不返回竖版缩略图
	if model.IsAndroid(plat) && commonParams.Build < 6730000 {
		res.Vs = s.buildVs(cid, attr, vs, false, false)
	} else if s.isPortrait(a) && vs.SdCount > 0 && vs.SdImg != "" { //竖屏稿件则返回竖屏缩略图
		res.Vs = s.buildVs(cid, attr, vs, false, true)
	} else {
		res.Vs = s.buildVs(cid, attr, vs, false, false)
	}
	if vs.HDCount > 0 && vs.HDImg != "" {
		res.HdVs = s.buildVs(cid, attr, vs, true, false)
	}
	return res, nil
}

func (s *Service) isPortrait(a *api.Arc) bool {
	//首映稿件首映前是竖屏不展示
	width := a.Dimension.Width
	height := a.Dimension.Height
	//交换位置
	if a.Dimension.Rotate > 0 {
		width, height = height, width
	}
	//是竖屏
	if height > width {
		return true
	}
	return false
}

func (s *Service) buildVs(cid int64, attr int32, vs *videoshot.Videoshot, isHD bool, isPortrait bool) *api.VideoShot {
	if vs == nil {
		log.Error("buildVs err cid(%d)", cid)
		return nil
	}
	shot := &api.VideoShot{
		XLen:  10,
		YLen:  10,
		Image: make([]string, 0, vs.HDCount),
		Attr:  attr,
	}
	if isHD { //高清图
		//http://boss.hdslb.com/videoshotpvhdboss/324889741_afe150-0001.jpg
		shot.XSize = 480
		shot.YSize = 270
		shot.PvData = s.c.Videoshot.BossURI + fmt.Sprintf("%s.bin", vs.HDImg)
		for i := int64(0); i < vs.HDCount; i++ {
			shot.Image = append(shot.Image, s.c.Videoshot.BossURI+fmt.Sprintf("%s-%04d.jpg", vs.HDImg, i+1))
		}
	} else if isPortrait { //竖屏稿件缩略图
		shot.PvData = s.c.Videoshot.BossURI + fmt.Sprintf("%s.bin", vs.SdImg)
		for i := int64(0); i < vs.SdCount; i++ {
			shot.Image = append(shot.Image, s.c.Videoshot.BossURI+fmt.Sprintf("%s-%04d.jpg", vs.SdImg, i+1))
		}
	} else {
		shot.XSize = 160
		shot.YSize = 90
		shot.PvData = s.c.Videoshot.NewURI + fmt.Sprintf("%d.bin", cid)
		for i := int64(0); i < vs.Count; i++ {
			if i == 0 {
				shot.Image = append(shot.Image, s.c.Videoshot.NewURI+fmt.Sprintf("%d.jpg", cid))
				continue
			}
			shot.Image = append(shot.Image, s.c.Videoshot.NewURI+fmt.Sprintf("%d-%d.jpg", cid, i))
		}
	}
	return shot
}
