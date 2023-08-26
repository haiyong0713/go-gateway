package feedcard

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/game"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	shopping "go-gateway/app/app-svr/app-card/interface/model/card/show"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	jsonlargecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	jsonselect "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/select"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	feedMgr "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	resourceV2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	viprpc "git.bilibili.co/bapis/bapis-go/vip/service"
)

type FanoutResult struct {
	Archive struct {
		Archive      map[int64]*arcgrpc.ArcPlayer
		StoryArchive map[int64]*arcgrpc.ArcPlayer
	}
	Tag  map[int64]*taggrpc.Tag
	Live struct {
		Room       map[int64]*live.Room
		InlineRoom map[int64]*live.Room
	}
	Article map[int64]*article.Meta
	Audio   map[int64]*audio.Audio
	Dynamic struct {
		Picture map[int64]*bplus.Picture
	}
	Bangumi struct {
		EP                map[int64]*bangumi.EpPlayer
		Season            map[int32]*episodegrpc.EpisodeCardsProto
		SeasonByAid       map[int32]*episodegrpc.EpisodeCardsProto
		InlinePGC         map[int32]*pgcinline.EpisodeCard
		Remind            *bangumi.Remind
		Update            *bangumi.Update
		PgcEpisodeByAids  map[int64]*pgccard.EpisodeCard
		PgcEpisodeByEpids map[int32]*pgccard.EpisodeCard
		PgcSeason         map[int32]*pgcAppGrpc.SeasonCardInfoProto
		EpMaterial        map[int64]*deliverygrpc.EpMaterial
	}
	Channel map[int64]*channelgrpc.ChannelCard
	Tunnel  map[int64]*tunnelgrpc.FeedCard
	ThumbUp struct {
		HasLikeArchive map[int64]int8
	}
	Account struct {
		Card            map[int64]*accountgrpc.Card
		RelationStatMid map[int64]*relationgrpc.StatReply
		IsAttention     map[int64]int8
	}
	Banner struct {
		Banners []*banner.Banner
		Version string
	}
	Inline     *jsonlargecover.Inline
	FollowMode *jsonselect.FollowMode
	StoryIcon  map[int64]*appcardmodel.GotoIcon
	Shop       map[int64]*shopping.Shopping
	Vip        *viprpc.TipsRenewReply
	Favourite  map[int64]int8
	HotAidSet  sets.Int64
	Coin       map[int64]int64
	LiveBadge  struct {
		LeftBottomBadgeStyle *operate.LiveBottomBadge
		LeftCoverBadgeStyle  []*operate.V9LiveLeftCoverBadge
	}
	MultiMaterials     map[int64]*feedMgr.Material
	Specials           map[int64]*operate.Card
	ThreePointMetaText *threePointMeta.ThreePointMetaText
	Game               map[int64]*game.Game
	Reservation        map[int64]*activitygrpc.UpActReserveRelationInfo
	SpecialCard        map[int64]*resourceV2grpc.AppSpecialCard
	OpenCourseMark     map[int64]bool
	LikeStatState      map[int64]*thumbupgrpc.StatState
}
