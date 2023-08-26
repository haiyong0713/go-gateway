package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/archive"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	steampunkgrpc "git.bilibili.co/bapis/bapis-go/pcdn/steampunk"
	batch "git.bilibili.co/bapis/bapis-go/video/vod/playurlugcbatch"
	volume "git.bilibili.co/bapis/bapis-go/video/vod/playurlvolume"
)

func (s *Service) baseForBatchAv(c context.Context, aids []int64, mid int64, platform, device, buvid string, areaLimitValidate bool) (am map[int64]*api.Arc, history map[int64]*hisApi.ModelHistory, isVip, isControl bool, err error) {
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		var incArcs map[int64]*api.ArcInternal
		if am, incArcs, err = s.ArchivesAndInc(ctx, aids, mid, platform, device); err != nil {
			return err
		}
		//地区限制校验
		if areaLimitValidate && len(am) > 0 {
			_ = s.archiveAutoPlayValidate(ctx, am, incArcs)
		}
		return nil
	})
	// 历史进度增加降级开关
	if !s.c.Switch.HistorySeek && (buvid != "" || mid != 0) {
		eg.Go(func(ctx context.Context) (err error) {
			history, err = s.hisdao.Progress(ctx, aids, mid, buvid)
			if err != nil {
				log.Error("s.hisdao.Progress aids(%+v) mid(%d) buvid(%s) error(%+v) or res is nil", aids, mid, buvid, err)
				return nil
			}
			return nil
		})
	}
	if mid > 0 {
		// vip管控开关开启&&只针对粉版修改
		witchControl := false
		if s.c.Switch.VipControl && (platform == "iphone" || platform == "android") {
			witchControl = true
		}
		eg.Go(func(ctx context.Context) error {
			infoRly, e := s.arc.VipInfo(ctx, mid, buvid, witchControl)
			if e != nil || infoRly == nil {
				log.Error("s.arc.VipInfo(%d) or info is nil error(%v)", mid, e)
				return nil
			}
			if infoRly.Res != nil {
				isVip = infoRly.Res.IsValid()
			}
			if infoRly.Control != nil && infoRly.Control.Control {
				isControl = true
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("baseForBatchAv eg.wait() err(%+v) aids(%+v) mid(%d)", err, aids, mid)
		return nil, nil, false, false, err
	}
	return am, history, isVip, isControl, nil
}

func (s *Service) batchPlayURL(c context.Context, vItem []*batch.RequestVideoItem, pAids []int64, arg *api.BatchPlayArg) (ugcItem map[uint64]*batch.ResponseItem, pgcItem map[int64]*archive.PGCPlayurl, location *locgrpc.InfoCompleteReply, volItem map[uint64]*volume.VolumeItem) {
	if (arg.MobiApp == "iphone" && arg.Build <= model.QnIOSBuild) || (arg.MobiApp == "android" && arg.Build < model.QnAndroidBuild) {
		arg.Qn = s.c.Custom.PlayerQn
	}
	eg := errgroup.WithContext(c)
	ugcItem = make(map[uint64]*batch.ResponseItem)
	pgcItem = make(map[int64]*archive.PGCPlayurl)
	volItem = make(map[uint64]*volume.VolumeItem)

	var locInfo *locgrpc.InfoCompleteReply
	var pcdnItem map[uint64]*steampunkgrpc.CidResources
	if len(vItem) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			// player
			if ugcItem, err = s.playurldao.PlayurlBatch(ctx, vItem, arg, s.c.Custom.BackupNum, true); err != nil {
				log.Error("ArcsWithSP s.playurldao.PlayurlBatch err(%+v)", err)
			}
			return nil
		})
		//获取pcdn
		if steamHash(arg.Buvid, "2233") <= s.c.Custom.PCDNGrey {
			s.infoProm.Incr("bpcdn-gray-hit")
			cids := make([]uint64, 0, len(vItem))
			for _, item := range vItem {
				cids = append(cids, item.GetCid())
			}
			eg.Go(func(ctx context.Context) (err error) {
				if pcdnItem, err = s.playurldao.BatchGetPcdnUrl(ctx, cids, arg); err != nil {
					log.Error("s.playurldao.BatchGetPcdnUrl err(%+v)", err)
				}
				return nil
			})
		}
	}
	if !s.c.Switch.VoiceBalance && arg.VoiceBalance == 1 && len(vItem) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			cids := make([]uint64, 0, len(vItem))
			for _, item := range vItem {
				cids = append(cids, item.Cid)
			}
			if volItem, err = s.playurldao.PlayurlVolume(ctx, cids, arg); err != nil {
				log.Error("batchPlayURL s.playurldao.PlayurlVolume cids(%+v) arg(%+v) err(%+v)", cids, arg, err)
			}
			return nil
		})
	}
	if len(pAids) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			// pgc player
			if pgcItem, err = s.arc.PGCPlayURLs(ctx, pAids, arg.MobiApp, arg.Ip, arg.Session, arg.Fnval, arg.Fnver); err != nil {
				log.Error("ArcsWithAP s.arc.PGCPlayurls err(%+v)", err)
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) (err error) {
		locInfo, err = s.locGRPC.Info2(ctx, &locgrpc.AddrReq{Addr: arg.Ip})
		if err != nil {
			log.Error("s.locGRPC.Info2 err(%+v) ip(%s)", err, arg.Ip)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("batchPlayURL eg.wait() err(%+v) vItem(%+v) padis(%+v)", err, vItem, pAids)
		return nil, nil, nil, nil
	}
	for cid, item := range ugcItem {
		if i, ok := pcdnItem[cid]; ok {
			s.joinPCDNToPlayUrlInfo(i.Data, item)
		}
	}
	return ugcItem, pgcItem, locInfo, volItem
}

func (s *Service) playerProgress(history map[int64]*hisApi.ModelHistory, aid, cid int64) int64 {
	h, ok := history[aid]
	if !ok || h == nil || h.Cid != cid {
		return 0
	}
	if h.Pro > 0 {
		progressRatio := int64(1000)
		return h.Pro * progressRatio
	}
	return h.Pro
}

//nolint:gomnd
func steamHash(str, salt string) int64 {
	// md5
	b := []byte(str)
	s := []byte(salt)
	h := md5.New()
	h.Write(b) // 先写盐值
	h.Write(s)
	src := h.Sum(nil)

	var dst = make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)

	// to upper
	dst = bytes.ToUpper(dst)

	// javahash
	var n int32 = 0
	for i := 0; i < len(dst); i++ {
		n = n*31 + int32(dst[i])
	}

	// mod 100000
	if n >= 0 {
		return int64(n % 100000)
	} else {
		return int64(n%100000 + 100000)
	}
}
