package client

import (
	"context"
	"go-common/library/log"
	http "go-common/library/net/http/blademaster"
	"go-common/library/sync/errgroup"
	archive "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	espServiceClient "go-gateway/app/web-svr/esports/service/api/v1"
	"sync"

	dynapi "git.bilibili.co/bapis/bapis-go/dynamic/service/publish"

	api2 "git.bilibili.co/bapis/bapis-go/account/service/oauth2"
	api "git.bilibili.co/bapis/bapis-go/passport/service/sns"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"

	tunnel "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"

	relationAPI "git.bilibili.co/bapis/bapis-go/account/service/relation"
	audit "git.bilibili.co/bapis/bapis-go/aegis/strategy/service"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	tvbwapi "git.bilibili.co/bapis/bapis-go/bw/game/common"
	cheesePayApi "git.bilibili.co/bapis/bapis-go/cheese/service/pay"
	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	datamartapi "git.bilibili.co/bapis/bapis-go/crm/service/datamart"
	upgroupRpc "git.bilibili.co/bapis/bapis-go/crm/service/profile-manager"
	upratingRpc "git.bilibili.co/bapis/bapis-go/crm/service/uprating"
	fligrpc "git.bilibili.co/bapis/bapis-go/filter/service"
	grab "git.bilibili.co/bapis/bapis-go/garb/service"
	liveActivityapi "git.bilibili.co/bapis/bapis-go/live/activity-task/grpc"
	livedataapi "git.bilibili.co/bapis/bapis-go/live/data-guru/v1"
	liveapi "git.bilibili.co/bapis/bapis-go/live/xroom"
	naPage "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	seasonapi "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	actPlatform "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	videoup "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	actApi "go-gateway/app/web-svr/activity/interface/api"
)

const (
	_aidBulkSize = 50
)

var (
	ActPlatClient      actPlatform.ActPlatClient
	ArchiveClient      archive.ArchiveClient
	GarbClient         grab.GarbClient
	NaPageClient       naPage.NaPageClient
	AccountClient      accapi.AccountClient
	ArticleClient      artapi.ArticleGRPCClient
	SeasonClient       seasonapi.SeasonClient
	DataMartClient     datamartapi.DataMartClient
	FilterClient       fligrpc.FilterClient
	VideoClient        videoup.VideoUpOpenClient
	RelationClient     relationAPI.RelationClient
	LiveActivityClient liveActivityapi.ActivityInfoClient
	SilverbulletClient silverbulletapi.GaiaClient
	CheesePayClient    cheesePayApi.PayClient
	RatingClient       upratingRpc.UpRatingClient
	UpGroupClient      upgroupRpc.ProfileManagerClient
	TagClient          tagrpc.TagRPCClient
	TunnelClient       tunnel.TunnelClient
	AuditClient        audit.AegisStrategyServiceClient
	LiveClient         liveapi.RoomClient
	PublishClient      dynapi.PublishClient
	TvBwClient         tvbwapi.BwInterfaceClient
	HttpClient         *http.Client
	PassportClient     api.PassportSNSClient
	BiliOAuth2Client   api2.Oauth2Client
	EspServiceClient   espServiceClient.EsportsServiceClient
	ActivityClient     actApi.ActivityClient
	LiveDataClient     livedataapi.DCTaskManagerClient
)

func New(cfg *conf.Config) {
	initialize.NewE(actPlatform.NewClient, func() (err error) {
		ActPlatClient, err = actPlatform.NewClient(nil)
		return
	})
	var err error
	ArchiveClient, err = archive.NewClient(cfg.ArchiveClient)
	if err != nil {
		panic(err)
	}
	GarbClient, err = grab.NewClient(cfg.GarbClient)
	if err != nil {
		panic(err)
	}
	NaPageClient, err = naPage.NewClient(cfg.NaPageClient)
	if err != nil {
		panic(err)
	}
	AccountClient, err = accapi.NewClient(cfg.AccClient)
	if err != nil {
		panic(err)
	}
	ArticleClient, err = artapi.NewClient(cfg.ArtClient)
	if err != nil {
		panic(err)
	}
	SeasonClient, err = seasonapi.NewClient(cfg.SeasonClient)
	if err != nil {
		panic(err)
	}
	DataMartClient, err = datamartapi.NewClient(cfg.DataMartClient)
	if err != nil {
		panic(err)
	}
	FilterClient, err = fligrpc.NewClient(cfg.FliClient)
	if err != nil {
		panic(err)
	}
	VideoClient, err = videoup.NewClient(cfg.VideoClient)
	if err != nil {
		panic(err)
	}
	RelationClient, err = relationAPI.NewClient(cfg.RelationClient)
	if err != nil {
		panic(err)
	}
	SilverbulletClient, err = silverbulletapi.NewClient(cfg.SilverGaiaClient)
	if err != nil {
		panic(err)
	}
	if CheesePayClient, err = cheesePayApi.NewClient(cfg.CheesePayClient); err != nil {
		{
			panic(err)
		}
	}
	LiveActivityClient, err = liveActivityapi.NewClient(cfg.LiveActivityClient)
	if err != nil {
		panic(err)
	}

	TagClient, err = tagrpc.NewClient(cfg.TagClient)
	if err != nil {
		panic(err)
	}
	if RatingClient, err = upratingRpc.NewClient(cfg.UpClientNew); err != nil {
		panic(err)
	}

	if UpGroupClient, err = upgroupRpc.NewClient(cfg.UpClientNew); err != nil {
		panic(err)
	}
	if TunnelClient, err = tunnel.NewClient(cfg.TunnelClient); err != nil {
		panic(err)
	}
	if AuditClient, err = audit.NewClient(cfg.AuditClient); err != nil {
		panic(err)
	}
	if LiveClient, err = liveapi.NewClient(cfg.LiveXRoomClient); err != nil {
		panic(err)
	}
	if TvBwClient, err = tvbwapi.NewClient(cfg.TvBwClient); err != nil {
		panic(err)
	}
	if PassportClient, err = api.NewClient(cfg.PassportSNSClient); err != nil {
		panic(err)
	}
	if BiliOAuth2Client, err = api2.NewClient(cfg.BiliOAuth2Client); err != nil {
		panic(err)
	}
	if EspServiceClient, err = espServiceClient.NewClient(cfg.EsportSercieClient); err != nil {
		panic(err)
	}
	if LiveDataClient, err = livedataapi.NewClient(cfg.LiveDataClient); err != nil {
		panic(err)
	}
	if PublishClient, err = dynapi.NewClient(cfg.PublishClient); err != nil {
		panic(err)
	}
	HttpClient = http.NewClient(cfg.HTTPClient)
	if ActivityClient, err = actApi.NewLocalClient(cfg.ActivityClient); err != nil {
		panic(err)
	}
}

func Archives(c context.Context, aids []int64) (archives map[int64]*archive.Arc, err error) {
	var (
		mutex         = sync.Mutex{}
		aidsLen       = len(aids)
		group, errCtx = errgroup.WithContext(c)
	)
	archives = make(map[int64]*archive.Arc, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func() (err error) {
			var arcs *archive.ArcsReply
			arg := &archive.ArcsRequest{Aids: partAids}
			if arcs, err = ArchiveClient.Arcs(errCtx, arg); err != nil || arcs == nil {
				log.Errorc(c, "Vote.Archives (%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, v := range arcs.Arcs {
				archives[v.Aid] = v
			}
			mutex.Unlock()
			return
		})
	}
	err = group.Wait()
	return

}
