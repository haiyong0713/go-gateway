package media

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	pgcreview "git.bilibili.co/bapis/bapis-go/pgc/service/review"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	pgcmedia "git.bilibili.co/bapis/bapis-go/pgc/servant/media"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"go-common/component/metadata/device"
	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	appcardca "go-gateway/app/app-svr/app-card/interface/model/card"
	protov2 "go-gateway/app/app-svr/app-card/interface/model/card/proto"
	appcardmeta "go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	archiveapi "go-gateway/app/app-svr/archive/service/api"
)

const (
	// biz type 媒资类型
	BizChannelType = 0
	BizMediaType   = 1
	//点赞资源
	LikeResource             = "https://i0.hdslb.com/bfs/archive/b9f49c9b33532c5d05f5ea701ecd063f81910e94.json"
	LikeResourceHash         = "c8b42c2a76890e703b15874175268b4b"
	DisLikeResource          = "https://i0.hdslb.com/bfs/archive/8aee6952487d118b4207c1afa2fd38616bd7545a.json"
	DisLikeResourceHash      = "bdbc35ebc88d178d1f409145dadec806"
	LikeNightResource        = "https://i0.hdslb.com/bfs/archive/3ed718f59e9e9cf1ce148105c9db9559951d5a7d.json"
	LikeNightResourceHash    = "bc9fecf2624a569c05cef8097e20eb37"
	DisLikeNightResource     = "https://i0.hdslb.com/bfs/archive/c9a20055b712068bfe293878639dc9066ba2690b.json"
	DisLikeNightResourceHash = "c370e8d031381f4716d7564956a8b182"
)

func SmallCardItem(in *archiveapi.Arc, dev device.Device) *SmallItem {
	return &SmallItem{
		Title:          in.GetTitle(),
		CoverImageUri:  in.GetPic(),
		CoverLeftIcon1: int64(appcardmdl.IconPlay),
		CoverLeftText1: appcardmdl.StatString(in.GetStat().View, ""),
		CoverLeftIcon2: int64(appcardmdl.IconComment),
		CoverLeftText2: appcardmdl.StatString(in.GetStat().Reply, ""),
		CoverRightText: appcardmdl.DurationString(in.GetDuration()),
		Param:          in.Aid,
		Mid:            in.GetAuthor().Mid,
		Uri: appcardmdl.FillURI(appcardmdl.GotoAv, dev.Plat(), int(dev.Build), strconv.FormatInt(in.GetAid(), 10),
			appcardmdl.ArcPlayHandler(in, nil, "", nil, int(dev.Build), dev.RawMobiApp, false)),
	}
}

func BigCardItem(in *archiveapi.Arc, ap *archiveapi.PlayerInfo, hasLike int32, dev device.Device, isFav bool) *BigItem {
	rly := &BigItem{
		CanPlay:        in.Rights.Autoplay,
		Title:          in.GetTitle(),
		CoverImageUri:  in.GetPic(),
		CoverLeftIcon1: int64(appcardmdl.IconPlay),
		CoverLeftText1: appcardmdl.StatString(in.GetStat().View, ""),
		CoverLeftIcon2: int64(appcardmdl.IconComment),
		CoverLeftText2: appcardmdl.StatString(in.GetStat().Reply, ""),
		CoverRightText: appcardmdl.DurationString(in.GetDuration()),
		IsFav:          isFav,
		UserCard: &UserCard{
			Mid:      in.GetAuthor().Mid,
			UserName: in.GetAuthor().Name,
			UserFace: in.GetAuthor().Face,
			UserUrl:  appcardmdl.FillURI(appcardmdl.GotoMid, 0, 0, strconv.FormatInt(in.GetAuthor().Mid, 10), nil),
		},
		Param: in.Aid,
		LikeButton: &LikeButton{
			Aid:                  in.Aid,
			Count:                in.GetStat().Like,
			ShowCount:            true,
			Selected:             hasLike,
			Event:                string(appcardmdl.EventlikeClick),
			EventV2:              string(appcardmdl.EventV2ButtonClick),
			LikeResource:         constructLikeButtonResource(LikeResource, LikeResourceHash),
			DisLikeResource:      constructLikeButtonResource(DisLikeResource, DisLikeResourceHash),
			LikeNightResource:    constructLikeButtonResource(LikeNightResource, LikeNightResourceHash),
			DisLikeNightResource: constructLikeButtonResource(DisLikeNightResource, DisLikeNightResourceHash),
		},
		Uri: appcardmdl.FillURI(appcardmdl.GotoAv, dev.Plat(), int(dev.Build), strconv.FormatInt(in.GetAid(), 10),
			appcardmdl.ArcPlayHandler(in, ap, "", nil, int(dev.Build), dev.RawMobiApp, true)),
	}
	rly.ThreePointMeta = appcardmeta.ConstructPanelMeta("traffic.movie-channel-detail-video.inline.three-point.click", "")
	rly.SharePlane = appcardca.ConstructSharePlane(in)
	rly.InlineProgressBar = appcardca.GetInlineProgressBar()
	rly.PlayerArgs = &protov2.PlayerArgs{Aid: in.Aid, Cid: in.FirstCid, Duration: in.Duration}
	return rly
}

func constructLikeButtonResource(url, hash string) *LikeButtonResource {
	return &LikeButtonResource{
		Url:  url,
		Hash: hash,
	}
}

func DetailItem(mediaIn *pgcmedia.MediaBizInfoGetReply) *MediaDetailReply {
	out := &MediaDetailReply{}
	if mediaIn.GetMediaPeopleInfo() != nil {
		castTmp := &Cast{Title: "演职人员"}
		//制作人员
		for _, v := range mediaIn.GetMediaPeopleInfo().GetCrew() {
			realName := v.GetCnName()
			if realName == "" {
				realName = v.GetRealName()
			}
			squareUrl := v.GetSquareUrl()
			if squareUrl == "" {
				squareUrl = v.GetAvatarUrl()
			}
			if squareUrl == "" && realName == "" && v.GetOccupation() == "" {
				continue
			}
			castTmp.Person = append(castTmp.Person, &MediaPerson{
				RealName:  realName,
				SquareUrl: squareUrl,
				Character: v.GetOccupation(),
				PersonId:  v.GetPersonId(),
				Type:      "crew",
			})
		}
		//角色影人
		for _, v := range mediaIn.GetMediaPeopleInfo().GetPlayers() {
			realName := v.GetCnName()
			if realName == "" {
				realName = v.GetRealName()
			}
			squareUrl := v.GetSquareUrl()
			if squareUrl == "" {
				squareUrl = v.GetAvatarUrl()
			}
			if squareUrl == "" && realName == "" && v.GetRoleName() == "" {
				continue
			}
			castTmp.Person = append(castTmp.Person, &MediaPerson{
				RealName:  realName,
				SquareUrl: squareUrl,
				Character: fmt.Sprintf("饰 %s", v.GetRoleName()),
				PersonId:  v.GetPersonId(),
				Type:      "player",
			})
		}
		//演职人员
		if len(castTmp.Person) > 0 {
			out.Cast = castTmp
		}
		//制作信息
		if len(mediaIn.GetMediaPeopleInfo().GetStaff()) > 0 {
			out.Staff = &Staff{Title: "制作信息", Text: mediaIn.GetMediaPeopleInfo().GetStaff()}
		}
	}
	//剧情简介
	if mediaIn.GetOverview() != "" {
		out.Overview = &Overview{Title: "剧情简介", Text: mediaIn.GetOverview()}
	}
	return out
}

// nolint:gomnd
func CardItem(mediaIn *pgcmedia.MediaBizInfoGetReply, seasonInfo *seasongrpc.CardInfoProto, isLike bool, tagName string, reviewInfo *pgcreview.ReviewInfoReply, isAllowReply, articleId int32) *MediaCard {
	if mediaIn == nil {
		return nil
	}
	out := &MediaCard{
		Cover:    mediaIn.CoverPhoto_3VS4Url,
		CurTitle: mediaIn.CurTitle,
	}
	styles := make([]string, 0)
	//频道
	categoryDesc := mediaIn.CategoryDesc
	labelEnd := "开播"
	switch mediaIn.CategoryId {
	case 1, 4:
		categoryDesc = "动画"
	case 2: //电影：上映
		labelEnd = "上映"
	default:
	}
	out.ButSecond = &Supernatant{
		Title: "点评",
		Item:  CommentCard(mediaIn, tagName, isAllowReply, articleId),
	}
	if categoryDesc != "" {
		styles = append(styles, categoryDesc)
	}
	if len(mediaIn.Areas) > 0 {
		styles = append(styles, strings.Join(mediaIn.Areas, " "))
	}
	if len(mediaIn.Styles) > 0 {
		styles = append(styles, strings.Join(mediaIn.Styles, " "))
	}
	if len(styles) > 0 {
		out.Style = strings.Join(styles, " | ")
	}
	label := make([]string, 0)
	if mediaIn.FirstReleaseDate != 0 {
		label = append(label, fmt.Sprintf("%s%s", time.Unix(mediaIn.FirstReleaseDate, 0).Format("2006年01月02日"), labelEnd))
	}
	if mediaIn.Runtime > 0 {
		label = append(label, fmt.Sprintf("%d分钟", mediaIn.Runtime))
	}
	if len(label) > 0 {
		out.Label = strings.Join(label, " | ")
	}
	//是否支持立即观看
	if seasonInfo != nil {
		out.ButFirst = &Button{Title: "立即观看", Link: seasonInfo.GetUrl(), Id: fmt.Sprintf("%d", seasonInfo.GetSeasonId()), ButType: ButType_BUT_REDIRECT}
	} else { //想看
		out.ButFirst = &Button{Title: "想看", HasTitle: "已想看", Id: fmt.Sprintf("%d", mediaIn.MediaBizId), ButType: ButType_BUT_LIKE, Icon: int64(appcardmdl.IconFavorite)}
		if isLike {
			out.ButFirst.FollowState = 1
		}
	}
	//评分>0&&允许点评
	if reviewInfo.GetScore() > 0 && isAllowReply == 1 {
		out.Scores = &Scores{Score: reviewInfo.GetScore()}
	}
	return out
}

func ShowTabs(in []*channelgrpc.ShowTab, isAllowReply int32) []*ShowTab {
	var out []*ShowTab
	for _, v := range in {
		var tabType TabType
		switch v.TabType {
		case channelgrpc.ShowTabType_SHOW_TAB_FEED_BIG:
			tabType = TabType_TAB_FEED_BID
		case channelgrpc.ShowTabType_SHOW_TAB_FEED_SMALL:
			tabType = TabType_TAB_FEED_SMALL
		case channelgrpc.ShowTabType_SHOW_TAB_MOVIE_DETAIL:
			tabType = TabType_TAB_OGV_DETAIL
		case channelgrpc.ShowTabType_SHOW_TAB_MOVIE_REPLY:
			if isAllowReply != 1 {
				continue
			}
			tabType = TabType_TAB_OGV_REPLY
		default:
			continue
		}
		out = append(out, &ShowTab{
			TabType: tabType,
			Title:   v.Title,
			Url:     v.Url,
		})
	}
	return out
}

func CommentCard(mediaIn *pgcmedia.MediaBizInfoGetReply, tagName string, isAllowReply, articleId int32) []*CommentItem {
	var (
		tid, stickerID int
		tags           = tagName
	)
	switch mediaIn.CategoryId {
	case 1, 4:
		tid = 27
	case 2: //电影：
		tid = 182
		stickerID = 1830868
		tags = fmt.Sprintf("%s,%s", tagName, "视频影评")
	case 3, 5, 7: //电视剧、综艺、纪录片
	default:
		return nil
	}
	url := fmt.Sprintf("bilibili://uper/center_plus?tab_index=1&is_pop_home=0&relation_from=movie&post_config={\"first_entrance\":\"电影\"}&tid=%d&sticker_id_v2=%d&tags=%s", tid, stickerID, tags)
	articleURL := fmt.Sprintf("https://member.bilibili.com/article-text/mobile?media_id=%d", mediaIn.GetMediaBizId())
	if articleId > 0 {
		articleURL = fmt.Sprintf("bilibili://article/%d", articleId)
	}
	mediaID := strconv.FormatInt(mediaIn.GetMediaBizId(), 10)
	rly := make([]*CommentItem, 0)
	if isAllowReply == 1 { //是否开始写长评，短评论
		rly = append(rly,
			&CommentItem{ActionType: "short-evaluate", Type: CommentType_comment_type_judge, Title: "写短评", Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220602/82ac2611e49c304c91fb79cc76b9b762/iLIKVaHBlQ.png", Id: mediaID, Url: fmt.Sprintf("activity://bangumi/review/short-review-publish?MEDIA_ID=%d", mediaIn.GetMediaBizId())},
			&CommentItem{ActionType: "long-evaluate", Type: CommentType_comment_type_judge, Title: "写长评", Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220602/82ac2611e49c304c91fb79cc76b9b762/6xbMi6OJ8r.png", Id: mediaID, Url: articleURL},
		)
	}
	rly = append(rly, &CommentItem{ActionType: "shoot-evaluate", Type: CommentType_comment_type_redirect, Title: "拍视频", Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220602/82ac2611e49c304c91fb79cc76b9b762/pwc5XISOgd.png", Url: url})
	return rly
}
