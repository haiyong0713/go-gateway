package dao

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"sync"
	"time"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	avecode "go-gateway/app/app-svr/archive/ecode"

	arcSvc "go-gateway/app/app-svr/archive/service/api"

	coinSvc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

type CoinAddOpt struct {
	Mid        int64
	Dev        *device.Device
	Net        *network.Network
	ItemType   int32
	Oid, SubID int64
	CoinNum    int32
	ThumbUp    bool
	NoSilver   bool
}

func (d *dao) CoinAdd(ctx context.Context, opt CoinAddOpt) (err error) {
	req := &coinSvc.AddCoinReq{
		MaxCoin: 1, Number: int64(opt.CoinNum),
		Mid: opt.Mid, Business: coinBusiness(opt.ItemType),
		MobiApp: opt.Dev.RawMobiApp, Device: opt.Dev.Device,
		Platform: opt.Dev.RawPlatform,
		From:     coinSvc.From_SourceFromListenerSingle,
	}
	autoNumber := opt.CoinNum == 0

	thumbOpt := ThumbActionOpt{
		Mid: opt.Mid, Dev: opt.Dev, Net: opt.Net,
		Oid: opt.Oid, SubID: opt.SubID, ItemType: opt.ItemType,
		NoSilver: true,
	}
	silverOpt := interactSilverOpt{
		Scene:    _silverSceneVideoCoin,
		Action:   _silverActionVideoCoin,
		ItemType: playItem2GaiaItemType[opt.ItemType],
		Oid:      opt.Oid, CoinNum: opt.CoinNum,
	}
	if opt.ThumbUp {
		silverOpt.Action = _silverActionVideoCoinLike
		silverOpt.Scene = _silverSceneVideoCoinLike
	}

	var arcInfo *arcSvc.Arc
	switch opt.ItemType {
	case model.PlayItemUGC:
		resp, err := d.arcGRPC.Arc(ctx, &arcSvc.ArcRequest{Aid: opt.Oid})
		if err != nil {
			return wrapDaoError(err, "arcGRPC.Arc", opt.Oid)
		}
		if resp.GetArc().GetState() < 0 {
			return errors.WithMessagef(avecode.ArchiveNotExist, "archive(%d) state<0", opt.Oid)
		}
		req.Aid = opt.Oid
		if autoNumber {
			req.Number = 1
		}
		if resp.GetArc().GetCopyright() == 1 {
			if autoNumber {
				req.Number = 2
			}
			req.MaxCoin = 2
		}
		arcInfo = resp.GetArc()
		req.PubTime = arcInfo.GetPubDate().Time().Unix()
		req.Upmid = arcInfo.GetAuthor().Mid
		req.Typeid = arcInfo.GetTypeID()

		thumbOpt.ItemType = model.PlayItemUGC
	case model.PlayItemOGV:
		ep2av, err := d.Epids2Aids(ctx, []int32{int32(opt.Oid)})
		if err != nil {
			return err
		}
		var ok bool
		req.Aid, ok = ep2av[int32(opt.Oid)]
		if !ok {
			return errors.WithMessagef(avecode.ArchiveNotExist, "failed to find corresponding avid for epsiode(%d)", opt.Oid)
		}
		resp, err := d.arcGRPC.Arc(ctx, &arcSvc.ArcRequest{Aid: req.Aid})
		if err != nil {
			return wrapDaoError(err, "arcGRPC.Arc", req.Aid)
		}
		if resp.GetArc().GetState() < 0 {
			return errors.WithMessagef(avecode.ArchiveNotExist, "archive(%d) state<0", req.Aid)
		}
		if autoNumber {
			req.Number = 2
		}
		req.MaxCoin = 2
		arcInfo = resp.GetArc()
		req.PubTime = arcInfo.GetPubDate().Time().Unix()
		req.Upmid = arcInfo.GetAuthor().Mid
		req.Typeid = arcInfo.GetTypeID()

		thumbOpt.ItemType = model.PlayItemOGV
	case model.PlayItemAudio:
		req.Aid = opt.Oid
		if autoNumber {
			req.Number = 2
		}
		req.MaxCoin = 2
		auInfos, err := d.SongInfosV1(ctx, SongInfosOpt{SongIds: []int64{opt.Oid}, RemoteIP: opt.Net.RemoteIP})
		if err != nil {
			return err
		}
		au, ok := auInfos[opt.Oid]
		if !ok {
			return errors.WithMessagef(avecode.ArchiveNotExist, "audio(%d) not found", opt.Oid)
		}
		if !au.IsNormal() {
			return errors.WithMessagef(avecode.ArchiveNotExist, "audio is not normal details(%+v)", au)
		}
		req.PubTime = au.Ctime.Time().Unix()
		req.Upmid = au.Author.Mid

		thumbOpt.ItemType = model.PlayItemAudio
	}

	// 稿件走风控
	if !opt.NoSilver && (opt.ItemType == model.PlayItemUGC || opt.ItemType == model.PlayItemOGV) {
		silverOpt.PubTime = arcInfo.PubDate.Time().Format(time.RFC3339)
		silverOpt.UpMid, silverOpt.Title = arcInfo.GetAuthor().Mid, arcInfo.Title
		silverOpt.PlayNum = arcInfo.Stat.View

		if resp, err := d.coinSilver(ctx, silverOpt); err != nil {
			log.Warnc(ctx, "SilverBullet failed to check coin event(%+v)", opt)
		} else {
			if resp.IsRejected() {
				return ErrSilverBulletHit
			}
		}
	}

	eg := errgroup.WithContext(ctx)
	eg.Go(func(c context.Context) error {
		if opt.ThumbUp {
			if err := d.ThumbAction(c, thumbOpt); err != nil {
				log.Warnc(c, "failed to thumbup while coin2like: %v", err)
			}
		}
		return nil
	})

	eg.Go(func(c context.Context) (err error) {
		_, err = d.coinGRPC.AddCoin(ctx, req)
		if err != nil {
			err = wrapDaoError(err, "coinGRPC.AddCoin", req)
		}
		return
	})
	err = eg.Wait()
	// 交互成功再上报
	if err == nil {
		d.coinAddReport(ctx, opt.Dev, opt.Net, opt.Mid, req.Aid, opt.ItemType, req.Number)
	}

	return
}

func coinBusiness(typ int32) string {
	switch typ {
	case model.PlayItemUGC, model.PlayItemOGV:
		return CoinBusinessUGCOGV
	case model.PlayItemAudio:
		return CoinBusinessAudio
	default:
		return ""
	}
}

type CoinNumsOpt struct {
	Business string
	Oids     []int64
	Net      *network.Network
}

var coinBusiness2AvType = map[string]string{
	CoinBusinessUGCOGV:  "1",
	CoinBusinessArticle: "2",
	CoinBusinessAudio:   "3",
}

const (
	// 获取稿件硬币数
	_coinCounts = "/x/internal/v1/coin/creation/counts"
)

func (d *dao) CoinNums(ctx context.Context, opt CoinNumsOpt) (ret map[int64]int64, err error) {
	ret = make(map[int64]int64)
	avType := coinBusiness2AvType[opt.Business]
	eg := errgroup.WithContext(ctx)
	type coinNumResp struct {
		Count int64 `json:"count"`
	}
	mu := sync.Mutex{}
	for _, oid := range opt.Oids {
		oCopy := oid
		eg.Go(func(c context.Context) error {
			p := url.Values{
				"aid":    []string{strconv.FormatInt(oCopy, 10)},
				"avtype": []string{avType},
				"ip":     []string{opt.Net.RemoteIP},
			}
			res := &model.BmGenericResp{}
			err := d.coinHTTP.Get(c, _coinCounts, p, res)
			if err != nil {
				return errors.WithMessagef(err, "failed to get coinNums for oid(%d) business(%s)", oCopy, opt.Business)
			}
			if err = res.IsNormal(); err != nil {
				return errors.WithMessagef(err, "resp not normal while getting coinNums for oid(%d) business(%s) msg(%s)", oCopy, opt.Business, res.Msg)
			}
			data := coinNumResp{}
			if err = json.Unmarshal(res.Data.Bytes(), &data); err != nil {
				return errors.WithMessagef(err, "error unmarshal get coinNums resp for oid(%d) business(%s) data(%s)", oCopy, opt.Business, res.Data)
			}
			mu.Lock()
			ret[oCopy] = data.Count
			mu.Unlock()
			return nil
		})
	}
	err = eg.Wait()

	return
}
