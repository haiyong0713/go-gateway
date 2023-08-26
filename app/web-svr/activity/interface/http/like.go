package http

import (
	"encoding/base64"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	xbinding "go-common/library/net/http/blademaster/binding"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/model/lol"
	"go-gateway/app/web-svr/activity/interface/service"
	"go-gateway/pkg/idsafe/bvid"
)

func guessPredictList(c *bm.Context) {
	var (
		err   error
		total int
		list  []*lol.ContestDetail
	)
	v := new(struct {
		Pn int `form:"pn" validate:"min=1" default:"1"`
		Ps int `form:"ps" validate:"min=1" default:"10"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)

	if list, total, err = service.LolSvc.UserPredictListV2(c, mid, v.Pn, v.Ps); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func v2GuessList(c *bm.Context) {
	arg := new(struct {
		Mid      int64 `form:"mid"`
		Oid      int64 `form:"oid"`
		Business int64 `form:"business"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(service.LikeSvc.GuessListV2(c, arg.Mid, arg.Oid, arg.Business))
}

func subject(c *bm.Context) {
	params := c.Request.Form
	sidStr := params.Get("sid")
	sid, err := strconv.ParseInt(sidStr, 10, 32)
	if err != nil {
		log.Error("strconv.ParseInt(%s) error(%v)", sidStr, err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(service.LikeSvc.Subject(c, sid))
}

func subjectInfo(c *bm.Context) {
	var (
		err error
		sub *like.SubjectItem
	)
	arg := new(struct {
		Sid  int64  `form:"sid" validate:"min=1"`
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
	})
	if err = c.Bind(arg); err != nil {
		return
	}
	if arg.Bvid != "" {
		arg.Aid, _ = bvid.BvToAv(arg.Bvid)
	}
	if arg.Aid > 0 {
		c.JSON(service.LikeSvc.ActSubjectWithAid(c, arg.Sid, arg.Aid))
		return
	} else {
		if sub, err = service.LikeSvc.ActSubject(c, arg.Sid); err == nil {
			c.JSON(sub.PublicData(), nil)
			return
		}
	}
	c.JSON(nil, err)
}

func subjectProtos(c *bm.Context) {
	arg := new(struct {
		Sids []int64 `form:"sids,split" validate:"min=1,max=50,dive,min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(service.LikeSvc.Protocols(c, arg.Sids))
}

func vote(c *bm.Context) {
	var (
		mid int64
	)
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	params := c.Request.Form
	voteStr := params.Get("vote")
	vote, err := strconv.ParseInt(voteStr, 10, 64)
	if err != nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	stageStr := params.Get("stage")
	stage, err := strconv.ParseInt(stageStr, 10, 64)
	if err != nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	aidStr := params.Get("aid")
	aid, err := strconv.ParseInt(aidStr, 10, 64)
	if err != nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if strRe, _ := service.LikeSvc.OnlineVote(c, mid, vote, stage, aid); !strRe {
		c.JSON(nil, xecode.NotModified)
		return
	}
	c.JSON("ok", nil)
}

func ltime(c *bm.Context) {
	params := c.Request.Form
	sidStr := params.Get("sid")
	sid, err := strconv.ParseInt(sidStr, 10, 64)
	if err != nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	index, err := service.LikeSvc.Ltime(c, sid)
	if err != nil {
		log.Error("error(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if index == nil {
		c.JSON(nil, xecode.NothingFound)
		return
	}
	c.JSON(index, nil)
}

func likeAct(c *bm.Context) {
	p := new(like.ParamAddLikeAct)
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	request := c.Request
	p.UA = request.UserAgent()
	p.Referer = request.Referer()
	p.IP = metadata.String(c, metadata.RemoteIP)
	if p.Buvid == "" {
		buvid := request.Header.Get(_headerBuvid)
		if buvid == "" {
			cookie, _ := request.Cookie(_buvid)
			if cookie != nil {
				buvid = cookie.Value
			}
		}
		p.Buvid = buvid
	}
	if p.Origin == "" {
		p.Origin = request.Header.Get("Origin")
	}
	p.API = c.Request.URL.Path
	if res, err := service.LikeSvc.LikeAct(c, p, mid); err == nil && res != nil {
		c.JSON(res.ActID, nil)
	} else {
		c.JSON(nil, err)
	}
}

func likeActBySidCVId(c *bm.Context) {
	p := new(like.ParamAddLikeActWithSidCVId)
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	if err := service.LikeSvc.LikeActBySidCVId(c, p, mid); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

func batchLikeAct(c *bm.Context) {
	p := new(struct {
		Sid  int64   `form:"sid" validate:"min=1"`
		Lids []int64 `form:"lids,split" validate:"min=1,max=6,dive,min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.BatchLikeAct(c, mid, p.Sid, p.Lids))
}

func likeActLikes(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.LikeActLikes(c, p.Sid, mid))
}

func storyKingAct(c *bm.Context) {
	p := new(like.ParamStoryKingAct)
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	request := c.Request
	p.UA = request.UserAgent()
	p.Referer = request.Referer()
	p.IP = metadata.String(c, metadata.RemoteIP)
	if p.Buvid == "" {
		buvid := request.Header.Get(_headerBuvid)
		if buvid == "" {
			cookie, _ := request.Cookie(_buvid)
			if cookie != nil {
				buvid = cookie.Value
			}
		}
		p.Buvid = buvid
	}
	if p.Origin == "" {
		p.Origin = request.Header.Get("Origin")
	}
	c.JSON(service.LikeSvc.StoryKingAct(c, p, mid))
}

func upAddTimes(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.UpAddVoteTime(c, p.Sid, mid))
}

func upVoteAddTimes(c *bm.Context) {
	p := new(struct {
		Sid   int64 `form:"sid" validate:"min=1"`
		Lid   int64 `form:"lid" validate:"min=1"`
		IsAdd int64 `form:"is_add"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.UpVoteAppendTimes(c, p.Sid, p.Lid, p.IsAdd))
}

func upAddTime(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.UpAppendTime(c, p.Sid, p.Mid))
}

func storyKingLeft(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.StoryKingLeftTime(c, p.Sid, mid))
}

func upList(c *bm.Context) {
	var mid int64
	p := new(like.ParamList)
	if err := c.Bind(p); err != nil {
		return
	}
	if p.Sid != conf.Conf.Taaf.Sid && p.Sid != conf.Conf.Timemachine.FlagSid {
		if p.Ps > 100 {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.UpList(c, p, mid))
}

func likeactUpList(c *bm.Context) {
	var mid int64
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.LikeActUpList(c, p.Sid, mid))
}

func upListRelation(c *bm.Context) {
	var mid int64
	p := new(struct {
		Sid int64 `form:"sid"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.UpListRelation(c, p.Sid, mid))
}

func likeSlider(c *bm.Context) {
	v := new(struct {
		Sid  int64   `form:"sid" validate:"min=1"`
		Lids []int64 `form:"lids,split" validate:"min=1,max=50,dive,min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	mid := int64(0)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.Slider(c, v.Lids, v.Sid, mid))
}

func likeOne(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	mid := int64(0)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.OneItem(c, v.Lid, v.Sid, mid))
}

func missionLike(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.MissionLike(c, p.Sid, mid))
}

func missionLikeAct(c *bm.Context) {
	p := new(like.ParamMissionLikeAct)
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.MissionLikeAct(c, p, mid))
}

func missionInfo(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.MissionInfo(c, p.Sid, p.Lid, mid))

}
func missionTops(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Num int   `form:"num" validate:"min=1,max=200"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(service.LikeSvc.MissionTops(c, p.Sid, p.Num))
}

func missionUser(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(service.LikeSvc.MissionUser(c, p.Sid, p.Lid))
}

func missionRank(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.MissionRank(c, p.Sid, mid))
}

func missionFriends(c *bm.Context) {
	p := new(like.ParamMissionFriends)
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.MissionFriendsList(c, p, mid))
}

func missionAward(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.MissionAward(c, p.Sid, mid))
}

func missionAchieve(c *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		ID  int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.MissionAchieve(c, p.Sid, p.ID, mid))
}

func likeActList(c *bm.Context) {
	v := new(struct {
		Sid  int64   `form:"sid" validate:"min=1"`
		Lids []int64 `form:"lids,split" validate:"min=1,max=50,dive,min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	mid := int64(0)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.LikeActList(c, v.Sid, mid, v.Lids))
}

func likeDel(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.LikeDel(c, v.Sid, v.Lid, mid))
}

func likeMyList(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Ps  int   `form:"ps" default:"15" validate:"min=1,max=50"`
		Pn  int   `form:"pn" default:"1" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.LikeMyList(c, v.Sid, mid, v.Ps, v.Pn))
}

func subjectInit(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.SubjectInitialize(c, v.Sid-1))
}

func likeInit(c *bm.Context) {
	v := new(struct {
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.LikeInitialize(c, v.Lid-1))
}

func subjectLikeListInit(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.SubjectLikeListInitialize(c, v.Sid))
}

func likeActCountInit(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.LikeActCountInitialize(c, v.Sid))
}

func tagList(c *bm.Context) {
	var (
		err  error
		cnt  int
		list []*like.Like
	)
	v := new(struct {
		Sid   int64  `form:"sid" validate:"min=1"`
		TagID int64  `form:"tag_id" validate:"min=1"`
		Type  string `form:"type" default:"ctime"`
		Pn    int    `form:"pn" default:"1" validate:"min=1"`
		Ps    int    `form:"ps" default:"30" validate:"min=1,max=30"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Type != "ctime" && v.Type != "random" {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if list, cnt, err = service.LikeSvc.TagArcList(c, v.Sid, v.TagID, v.Pn, v.Ps, v.Type, metadata.String(c, metadata.RemoteIP)); err != nil {
		c.JSON(nil, err)
		return
	}
	data := map[string]interface{}{
		"page": map[string]int{
			"num":   v.Pn,
			"size":  v.Ps,
			"total": cnt,
		},
		"list": list,
	}
	c.JSON(data, nil)
}

func regionList(c *bm.Context) {
	var (
		err  error
		cnt  int
		list []*like.Like
	)
	v := new(struct {
		Sid  int64  `form:"sid" validate:"min=1"`
		Rid  int32  `form:"rid" validate:"min=1"`
		Type string `form:"type" default:"ctime"`
		Pn   int    `form:"pn" default:"1" validate:"min=1"`
		Ps   int    `form:"ps" default:"30" validate:"min=1,max=30"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Type != "ctime" && v.Type != "random" {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if list, cnt, err = service.LikeSvc.RegionArcList(c, v.Sid, v.Rid, v.Pn, v.Ps, v.Type, metadata.String(c, metadata.RemoteIP)); err != nil {
		c.JSON(nil, err)
		return
	}
	data := map[string]interface{}{
		"page": map[string]int{
			"num":   v.Pn,
			"size":  v.Ps,
			"total": cnt,
		},
		"list": list,
	}
	c.JSON(data, nil)
}

func tagStats(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.TagLikeCounts(c, v.Sid))
}

func subjectStat(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.SubjectStat(c, v.Sid))
}

func setSubjectStat(c *bm.Context) {
	v := new(like.SubjectStat)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.SetSubjectStat(c, v))
}

func viewRank(c *bm.Context) {
	v := new(struct {
		Sid  int64  `form:"sid" validate:"min=1"`
		Pn   int    `form:"pn" default:"1" validate:"min=1"`
		Ps   int    `form:"ps" default:"20" validate:"min=1"`
		Type string `form:"type"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, count, err := service.LikeSvc.ViewRank(c, v.Sid, v.Pn, v.Ps, v.Type)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	data["list"] = list
	data["page"] = map[string]int{
		"pn":    v.Pn,
		"ps":    v.Ps,
		"count": count,
	}
	c.JSON(data, err)
}

func setViewRank(c *bm.Context) {
	v := new(struct {
		Sid  int64   `form:"sid" validate:"min=1"`
		Aids []int64 `form:"aids,split" validate:"min=1,dive,min=1"`
		Type string  `form:"type"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.SetViewRank(c, v.Sid, v.Aids, v.Type))
}

func groupData(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	ck := c.Request.Header.Get("cookie")
	c.JSON(service.LikeSvc.ObjectGroup(c, v.Sid, ck))
}

func upListGroup(c *bm.Context) {
	var mid int64
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(service.LikeSvc.UpListGroup(c, v.Sid, mid))
}

func upListHis(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.LikeHisList(c, v.Sid))
}

func setLikeContent(c *bm.Context) {
	v := new(struct {
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.SetLikeContent(c, v.Lid))
}

func addLikeAct(c *bm.Context) {
	v := new(struct {
		Sid   int64 `form:"sid" validate:"min=1"`
		Lid   int64 `form:"lid" validate:"min=1"`
		Score int64 `form:"score"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if v.Score == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, service.LikeSvc.AddLikeActCache(c, v.Sid, v.Lid, v.Score))
}

func likeActCache(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.LikeActCache(c, v.Sid, v.Lid))
}

func likeActState(c *bm.Context) {
	v := new(struct {
		Sid  int64   `form:"sid" validate:"min=1"`
		Mid  int64   `form:"mid" validate:"min=1"`
		Lids []int64 `form:"lids,split" validate:"required,min=1,max=50,dive,gt=0"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.LikeActState(c, v.Sid, v.Mid, v.Lids))
}

func likeOidsInfo(c *bm.Context) {
	v := new(struct {
		Type int     `form:"type" validate:"min=1"`
		Oids []int64 `form:"oids,split" validate:"required,min=1,max=50,dive,gt=0"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.LikeOidsInfo(c, v.Type, v.Oids))
}

func likeAddOther(c *bm.Context) {
	var err error
	arg := new(like.ParamOther)
	if err = c.BindWith(arg, xbinding.Form); err != nil {
		return
	}
	img := c.Request.Form.Get("img")
	if img != "" {
		strpos := strings.Index(img, ",")
		imgStr := img[strpos+1:]
		if arg.Image, err = base64.StdEncoding.DecodeString(imgStr); err != nil || len(arg.Image) == 0 {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	} else {
		var imageFile multipart.File
		imageFile, _, err = c.Request.FormFile("file")
		if err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
		defer imageFile.Close()
		if arg.Image, err = ioutil.ReadAll(imageFile); err != nil || len(arg.Image) == 0 {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	arg.FileType = http.DetectContentType(arg.Image)
	if len(arg.Image) > 5120*1024 {
		c.JSON(nil, ecode.ActivityBodyTooLarge)
		return
	}
	if arg.FileType != "image/jpeg" && arg.FileType != "image/jpg" && arg.FileType != "image/png" {
		c.JSON(nil, ecode.ActivityFileTypeFail)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.LikeAddOther(c, arg, mid))
}

func likeAddText(c *bm.Context) {
	arg := new(like.ParamText)
	if err := c.Bind(arg); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.LikeAddText(c, arg, mid))
}

func likeTotal(c *bm.Context) {
	arg := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	if total, err := service.LikeSvc.LikeTotal(c, arg.Sid); err != nil {
		c.JSON(nil, err)
	} else {
		c.JSON(map[string]interface{}{"total": total}, nil)
	}
}

func subInfo(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.ActSubject(c, v.Sid))
}

func subInfos(c *bm.Context) {
	v := new(struct {
		Sids []int64 `form:"sids,split" validate:"min=1,max=20,dive,min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.ActSubjects(c, v.Sids))
}

func subjectUp(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.SubjectUp(c, v.Sid))
}

func likeUp(c *bm.Context) {
	v := new(struct {
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.LikeUp(c, v.Lid))
}

func actSetReload(c *bm.Context) {
	v := new(struct {
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.ActSetReload(c, v.Lid))
}

func likeCtimeCache(c *bm.Context) {
	v := new(struct {
		Lid int64 `form:"lid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.AddLikeCtimeCache(c, v.Lid))
}

func delLikeCtimeCache(c *bm.Context) {
	v := new(struct {
		Lid      int64 `form:"lid" validate:"min=1"`
		Sid      int64 `form:"sid" validate:"min=1"`
		LikeType int64 `form:"like_type"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.DelLikeCtimeCache(c, v.Lid, v.Sid, v.LikeType))
}

func likeCheckJoin(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	join, err := service.LikeSvc.LikeCheckJoin(c, mid, v.Sid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"join": join}, nil)
}

func actListInfo(c *bm.Context) {
	var loginMid int64
	v := new(struct {
		Type     int   `form:"type" validate:"required"`
		Platform int   `form:"platform"`
		Mid      int64 `form:"mid"`
		Region   int   `form:"region"`
		AppType  int64 `form:"app_type"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
		if loginMid != 0 {
			v.Mid = loginMid
		}
	}
	c.JSON(service.LikeSvc.ActListInfo(c, v.Mid, v.Type, v.Platform, v.Region, v.AppType))
}

func addUpListHis(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, service.LikeSvc.AddLikeHisList(c, v.Sid))
}

func likeActToken(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	token, err := service.LikeSvc.LikeActToken(c, v.Sid, mid)
	if err != nil || token == nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(map[string]interface{}{"token": token.Token}, nil)
}

func inviteTimes(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.InviteTimes(c, v.Sid, mid))
}

func festivalProcess(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.FestivalProcess(c, mid))
}

func yellowGreenVote(ctx *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1" default:"15439"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.YellowGreenVote(ctx, p.Sid))
}

func actFilter(ctx *bm.Context) {
	arg := new(like.ParamFilter)
	if err := ctx.Bind(arg); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.ActFilter(ctx, arg))
}

func viewData(ctx *bm.Context) {
	var (
		total int
		list  []*like.WebDataRes
		err   error
	)
	arg := new(like.ParamViewData)
	if err = ctx.Bind(arg); err != nil {
		return
	}
	if list, total, err = service.LikeSvc.ViewData(ctx, arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   arg.Pn,
		"size":  arg.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	ctx.JSON(data, nil)
}

func memoryData(ctx *bm.Context) {
	p := new(struct {
		Type int64 `form:"type" validate:"min=1"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.WatchData(ctx, p.Type))
}

func cacheData(ctx *bm.Context) {
	p := new(like.CacheData)
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.CacheData(ctx, p))
}

func addData(ctx *bm.Context) {
	p := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.AddData(ctx, p.Sid))
}

func upActReserveAudit(ctx *bm.Context) {
	p := new(struct {
		Sid   int64 `form:"sid" validate:"min=1"`
		State int64 `form:"state" validate:"min=1"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.UpActReserveAuditCallBack(ctx, p.Sid, p.State))
}

func tagConvert(ctx *bm.Context) {
	p := new(struct {
		TagIDs string `form:"tag_ids" validate:"required"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.GetTagConvert(ctx, p.TagIDs))
}
