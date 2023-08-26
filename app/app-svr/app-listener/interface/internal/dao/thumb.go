package dao

import (
	"context"
	"time"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	avecode "go-gateway/app/app-svr/archive/ecode"

	arcSvc "go-gateway/app/app-svr/archive/service/api"

	thumbupSvc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	thumbErr "git.bilibili.co/bapis/bapis-go/community/service/thumbup/ecode"
	"github.com/pkg/errors"
)

type ThumbActionOpt struct {
	Mid          int64
	Dev          *device.Device
	Net          *network.Network
	ItemType     int32
	Oid, SubID   int64
	Action       v1.ThumbUpReq_ThumbType
	NoSilver     bool // 不走风控
	IsTripleLike bool
}

func (d *dao) ThumbAction(ctx context.Context, opt ThumbActionOpt) (err error) {
	var messageID, upID int64
	silverOpt := interactSilverOpt{
		ItemType: playItem2GaiaItemType[opt.ItemType],
		Oid:      opt.Oid,
	}
	if opt.IsTripleLike {
		silverOpt.Scene, silverOpt.Action = _silverSceneVideoTripleLike, _silverActionVideoTripleLike
	}

	switch opt.ItemType {
	case model.PlayItemUGC:
		messageID = opt.Oid
	case model.PlayItemOGV:
		epdetail, err := d.Epids2Aids(ctx, []int32{int32(opt.Oid)})
		if err != nil {
			return err
		}
		aid, ok := epdetail[int32(opt.Oid)]
		if !ok {
			return errors.WithMessagef(avecode.ArchiveNotExist, "can not found corresponding avid for episode(%d)", opt.Oid)
		}
		messageID = aid
	case model.PlayItemAudio:
		auInfos, err := d.SongInfosV1(ctx, SongInfosOpt{SongIds: []int64{opt.Oid}, RemoteIP: opt.Net.RemoteIP})
		if err != nil {
			return err
		}
		au, ok := auInfos[opt.Oid]
		if !ok {
			return errors.WithMessagef(avecode.ArchiveNotExist, "can not found correspondng auinfo for auid(%d)", opt.Oid)
		}
		if !au.IsNormal() {
			return errors.WithMessagef(avecode.ArchiveNotExist, "audio is not normal. detail(%+v)", au)
		}
		upID, messageID = au.Author.Mid, opt.Oid
	}
	// ugc/ogv 通用获取信息
	if opt.ItemType == model.PlayItemUGC || opt.ItemType == model.PlayItemOGV {
		arcInfo, err := d.arcGRPC.Arc(ctx, &arcSvc.ArcRequest{Aid: opt.Oid})
		if err != nil {
			return wrapDaoError(err, "arcGRPC.Arc", opt.Oid)
		}
		if arcInfo.Arc.State < 0 {
			return errors.WithMessagef(avecode.ArchiveNotExist, "archive(%d) state<0", opt.Oid)
		}

		silverOpt.PubTime = arcInfo.GetArc().GetPubDate().Time().Format(time.RFC3339)
		upID = arcInfo.Arc.Author.Mid
		silverOpt.UpMid, silverOpt.Title = upID, arcInfo.Arc.Title
		silverOpt.PlayNum = arcInfo.Arc.Stat.View

		// 先判断风控
		if opt.Action == v1.ThumbUpReq_LIKE && !opt.NoSilver {
			resp, err := d.thumbSilver(ctx, silverOpt)
			if err == nil {
				if resp.IsRejected() {
					return ErrSilverBulletHit
				}
			} else {
				log.Warnc(ctx, "SilverBullet failed to check thumb event(%+v)", opt)
			}
		}
	}

	if opt.Mid > 0 {
		req := &thumbupSvc.LikeReq{
			Mid: opt.Mid, Action: thumbAction(opt.Action),
			MessageID: messageID,
			UpMid:     upID, Business: thumbBusiness(opt.ItemType),
			IP: opt.Net.RemoteIP, MobiApp: opt.Dev.RawMobiApp,
			Platform: opt.Dev.RawPlatform, Device: opt.Dev.Device,
			From: thumbupSvc.From_SourceFromListenerSingle,
		}
		_, err = d.thumbupGRPC.Like(ctx, req)
		err = wrapDaoError(err, "thumbupGRPC.Like", req)
	} else {
		req := &thumbupSvc.BuvidLikeReq{
			Buvid: opt.Dev.Buvid, Action: thumbAction(opt.Action),
			MessageID: messageID,
			UpMid:     upID, Business: thumbBusiness(opt.ItemType),
			IP: opt.Net.RemoteIP, MobiApp: opt.Dev.RawMobiApp,
			Platform: opt.Dev.RawPlatform, Device: opt.Dev.Device,
			From: thumbupSvc.From_SourceFromListenerSingle,
		}
		_, err = d.thumbupGRPC.BuvidLike(ctx, req)
		err = wrapDaoError(err, "thumbupGRPC.BuvidLike", req)
	}
	if err != nil &&
		!ecode.EqualError(thumbErr.ThumbupDupLikeErr, err) &&
		!ecode.EqualError(thumbErr.ThumbupDupDislikeErr, err) &&
		!ecode.EqualError(thumbErr.ThumbupCancelLikeErr, err) &&
		!ecode.EqualError(thumbErr.ThumbupCancelDislikeErr, err) {
		return err
	} else if err == nil {
		d.thumbReport(ctx, opt.Dev, opt.Net, opt.Mid, opt.Oid, opt.ItemType, opt.Action)
	}

	return nil
}

func thumbAction(act v1.ThumbUpReq_ThumbType) thumbupSvc.Action {
	switch act {
	case v1.ThumbUpReq_CANCEL_LIKE:
		return thumbupSvc.Action_ACTION_CANCEL_LIKE
	case v1.ThumbUpReq_DISLIKE:
		return thumbupSvc.Action_ACTION_DISLIKE
	case v1.ThumbUpReq_CANCEL_DISLIKE:
		return thumbupSvc.Action_ACTION_CANCEL_DISLIKE
	default:
		return thumbupSvc.Action_ACTION_LIKE
	}
}

func thumbBusiness(itemType int32) string {
	switch itemType {
	case model.PlayItemUGC, model.PlayItemOGV:
		return ThumbUpBusinessUGCOGV
	case model.PlayItemAudio:
		return ThumbUpBusinessAudio
	default:
		return ""
	}
}
