package bapi

import (
	"regexp"
	"strings"
)

var (
	grpcMd = map[string]map[string]map[string][]string{
		"account.service": {
			"account.service": {
				"Account": {
					"/account.service.Account/Info3",
					"/account.service.Account/Infos3",
					"/account.service.Account/InfosByName3",
					"/account.service.Account/Card3",
					"/account.service.Account/Cards3",
					"/account.service.Account/ProfileWithoutPrivacy3",
					"/account.service.Account/ProfilesWithoutPrivacy3",
					"/account.service.Account/Profile3",
					"/account.service.Account/ProfileWithStat3",
					"/account.service.Account/ProfileStatPrivacyByAuth3",
					"/account.service.Account/AddExp3",
					"/account.service.Account/AddMoral3",
					"/account.service.Account/Relation3",
					"/account.service.Account/Attentions3",
					"/account.service.Account/Blacks3",
					"/account.service.Account/Relations3",
					"/account.service.Account/RichRelations3",
					"/account.service.Account/Vip3",
					"/account.service.Account/Vips3",
					"/account.service.Account/Prompting",
				},
			},
		},
		"community.service.dm-service": {
			"dm.service.api": {
				"DM": {
					"/dm.service.api.DM/AddDM",
					"/dm.service.api.DM/ChangeDMState",
					"/dm.service.api.DM/ChangeDMPool",
					"/dm.service.api.DM/BuyAdvance",
					"/dm.service.api.DM/AdvanceState",
					"/dm.service.api.DM/Advances",
					"/dm.service.api.DM/PassAdvance",
					"/dm.service.api.DM/DenyAdvance",
					"/dm.service.api.DM/CancelAdvance",
					"/dm.service.api.DM/AddUserFilters",
					"/dm.service.api.DM/UserFilters",
					"/dm.service.api.DM/DelUserFilters",
					"/dm.service.api.DM/AddUpFilters",
					"/dm.service.api.DM/UpFilters",
					"/dm.service.api.DM/BanUsers",
					"/dm.service.api.DM/CancelBanUsers",
					"/dm.service.api.DM/EditUpFilters",
					"/dm.service.api.DM/AddGlobalFilter",
					"/dm.service.api.DM/GlobalFilters",
					"/dm.service.api.DM/DelGlobalFilters",
					"/dm.service.api.DM/FlushUpFilterCache",
					"/dm.service.api.DM/FlushUserFilterCache",
					"/dm.service.api.DM/FlushDMAdvanceCommentCache",
					"/dm.service.api.DM/FlushDMFilterGlobalCache",
					"/dm.service.api.DM/FlushDMMaskingCache",
				},
			},
		},
		" sycpb.cpm.bce-rtb-webapp": {
			"bcg.rtb.ad.api": {
				"Rtb": {
					"/bcg.rtb.ad.api.Rtb/AdSearch",
				},
			},
		},
		"cheese.service.auth": {
			"cheese.auth.service.v1": {
				"Auth": {
					"/cheese.auth.service.v1.Auth/CanWatchBySeasonId",
					"/cheese.auth.service.v1.Auth/CanWatchByEpisodeId",
					"/cheese.auth.service.v1.Auth/CanWatchByAid",
					"/cheese.auth.service.v1.Auth/SeasonPlayStatus",
					"/cheese.auth.service.v1.Auth/favoriteCount",
				},
			},
		},
		"account.service.oauth2": {
			"account.service.oauth2": {
				"Oauth2": {
					"/account.service.oauth2.Oauth2/AccessToken",
					"/account.service.oauth2.Oauth2/User",
					"/account.service.oauth2.Oauth2/UserOpenID",
					"/account.service.oauth2.Oauth2/MidByOpenID",
				},
			},
		},
		"live.xroomfeed": {
			"live.xroomfeed.v1": {
				"Dynamic": {
					"/live.xroomfeed.v1.Dynamic/getLiveIDs",
					"/live.xroomfeed.v1.Dynamic/getHistoryCardInfo",
					"/live.xroomfeed.v1.Dynamic/getCardInfo",
				},
			},
		},
		"main.account.vas-resource-service": {
			"vas.resource.v1": {
				"Resource": {
					"/vas.resource.v1.Resource/ResourceInfo",
					"/vas.resource.v1.Resource/ResourceInfosByBizId",
					"/vas.resource.v1.Resource/ResourceUse",
					"/vas.resource.v1.Resource/ResourceUseAsync",
					"/vas.resource.v1.Resource/ResourceUsage",
					"/vas.resource.v1.Resource/ResourceUseByPhone",
					"/vas.resource.v1.Resource/ResourcePreStore",
					"/vas.resource.v1.Resource/ResourcePreStoreGive",
					"/vas.resource.v1.Resource/ResourcesGetUsageTimeList",
					"/vas.resource.v1.Resource/ResourceOpenRecord",
					"/vas.resource.v1.Resource/ResourceGrantRecords",
					"/vas.resource.v1.Resource/CodeInfo",
					"/vas.resource.v1.Resource/CodeOpen",
					"/vas.resource.v1.Resource/CodeOpened",
					"/vas.resource.v1.Resource/CodesUseTimeList",
				},
			},
		},
		"cheese.service.coupon": {
			"cheese.service.coupon.v2": {
				"CouponPlatform": {
					"/cheese.service.coupon.v2.CouponPlatform/UpdateStatus",
					"/cheese.service.coupon.v2.CouponPlatform/CouponUsing",
					"/cheese.service.coupon.v2.CouponPlatform/CouponSimpleUse",
					"/cheese.service.coupon.v2.CouponPlatform/MyCouponList",
					"/cheese.service.coupon.v2.CouponPlatform/CouponInfos",
					"/cheese.service.coupon.v2.CouponPlatform/MyAvailableCoupons",
					"/cheese.service.coupon.v2.CouponPlatform/BatchBindRules",
				},
			},
		},
		"live.xfansmedal": {
			"live.xfansmedal": {
				"Reply": {
					"/live.xfansmedal.Reply/GetMedals",
				},
				"Anchor": {
					"/live.xfansmedal.Anchor/QueryMedal",
					"/live.xfansmedal.Anchor/UserReceptions",
				},
			},
		},
		"community.service.history": {
			"community.service.history": {
				"History": {
					"/community.service.history.History/AddHistory",
					"/community.service.history.History/Progress",
					"/community.service.history.History/Position",
					"/community.service.history.History/ClearHistory",
					"/community.service.history.History/Histories",
					"/community.service.history.History/HistoryCursor",
					"/community.service.history.History/Delete",
					"/community.service.history.History/FlushHistory",
				},
			},
		},
		"pgc.service.card": {
			"pgc.service.card.app": {
				"AppCard": {
					"/pgc.service.card.app.AppCard/FeedCardWithFollow",
					"/pgc.service.card.app.AppCard/TagCards",
					"/pgc.service.card.app.AppCard/MyFollows",
					"/pgc.service.card.app.AppCard/LocationLimit",
				},
			},
		},
		"ctr.predictor": {
			"ctr.predictor": {
				"Divin": {
					"/ctr.predictor.Divin/evaluate",
				},
			},
		},
		"main.silverbullet.silverbullet-proxy-service": {
			"silverbullet.service.silverbulletproxy": {
				"SilverbulletProxy": {
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/IsAllowedToDo",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/RiskInfo",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/LoginRiskInfo",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/Register",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/Validate",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/SendSms",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/CheckSms",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/SendMail",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/CheckMail",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/SendSmsWeb",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/CheckSmsWeb",
					"/silverbullet.service.silverbulletproxy.SilverbulletProxy/HideTel",
				},
			},
		},
		"pgc.service.ott": {
			"pgc.service.ott": {
				"Ott": {
					"/pgc.service.ott.Ott/View",
					"/pgc.service.ott.Ott/RecommendList",
					"/pgc.service.ott.Ott/RankIndex",
					"/pgc.service.ott.Ott/RankList",
					"/pgc.service.ott.Ott/AddFollow",
					"/pgc.service.ott.Ott/DeleteFollow",
					"/pgc.service.ott.Ott/MyFollow",
					"/pgc.service.ott.Ott/WildRecommendList",
					"/pgc.service.ott.Ott/SeriesSeason",
				},
			},
		},
		"inf.taishan.proxy": {
			"taishan.api": {
				"TaishanProxy": {
					"/taishan.api.TaishanProxy/put",
					"/taishan.api.TaishanProxy/get",
					"/taishan.api.TaishanProxy/scan",
					"/taishan.api.TaishanProxy/del",
					"/taishan.api.TaishanProxy/batch_get",
					"/taishan.api.TaishanProxy/batch_put",
					"/taishan.api.TaishanProxy/batch_del",
				},
			},
		},
		"studio.service.growup": {
			"growup.service.v1": {
				"Incentive": {
					"/growup.service.v1.Incentive/AccInfo",
					"/growup.service.v1.Incentive/AccCreate",
					"/growup.service.v1.Incentive/TradeCreate",
					"/growup.service.v1.Incentive/TradeDetail",
					"/growup.service.v1.Incentive/TradeRefund",
					"/growup.service.v1.Incentive/TradePay",
					"/growup.service.v1.Incentive/TradeRefundInfo",
					"/growup.service.v1.Incentive/FundTransfer",
				},
			},
		},
		"vipinfo.bck.service": {
			"vipinfo.bck.service.v1": {
				"VipInfo": {
					"/vipinfo.bck.service.v1.VipInfo/Info",
					"/vipinfo.bck.service.v1.VipInfo/Infos",
					"/vipinfo.bck.service.v1.VipInfo/Ping",
				},
			},
		},
		"account.service.usersuit": {
			"account.service.usersuite.v1": {
				"Usersuit": {
					"/account.service.usersuite.v1.Usersuit/InviteCountStat",
					"/account.service.usersuite.v1.Usersuit/Buy",
					"/account.service.usersuite.v1.Usersuit/Apply",
					"/account.service.usersuite.v1.Usersuit/InviteCode",
					"/account.service.usersuite.v1.Usersuit/Stat",
					"/account.service.usersuite.v1.Usersuit/Equip",
					"/account.service.usersuite.v1.Usersuit/GrantByMids",
					"/account.service.usersuite.v1.Usersuit/GroupPendantMid",
					"/account.service.usersuite.v1.Usersuit/PointFlag",
					"/account.service.usersuite.v1.Usersuit/MedalHomeInfo",
					"/account.service.usersuite.v1.Usersuit/MedalUserInfo",
					"/account.service.usersuite.v1.Usersuit/MedalInstall",
					"/account.service.usersuite.v1.Usersuit/MedalPopup",
					"/account.service.usersuite.v1.Usersuit/MedalMyInfo",
					"/account.service.usersuite.v1.Usersuit/MedalAllInfo",
					"/account.service.usersuite.v1.Usersuit/MedalGrant",
					"/account.service.usersuite.v1.Usersuit/MedalActivated",
					"/account.service.usersuite.v1.Usersuit/MedalActivatedMulti",
					"/account.service.usersuite.v1.Usersuit/Equipment",
					"/account.service.usersuite.v1.Usersuit/Equipments",
					"/account.service.usersuite.v1.Usersuit/Items",
					"/account.service.usersuite.v1.Usersuit/Groups",
					"/account.service.usersuite.v1.Usersuit/UserAchievementFeed",
					"/account.service.usersuite.v1.Usersuit/UserAchievementGroups",
					"/account.service.usersuite.v1.Usersuit/UserAchievementGroupDetail",
					"/account.service.usersuite.v1.Usersuit/ActiveAchievement",
					"/account.service.usersuite.v1.Usersuit/UserPreference",
					"/account.service.usersuite.v1.Usersuit/IsAchievementPublic",
					"/account.service.usersuite.v1.Usersuit/UpdateAchievementPublic",
				},
			},
		},
		"main.account.account-control-plane": {
			"account.service.account_control_plane.v1": {
				"AccountControlPlane": {
					"/account.service.account_control_plane.v1.AccountControlPlane/ListEffectivedControlUserStatus",
					"/account.service.account_control_plane.v1.AccountControlPlane/HasControlRole",
					"/account.service.account_control_plane.v1.AccountControlPlane/IsAllowedToDo",
					"/account.service.account_control_plane.v1.AccountControlPlane/ReleaseByAction",
					"/account.service.account_control_plane.v1.AccountControlPlane/ReleaseRole",
					"/account.service.account_control_plane.v1.AccountControlPlane/AddControlRole",
					"/account.service.account_control_plane.v1.AccountControlPlane/UpdateControlAnnotation",
					"/account.service.account_control_plane.v1.AccountControlPlane/ControlRole",
					"/account.service.account_control_plane.v1.AccountControlPlane/ListControlRole",
					"/account.service.account_control_plane.v1.AccountControlPlane/ListControlReleaseAction",
					"/account.service.account_control_plane.v1.AccountControlPlane/ControlReleaseAction",
					"/account.service.account_control_plane.v1.AccountControlPlane/ControlRoleStatus",
				},
			},
		},
		"crm.service.upaward": {
			"crm.service.award.v1": {
				"Tip": {
					"/crm.service.award.v1.Tip/DelPosTip",
					"/crm.service.award.v1.Tip/SavePosTip",
					"/crm.service.award.v1.Tip/GetUpTip",
					"/crm.service.award.v1.Tip/SaveUpTip",
					"/crm.service.award.v1.Tip/BatchSaveUpTip",
				},
			},
		},
		"video.vod.playurltvproj": {
			"video.vod.playurltvproj": {
				"PlayurlService": {
					"/video.vod.playurltvproj.PlayurlService/ProtobufPlayurl",
				},
			},
		},
		"cheese.service.dynamic": {
			"cheese.service.dynamic.v1": {
				"Dynamic": {
					"/cheese.service.dynamic.v1.Dynamic/MyPaid",
				},
			},
		},
		"crm.service.audit": {
			"crm.service.audit.v1": {
				"Audit": {
					"/crm.service.audit.v1.Audit/Ping",
					"/crm.service.audit.v1.Audit/AuditStat",
					"/crm.service.audit.v1.Audit/AuditStats",
					"/crm.service.audit.v1.Audit/AuditInfo",
					"/crm.service.audit.v1.Audit/AuditInfos",
					"/crm.service.audit.v1.Audit/AuditList",
					"/crm.service.audit.v1.Audit/Operate",
				},
			},
		},
		"live.xuser": {
			"live.xuser.v1": {
				"Guard": {
					"/live.xuser.v1.Guard/GetByUIDForGift",
					"/live.xuser.v1.Guard/GetPeakByTargetidUidsV2",
					"/live.xuser.v1.Guard/GetTopListGuardAttr",
				},
			},
		},
		"pgc.gateway.activity": {
			"pgc.gateway.activity.v1": {
				"Activity": {
					"/pgc.gateway.activity.v1.Activity/ShareUrl",
				},
			},
		},
		"ugcpay.service.rank": {
			"ugcpay.service.rank.v1": {
				"UGCPayRank": {
					"/ugcpay.service.rank.v1.UGCPayRank/RankElecAllAV",
					"/ugcpay.service.rank.v1.UGCPayRank/RankElecMonthAV",
					"/ugcpay.service.rank.v1.UGCPayRank/RankElecMonthUP",
					"/ugcpay.service.rank.v1.UGCPayRank/RankElecMonth",
					"/ugcpay.service.rank.v1.UGCPayRank/RankElecUpdateOrder",
					"/ugcpay.service.rank.v1.UGCPayRank/RankElecUpdateMessage",
					"/ugcpay.service.rank.v1.UGCPayRank/RankWithPanelByAV",
					"/ugcpay.service.rank.v1.UGCPayRank/UPRankWithPanelByUPMid",
					"/ugcpay.service.rank.v1.UGCPayRank/BNJRankWithPanel",
					"/ugcpay.service.rank.v1.UGCPayRank/ArchiveElecStatus",
				},
			},
		},
		"live.xgift": {
			"live.xgift.v1": {
				"Gift": {
					"/live.xgift.v1.Gift/smsReward",
				},
			},
		},
		"video.vod.playurlhtml5": {
			"video.vod.playurlhtml5": {
				"PlayurlService": {
					"/video.vod.playurlhtml5.PlayurlService/ProtobufPlayurl",
				},
			},
		},
		"cheese.service.settle": {
			"cheese.settle.service.v1": {
				"Settle": {
					"/cheese.settle.service.v1.Settle/CheckRerun",
					"/cheese.settle.service.v1.Settle/SettleRerun",
				},
			},
		},
		"archive.honor.service": {
			"archive.honor.service.v1": {
				"ArchiveHonor": {
					"/archive.honor.service.v1.ArchiveHonor/Honor",
					"/archive.honor.service.v1.ArchiveHonor/Honors",
				},
			},
		},
		"app.wall": {
			"app.wall.v1": {
				"AppWall": {
					"/app.wall.v1.AppWall/UnicomBindInfo",
				},
			},
		},
		"community.service.toview": {
			"community.service.toview": {
				"ToViews": {
					"/community.service.toview.ToViews/AddToView",
					"/community.service.toview.ToViews/AddToViews",
					"/community.service.toview.ToViews/DelToViews",
					"/community.service.toview.ToViews/DelToViewType",
					"/community.service.toview.ToViews/ClearToView",
					"/community.service.toview.ToViews/UserToViews",
					"/community.service.toview.ToViews/AllToViews",
				},
			},
		},
		"main.archive.creative": {
			"main.archive.creative": {
				"Creative": {
					"/main.archive.creative.Creative/FlowJudge",
					"/main.archive.creative.Creative/TaskTarget",
					"/main.archive.creative.Creative/ArchiveArgument",
					"/main.archive.creative.Creative/AchievementTarget",
					"/main.archive.creative.Creative/UpdateWaterMarkByJob",
					"/main.archive.creative.Creative/Ping",
					"/main.archive.creative.Creative/Close",
					"/main.archive.creative.Creative/CreditBalance",
					"/main.archive.creative.Creative/CreditDeduct",
					"/main.archive.creative.Creative/CreditRefund",
					"/main.archive.creative.Creative/CreditTransaction",
				},
			},
		},
		" sycpb.cpm.bce-brand-webapp": {
			"bcg.brand.ad.api": {
				"Brand": {
					"/bcg.brand.ad.api.Brand/AdSearch",
				},
			},
		},
		"sycpb.cpm.cpm-bce-slb": {
			"bcg.sunspot.test.api": {
				"Sunspot": {
					"/bcg.sunspot.test.api.Sunspot/AdSearch",
				},
			},
		},
		"passport.service.auth": {
			"passport.service.auth.v1": {
				"Auth": {
					"/passport.service.auth.v1.Auth/GetCookie",
					"/passport.service.auth.v1.Auth/GetToken",
					"/passport.service.auth.v1.Auth/GetMainToken",
					"/passport.service.auth.v1.Auth/GetTmpToken",
					"/passport.service.auth.v1.Auth/GetRefresh",
					"/passport.service.auth.v1.Auth/DelCookieCache",
					"/passport.service.auth.v1.Auth/DelTokenCache",
					"/passport.service.auth.v1.Auth/BatchGetCookies",
					"/passport.service.auth.v1.Auth/GetTokenV2",
				},
			},
		},
		"main.common-arch.sequence": {
			"sequence.v1": {
				"Seq": {
					"/sequence.v1.Seq/AutoIncrement",
					"/sequence.v1.Seq/SnowFlake",
				},
			},
		},
		"pgc.service.activity": {
			"pgc.service.activity.v1": {
				"Activity": {
					"/pgc.service.activity.v1.Activity/Share",
					"/pgc.service.activity.v1.Activity/ActivityInfo",
					"/pgc.service.activity.v1.Activity/ShareDetail",
					"/pgc.service.activity.v1.Activity/LatestShare",
					"/pgc.service.activity.v1.Activity/ShareAuth",
					"/pgc.service.activity.v1.Activity/Participate",
				},
			},
		},
		"garb.service": {
			"garb.service.v1": {
				"Garb": {
					"/garb.service.v1.Garb/UserEquip",
					"/garb.service.v1.Garb/UserEquipMulti",
					"/garb.service.v1.Garb/UserAsset",
					"/garb.service.v1.Garb/UserAssetList",
					"/garb.service.v1.Garb/GrantByAdmin",
					"/garb.service.v1.Garb/GrantByExpire",
					"/garb.service.v1.Garb/GrantByBiz",
					"/garb.service.v1.Garb/GrantPendantByOID",
					"/garb.service.v1.Garb/UserLoadEquip",
					"/garb.service.v1.Garb/UserLoadEquipByOID",
					"/garb.service.v1.Garb/UserUnloadEquip",
					"/garb.service.v1.Garb/UserUpdateAsset",
					"/garb.service.v1.Garb/UserDeleteAsset",
					"/garb.service.v1.Garb/UserDeleteAssetByAdmin",
					"/garb.service.v1.Garb/UserEmojiAccessInfo",
					"/garb.service.v1.Garb/UserAssetHistory",
					"/garb.service.v1.Garb/UserFanInfo",
					"/garb.service.v1.Garb/UserFanInfoList",
					"/garb.service.v1.Garb/UserFanRoll",
					"/garb.service.v1.Garb/UserWallet",
					"/garb.service.v1.Garb/UserSetting",
					"/garb.service.v1.Garb/SetUserSetting",
					"/garb.service.v1.Garb/SpaceBGUserEquip",
					"/garb.service.v1.Garb/SpaceBGUserAssetList",
					"/garb.service.v1.Garb/SpaceBGUserAssetListWithFan",
					"/garb.service.v1.Garb/SpaceBG",
					"/garb.service.v1.Garb/PendantEquipMulti",
					"/garb.service.v1.Garb/CardBGEquipMulti",
					"/garb.service.v1.Garb/SailingEquipMulti",
					"/garb.service.v1.Garb/TradeCreate",
					"/garb.service.v1.Garb/TradeQuery",
					"/garb.service.v1.Garb/MallGroupList",
					"/garb.service.v1.Garb/MallItem",
					"/garb.service.v1.Garb/MallList",
					"/garb.service.v1.Garb/MallSuit",
					"/garb.service.v1.Garb/SkinUserEquip",
					"/garb.service.v1.Garb/SkinList",
					"/garb.service.v1.Garb/SkinColorUserList",
					"/garb.service.v1.Garb/Items",
					"/garb.service.v1.Garb/FanRank",
					"/garb.service.v1.Garb/FanRecentRank",
					"/garb.service.v1.Garb/HandleVIPExpired",
					"/garb.service.v1.Garb/HandleItemInvalid",
					"/garb.service.v1.Garb/LoadingUserEquip",
				},
			},
		},
		"playurl.service": {
			"playurl.service.v2": {
				"PlayURL": {
					"/playurl.service.v2.PlayURL/PlayURL",
				},
			},
		},
		"community.service.coin": {
			"community.service.coin.v1": {
				"Coin": {
					"/community.service.coin.v1.Coin/AddCoin",
					"/community.service.coin.v1.Coin/ItemUserCoins",
					"/community.service.coin.v1.Coin/ItemsUserCoins",
					"/community.service.coin.v1.Coin/UserCoins",
					"/community.service.coin.v1.Coin/ModifyCoins",
					"/community.service.coin.v1.Coin/List",
					"/community.service.coin.v1.Coin/CoinsLog",
					"/community.service.coin.v1.Coin/AddUserCoinExp",
					"/community.service.coin.v1.Coin/UpdateAddCoin",
					"/community.service.coin.v1.Coin/TodayExp",
					"/community.service.coin.v1.Coin/SettleDetail",
					"/community.service.coin.v1.Coin/UpMemberState",
				},
			},
		},
		"main.community.tag": {
			"main.community.tag.v1": {
				"TagRPC": {
					"/main.community.tag.v1.TagRPC/Tag",
					"/main.community.tag.v1.TagRPC/TagByName",
					"/main.community.tag.v1.TagRPC/Tags",
					"/main.community.tag.v1.TagRPC/TagByNames",
					"/main.community.tag.v1.TagRPC/AddSub",
					"/main.community.tag.v1.TagRPC/CancelSub",
					"/main.community.tag.v1.TagRPC/UpdateCustomSort",
					"/main.community.tag.v1.TagRPC/UpBind",
					"/main.community.tag.v1.TagRPC/AdminBind",
					"/main.community.tag.v1.TagRPC/ArcTags",
					"/main.community.tag.v1.TagRPC/SubTags",
					"/main.community.tag.v1.TagRPC/CustomSortChannel",
					"/main.community.tag.v1.TagRPC/ResTags",
					"/main.community.tag.v1.TagRPC/ResTag",
					"/main.community.tag.v1.TagRPC/Channel",
					"/main.community.tag.v1.TagRPC/ChannelCategory",
					"/main.community.tag.v1.TagRPC/ChanneList",
					"/main.community.tag.v1.TagRPC/ChannelRecommend",
					"/main.community.tag.v1.TagRPC/ChannelDiscovery",
					"/main.community.tag.v1.TagRPC/ChannelSquare",
					"/main.community.tag.v1.TagRPC/ChannelResources",
					"/main.community.tag.v1.TagRPC/ChannelCheckBack",
					"/main.community.tag.v1.TagRPC/ChannelPartitionResources",
				},
			},
		},
		"community.service.dm": {
			"model": {
				"DM": {
					"/model.DM/SubjectInfos",
					"/model.DM/EditDMState",
					"/model.DM/EditDMPool",
					"/model.DM/EditDMAttr",
					"/model.DM/BuyAdvance",
					"/model.DM/AdvanceState",
					"/model.DM/Advances",
					"/model.DM/PassAdvance",
					"/model.DM/DenyAdvance",
					"/model.DM/CancelAdvance",
					"/model.DM/AddUserFilters",
					"/model.DM/UserFilters",
					"/model.DM/DelUserFilters",
					"/model.DM/AddUpFilters",
					"/model.DM/UpFilters",
					"/model.DM/BanUsers",
					"/model.DM/CancelBanUsers",
					"/model.DM/EditUpFilters",
					"/model.DM/AddGlobalFilter",
					"/model.DM/GlobalFilters",
					"/model.DM/DelGlobalFilters",
					"/model.DM/Mask",
					"/model.DM/SubtitleGet",
					"/model.DM/SubtitleSujectSubmit",
					"/model.DM/SubtitleSubjectSubmitGet",
					"/model.DM/DmDetail",
				},
			},
		},
		"cpm.divin": {
			"cpm.divin": {
				"Divin": {
					"/cpm.divin.Divin/evaluate",
				},
			},
		},
		"tv.interface": {
			"": {
				"TVInterface": {
					"/TVInterface/VideoAuthUgc",
					"/TVInterface/OgvAuth",
					"/TVInterface/PugvAuth",
					"/TVInterface/OgvMarkPlay",
				},
			},
		},
		"video.vod.playurlugcbatch": {
			"video.vod.playurlugcbatch": {
				"PlayurlService": {
					"/video.vod.playurlugcbatch.PlayurlService/ProtobufPlayurl",
				},
			},
		},
		"live.daoanchor": {
			"live.daoanchor.v1": {
				"DaoAnchor": {
					"/live.daoanchor.v1.DaoAnchor/FetchRoomByIDs",
					"/live.daoanchor.v1.DaoAnchor/RoomOnlineListByArea",
					"/live.daoanchor.v1.DaoAnchor/RoomOnlineInfoByArea",
					"/live.daoanchor.v1.DaoAnchor/RoomOnlineListFromDB",
					"/live.daoanchor.v1.DaoAnchor/RoomCreate",
					"/live.daoanchor.v1.DaoAnchor/RoomUpdate",
					"/live.daoanchor.v1.DaoAnchor/RoomBatchUpdate",
					"/live.daoanchor.v1.DaoAnchor/RoomExtendUpdate",
					"/live.daoanchor.v1.DaoAnchor/RoomExtendBatchUpdate",
					"/live.daoanchor.v1.DaoAnchor/RoomExtendIncre",
					"/live.daoanchor.v1.DaoAnchor/RoomExtendBatchIncre",
					"/live.daoanchor.v1.DaoAnchor/RoomTagCreate",
					"/live.daoanchor.v1.DaoAnchor/RoomAttrCreate",
					"/live.daoanchor.v1.DaoAnchor/RoomAttrSetEx",
					"/live.daoanchor.v1.DaoAnchor/AnchorUpdate",
					"/live.daoanchor.v1.DaoAnchor/AnchorBatchUpdate",
					"/live.daoanchor.v1.DaoAnchor/AnchorIncre",
					"/live.daoanchor.v1.DaoAnchor/AnchorBatchIncre",
					"/live.daoanchor.v1.DaoAnchor/FetchAreas",
					"/live.daoanchor.v1.DaoAnchor/FetchAttrByIDs",
					"/live.daoanchor.v1.DaoAnchor/DeleteTagByID",
					"/live.daoanchor.v1.DaoAnchor/PendantCreate",
					"/live.daoanchor.v1.DaoAnchor/PendantAddToRoom",
					"/live.daoanchor.v1.DaoAnchor/GetTagListByRoomId",
					"/live.daoanchor.v1.DaoAnchor/QueryAreaInfo",
					"/live.daoanchor.v1.DaoAnchor/AddPendantToRoom",
				},
				"Popularity": {
					"/live.daoanchor.v1.Popularity/GetPopularityByRoomIds",
				},
			},
		},
		"antispam.service": {
			"antispam.service.v1": {
				"Antispam": {
					"/antispam.service.v1.Antispam/CheckAction",
					"/antispam.service.v1.Antispam/TextSimilarity",
					"/antispam.service.v1.Antispam/TextCategorize",
					"/antispam.service.v1.Antispam/TextBadCaseFeedBack",
				},
			},
		},
		"ott.service": {
			"ott.service.v1": {
				"OTTService": {
					"/ott.service.v1.OTTService/ArcsAllow",
				},
			},
		},
		"live.xroom": {
			"live.xroom.v1": {
				"Room": {
					"/live.xroom.v1.Room/getMultiple",
					"/live.xroom.v1.Room/getMultipleByUids",
					"/live.xroom.v1.Room/isAnchor",
					"/live.xroom.v1.Room/getAreaInfo",
					"/live.xroom.v1.Room/getPendantByRoomIds",
					"/live.xroom.v1.Room/getStatusInfoByUids",
					"/live.xroom.v1.Room/getRoomPlayInfo",
					"/live.xroom.v1.Room/getLivKeyByRoomId",
					"/live.xroom.v1.Room/getRecordTranscodeConf",
					"/live.xroom.v1.Room/getAllLivingRoomsInfo",
				},
			},
		},
		"main.community.emote-service": {
			"main.community.emote_service.v1": {
				"EmoteService": {
					"/main.community.emote_service.v1.EmoteService/Packages",
					"/main.community.emote_service.v1.EmoteService/ListUserHiddenPackage",
					"/main.community.emote_service.v1.EmoteService/ListOpPackage",
					"/main.community.emote_service.v1.EmoteService/ListUserPanelPackage",
					"/main.community.emote_service.v1.EmoteService/ListBusinessPackage",
					"/main.community.emote_service.v1.EmoteService/SortPackage",
					"/main.community.emote_service.v1.EmoteService/AddPackage",
					"/main.community.emote_service.v1.EmoteService/AddGarbPackage",
					"/main.community.emote_service.v1.EmoteService/RemovePackage",
					"/main.community.emote_service.v1.EmoteService/EmoteByText",
					"/main.community.emote_service.v1.EmoteService/EmoteByPackage",
					"/main.community.emote_service.v1.EmoteService/BadgeStatus",
					"/main.community.emote_service.v1.EmoteService/UpdateBadgeStatus",
					"/main.community.emote_service.v1.EmoteService/EmoteOwnership",
					"/main.community.emote_service.v1.EmoteService/PackageOwnership",
				},
			},
		},
		"main.frontend.riskcontrol": {
			"frontend.riskcontrol.v1": {
				"RiskManagement": {
					"/frontend.riskcontrol.v1.RiskManagement/analyze",
				},
			},
		},
		"main.dynamic.feed": {
			"dynamic.service.feed.v1": {
				"Feed": {
					"/dynamic.service.feed.v1.Feed/UpdateNum",
				},
			},
		},
		"bbq.user.user-interface": {
			"bbq.user.interface.v1": {
				"User": {
					"/bbq.user.interface.v1.User/GetUserJumpBBQGrant",
					"/bbq.user.interface.v1.User/GetAccountDestroyLevel",
				},
			},
		},
		"crm.service.skyeye": {
			"crm.service.skyeye.v1": {
				"Skyeye": {
					"/crm.service.skyeye.v1.Skyeye/Ping",
					"/crm.service.skyeye.v1.Skyeye/AddEvent",
					"/crm.service.skyeye.v1.Skyeye/ReportEvent",
				},
			},
		},
	}
)

func GetByAppID(appid string) (map[string]map[string][]string, bool) {
	serviceMd, ok := grpcMd[appid]
	if !ok {
		return nil, false
	}
	return serviceMd, true
}

func GetByAppIDs(appids []string) (map[string]map[string]map[string][]string, bool) {
	ret := map[string]map[string]map[string][]string{}
	for _, appid := range appids {
		serviceMd, ok := grpcMd[appid]
		if ok {
			ret[appid] = serviceMd
		}
	}
	if len(ret) == 0 {
		return nil, false
	}
	return ret, true
}

func PrefixMatch(prefix string) (map[string]map[string]map[string][]string, bool) {
	res := map[string]map[string]map[string][]string{}
	for key := range grpcMd {
		if strings.HasPrefix(key, prefix) {
			res[key] = grpcMd[key]
		}
	}
	if len(res) == 0 {
		return nil, false
	}
	return res, true
}

func FuzzyMatch(candidate string) (map[string]map[string]map[string][]string, bool, error) {
	res := map[string]map[string]map[string][]string{}
	r, err := regexp.Compile(candidate)
	if err != nil {
		return nil, false, err
	}
	for key := range grpcMd {
		if r.Match([]byte(key)) {
			res[key] = grpcMd[key]
		}
	}
	if len(res) == 0 {
		return nil, false, nil
	}
	return res, true, nil
}
