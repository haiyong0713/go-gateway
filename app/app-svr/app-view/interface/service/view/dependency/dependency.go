package dependency

import (
	"context"

	appresource "go-gateway/app/app-svr/app-resource/interface/api/v1"
	viewapi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/model/bangumi"
	coin "go-gateway/app/app-svr/app-view/interface/model/coin"
	"go-gateway/app/app-svr/app-view/interface/model/game"
	"go-gateway/app/app-svr/app-view/interface/model/live"
	musicmdl "go-gateway/app/app-svr/app-view/interface/model/music"
	"go-gateway/app/app-svr/app-view/interface/model/view"
	archivehonor "go-gateway/app/app-svr/archive-honor/service/api"
	archive "go-gateway/app/app-svr/archive/service/api"
	resource "go-gateway/app/app-svr/resource/service/api/v1"
	resourcemodel "go-gateway/app/app-svr/resource/service/model"
	steinsgate "go-gateway/app/app-svr/steins-gate/service/api"
	ugcseason "go-gateway/app/app-svr/ugc-season/service/api"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	relation "git.bilibili.co/bapis/bapis-go/account/service/relation"
	ugcpayrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	act "git.bilibili.co/bapis/bapis-go/activity/service"
	advo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"
	channel "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	dm "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	reply "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	music "git.bilibili.co/bapis/bapis-go/crm/service/music-publicity-interface/toplist"
	garb "git.bilibili.co/bapis/bapis-go/garb/model"
	manageractive "git.bilibili.co/bapis/bapis-go/manager/service/active"
	natpage "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	pgcseason "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	videpup "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

type ArchiveDependency interface {
	View3(ctx context.Context, aid, mid int64, mobiApp, device, platform string) (*archive.ViewReply, error)
	ArcsPlayer(ctx context.Context, arcsPlayAv []*archive.PlayAv) (map[int64]*archive.ArcPlayer, error)
	NewRelateAids(ctx context.Context, aid, mid, zoneID int64, build, parentMode, autoplay, isAct int, buvid, sourcePage, trackid, cmd, tabid string, plat int8, pageVersion, fromSpmid string) (res *view.RelateRes, returnCode string, err error)
	Archives(ctx context.Context, aids []int64, mid int64, mobiApp, device string) (map[int64]*archive.Arc, error)
	DescriptionV2(ctx context.Context, aid int64) (desc string, descV2 []*archive.DescV2, mids []int64, err error)
	Argument(ctx context.Context, aid int64) (argueMsg string, err error)
	UpLikeImgCreative(ctx context.Context, mid int64, avid int64) (*viewapi.UpLikeImg, error)
	ArcRedirectUrl(ctx context.Context, aid int64) (*archive.RedirectPolicy, error)
}

type ArchiveHonorDependency interface {
	Honors(ctx context.Context, aid, build int64, mobiApp, device string) ([]*archivehonor.Honor, error)
}

type ArchiveExtraDependency interface {
	GetArchiveExtraValue(ctx context.Context, aid int64) (map[string]string, error)
}

type AccountDependency interface {
	// 账号
	Card3(ctx context.Context, mid int64) (*account.Card, error)
	Cards3(ctx context.Context, mids []int64) (map[int64]*account.Card, error)
	GetInfo(c context.Context, mid int64) (*account.Info, error)
	GetInfos(ctx context.Context, mids []int64) (*account.InfosReply, error)
	IsAttention(ctx context.Context, owners []int64, mid int64) map[int64]int8
	IsNewDevice(ctx context.Context, buvid, periods string) bool
	IsBlueV(ctx context.Context, mid int64) bool
	ContractRelation3(ctx context.Context, mid, owner int64) (*account.ContractRelationReply, error)
}

type RelationDenpendency interface {
	Relation(ctx context.Context, mid, fid int64) (*relation.FollowingReply, error)
	Stat(ctx context.Context, mid int64) (*relation.StatReply, error)
}

type ThumbupDependency interface {
	HasLike(ctx context.Context, mid int64, business, buvid string, aid int64) (thumbup.State, error)
}

type FavDependency interface {
	IsFavoredsResources(ctx context.Context, mid, aid, sid int64) map[int32]bool
}

type CoinDependency interface {
	ArchiveUserCoins(ctx context.Context, aid, mid, avtype int64) (*coin.ArchiveUserCoins, error)
}

type DanmuDependency interface {
	SubjectInfos(ctx context.Context, typ int32, plat int8, oids ...int64) (map[int64]*dm.SubjectInfo, error)
}

type HistoryDependency interface {
	Progress(ctx context.Context, aid, mid int64, buvid string) (*viewapi.History, error)
}

type ReplyDependency interface {
	// 评论
	GetReplyListPreface(ctx context.Context, mid int64, aid int64, buvid string) (*reply.ReplyListPrefaceReply, error)
	// 小黄条
	GetArchiveHonor(c context.Context, aid int64) (*reply.ArchiveHonorResp, error)
}

type AudioDependency interface {
	// 音频
	AudioByCids(ctx context.Context, cids []int64) (map[int64]*view.Audio, error)
}

type UgcpayRankDependency interface {
	// 充电
	RankElecMonthUP(ctx context.Context, upmid, build int64, mobiApp, platform, device string) (*ugcpayrank.RankElecUPResp, error)
	UPRankWithPanelByUPMid(c context.Context, mid, upmid, build int64, mobiApp, platform, device string) (*ugcpayrank.UPRankWithPanelReply, error)
}

type UgcpayDependency interface {
	AssetRelationDetail(ctx context.Context, mid, aid int64, platform string, canPreview bool) (*view.Asset, error)
}

type AssistDependency interface {
	// 协作
	Assist(c context.Context, upMid int64) ([]int64, error)
}

type GarbDependency interface {
	// 装扮
	ThumbupUserEquip(c context.Context, mid int64) (*garb.UserThumbup, error)
}

type UparcDependency interface {
	// up 主稿件
	UpArcCount(ctx context.Context, mid int64) (int64, error)
}

type LiveDependency interface {
	LivingRoom(ctx context.Context, uid int64, platform, brand, net string, build int, mid int64) (*live.Live, error)
}

type SteinsDependency interface {
	View(c context.Context, aid, mid int64, buvid string) (*steinsgate.ViewReply, error)
}

type PgcDependency interface {
	CardsInfoReply(ctx context.Context, seasonIds []int32) (map[int32]*pgcseason.CardInfoProto, error)
	PGC(ctx context.Context, aid, mid int64, build int, mobiApp, device string) (*bangumi.Season, error)
	Movie(ctx context.Context, aid, mid int64, build int, mobiApp, device string) (*bangumi.Movie, error)
}

type LocationDependency interface {
	AuthPIDs(ctx context.Context, pids, ipaddr string) (map[string]*locgrpc.Auth, error)
	Info2(ctx context.Context) (*locgrpc.InfoComplete, error)
	Archive(ctx context.Context, aid, mid int64, ipaddr, cdnip string) (*locgrpc.Auth, error)
	GetGroups(c context.Context, groupId []int64) (map[int64]*locgrpc.Auth, error)
}

type ResourceDependency interface {
	// resource
	GetPlayerCustomizedPanel(ctx context.Context, tids []int64) (*resource.GetPlayerCustomizedPanelV2Rep, error)
	PlayerIcon(ctx context.Context, aid, mid int64, tagIds []int64, typeId int32, showPlayicon bool, build int, mobiApp, device string) (res *resourcemodel.PlayerIcon, err error)
	PlayerIconNew(ctx context.Context, aid, mid int64, tagIds []int64, typeId int32, showPlayicon bool, build int, mobiApp, device string) (res *resourcemodel.PlayerIconRly, err error)
	ViewTab(ctx context.Context, aid int64, tagIDs, upIDs []int64, typeId, plat, build int32) (*viewapi.Tab, error)
}

type AppResourceDependency interface {
	CheckEntranceInfoc(ctx context.Context, in *appresource.CheckEntranceInfocRequest) (*appresource.CheckEntranceInfocReply, error)
}

type VideoupDependency interface {
	ArcViewAddit(ctx context.Context, aid int64) (*videpup.ArcViewAdditReply, error)
	GetMaterialList(ctx context.Context, aid, cid int64) ([]*viewapi.Bgm, []*viewapi.ViewMaterial, []*viewapi.ViewMaterial, error)
}

type MusicDependency interface {
	BgmEntrance(c context.Context, aid, cid int64, platform string) (*musicmdl.Entrance, error)
	ToplistEntrance(c context.Context, aid int64, musicId string) (*music.ToplistEntranceReply, error)
}

type UgcSeasonDependency interface {
	Season(ctx context.Context, seasonID int64) (*ugcseason.View, error)
}

type ChannelDependency interface {
	ResourceChannels(ctx context.Context, aid, mid, ty int64) (res []*channel.Channel, err error)
	ChannelHonor(ctx context.Context, aid int64) (*channel.ResourceHonor, error)
}

type ActivityDependency interface {
	IsReserveAct(ctx context.Context, id, mid int64) bool
	ActProtocol(ctx context.Context, messionID int64) (protocol *act.ActSubProtocolReply, err error)
	LiveBooking(c context.Context, mid, upmid int64) (*act.UpActReserveRelationInfo, error)
}

type NatPageDependency interface {
	NatInfoFromForeign(ctx context.Context, tids []int64, pageType int64, content map[string]string) (map[int64]*natpage.NativePage, error)
}

type ManagerDependency interface {
	CommonActivity(c context.Context, sid, mid int64, asPlat int32) (*manageractive.CommonActivityResp, error)
}

type AdDependency interface {
	AdGRPC(ctx context.Context, mobiApp, buvid, device string, build int, mid, upperID, aid int64, rid int32, tids []int64, resource []int32, network, adExtra, spmid, fromSpmid, from string, adTab bool) (*advo.SunspotAdReplyForView, error)
}

type GameDependency interface {
	Info(ctx context.Context, gameID int64, plat int8) (info *game.Info, err error)
}

type ViewDependency struct {
	Archive       ArchiveDependency
	ArchiveHornor ArchiveHonorDependency
	ArchiveExtra  ArchiveExtraDependency
	Account       AccountDependency
	Relation      RelationDenpendency
	ThumbUP       ThumbupDependency
	Fav           FavDependency
	Coin          CoinDependency
	Danmu         DanmuDependency
	History       HistoryDependency
	Reply         ReplyDependency
	Audio         AudioDependency
	UGCPayRank    UgcpayRankDependency
	UGCPay        UgcpayDependency
	Assist        AssistDependency
	Garb          GarbDependency
	UpArc         UparcDependency
	Live          LiveDependency
	Steins        SteinsDependency
	PGC           PgcDependency
	Location      LocationDependency
	Resource      ResourceDependency
	AppResource   AppResourceDependency
	VideoUP       VideoupDependency
	UGCSeason     UgcSeasonDependency
	Channel       ChannelDependency
	Activity      ActivityDependency
	NatPage       NatPageDependency
	Manager       ManagerDependency
	AD            AdDependency
	Game          GameDependency
	Music         MusicDependency
}
