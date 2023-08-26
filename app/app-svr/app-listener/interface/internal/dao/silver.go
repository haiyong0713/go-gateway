package dao

import (
	"context"
	stderr "errors"
	"strconv"
	"strings"

	"go-common/library/silverbullet/gaia"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	"go-gateway/pkg/riskcontrol"

	"github.com/pkg/errors"
)

const (
	_silverSceneVideoThumbUp    = "thumbup_video"
	_silverSceneVideoCoin       = "video_coin"
	_silverSceneVideoCoinLike   = "video_cointolike"
	_silverSceneVideoTripleLike = "video_triplelike"
	_silverSceneVideoFav        = "video_fav"

	_silverActionVideoLike       = "like"
	_silverActionVideoCoin       = "video_coin"
	_silverActionVideoFav        = "video_fav"
	_silverActionVideoCoinLike   = "video_cointolike"
	_silverActionVideoTripleLike = "video_triplelike"
)

var (
	playItem2GaiaItemType = map[int32]string{
		model.PlayItemUGC:   "av",
		model.PlayItemOGV:   "bangumi",
		model.PlayItemAudio: "audio",
	}
	// 被风控
	ErrSilverBulletHit = stderr.New("silverBullet hit")
)

type interactSilverOpt struct {
	Scene    string `json:"-"`
	Action   string `json:"action"`
	Oid      int64  `json:"avid"`
	ItemType string `json:"item_type"`
	Title    string `json:"title,omitempty"`
	UpMid    int64  `json:"up_mid"`
	CoinNum  int32  `json:"coin_num,omitempty"`
	PubTime  string `json:"pubtime"`
	PlayNum  int32  `json:"play_num,omitempty"`
}

type gaiaReport struct {
	*gaia.GaiaReport
}

func (g gaiaReport) IsRejected() bool {
	if !g.IsHit() {
		return false
	}
	for _, r := range g.Decisions {
		if strings.Contains(strings.ToLower(r), "reject") {
			return true
		}
	}
	return false
}

func (d *dao) prepareSliver(ctx context.Context, scene string, opt interactSilverOpt, setEctx ...func(gaia.CheckInterface)) (*gaiaReport, error) {
	if d.silverBullet == nil {
		return &gaiaReport{&gaia.GaiaReport{}}, nil
	}

	check, err := d.silverBullet.InitCheck(ctx, scene)
	if err != nil {
		return nil, errors.WithMessagef(err, "silverBullet.InitCheck failed scene(%s)", scene)
	}
	check.Put("data_source", "listener")
	check.Put("token", riskcontrol.ReportedLoginTokenFromCtx(ctx))
	err = check.AutoPut(opt)
	if err != nil {
		return nil, errors.WithMessagef(err, "silverBullet failed to put opt(%+v) into checkCtx", opt)
	}
	for _, f := range setEctx {
		f(check)
	}
	resp, err := check.Do()
	if err != nil {
		return nil, errors.WithMessagef(err, "silverBullet failed to do Check(%+v)", check)
	}
	return &gaiaReport{resp}, nil
}

func (d *dao) thumbSilver(ctx context.Context, opt interactSilverOpt) (*gaiaReport, error) {
	if len(opt.Action) <= 0 {
		opt.Action = _silverActionVideoLike
	}
	if len(opt.Scene) <= 0 {
		opt.Scene = _silverSceneVideoThumbUp
	}
	setEctx := func(g gaia.CheckInterface) {
		g.Put("av_title", opt.Title)
	}
	return d.prepareSliver(ctx, opt.Scene, opt, setEctx)
}

func (d *dao) coinSilver(ctx context.Context, opt interactSilverOpt) (*gaiaReport, error) {
	if len(opt.Action) <= 0 {
		opt.Action = _silverActionVideoCoin
	}
	if len(opt.Scene) <= 0 {
		opt.Scene = _silverSceneVideoCoin
	}
	return d.prepareSliver(ctx, opt.Scene, opt)
}

func (d *dao) favSilver(ctx context.Context, opt interactSilverOpt) (*gaiaReport, error) {
	if len(opt.Action) <= 0 {
		opt.Action = _silverActionVideoFav
	}
	if len(opt.Scene) <= 0 {
		opt.Scene = _silverSceneVideoFav
	}
	setEctx := func(c gaia.CheckInterface) {
		c.Put("uid", strconv.FormatInt(opt.UpMid, 10))
		c.Put("av_id", strconv.FormatInt(opt.Oid, 10))
		c.Put("fav_source", 2)
		c.Put("play_num_string", strconv.FormatInt(int64(opt.PlayNum), 10))
	}
	return d.prepareSliver(ctx, opt.Scene, opt, setEctx)
}
