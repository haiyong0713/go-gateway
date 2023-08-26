package duertv

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-car/job/model"
	"go-gateway/app/app-svr/app-car/job/model/bangumi"
	"go-gateway/app/app-svr/app-car/job/model/region"
	"go-gateway/app/app-svr/archive/service/api"

	chanGRPC "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

const (
	_epRex       = `(EP|ep|Ep|eP)[0-9]+`
	_pgcOffline  = 2
	_pgcOffline2 = 3
)

func epkey(id int64) string {
	return fmt.Sprintf("ep%d", id)
}

func sskey(id int64) string {
	return fmt.Sprintf("ss%d", id)
}

func avkey(id int64) string {
	return fmt.Sprintf("av%d", id)
}

func ckey(id int64) string {
	return fmt.Sprintf("cv%d", id)
}

type DuertvPush struct {
	ID             string `json:"id"`
	Pid            string `json:"pid"`
	Provider       string `json:"provider"`
	Partner        string `json:"partner"`
	ResourceStatus int    `json:"resource_status"`
	Name           string `json:"name"`
	SerialName     string `json:"serial_name"`
	AliasName      string `json:"alias_name"`
	Type           string `json:"type"`
	Category       string `json:"category"`
	SourceType     string `json:"source_type"`
	Tag            string `json:"tag"`
	Duration       int64  `json:"duration"`
	Season         int    `json:"season"`
	TotalEpisodes  int    `json:"total_episodes"`
	Episode        int    `json:"episode"`
	Director       string `json:"director"`
	Actor          string `json:"actor"`
	Region         string `json:"region"`
	ReleaseDate    int    `json:"release_date"`
	UpdateTime     int    `json:"update_time"`
	Cost           string `json:"cost"`
	Hot            int    `json:"hot"`
	Weight         int    `json:"weight"`
	Language       string `json:"language"`
	Definition     string `json:"definition"`
	Introduction   string `json:"introduction"`
	PosterURL      string `json:"poster_url"`
	ThumbURL       string `json:"thumb_url"`
	Token          string `json:"token"`
	Extend         string `json:"extend"`
	DataType       string `json:"data_type"`
}

type DuertvPushUGC struct {
	ID                 string  `json:"id"`
	Pid                string  `json:"pid"`
	Partner            string  `json:"partner"`
	Name               string  `json:"name"`
	Subtitle           string  `json:"subtitle,omitempty"`
	ResourceStatus     int     `json:"resource_status"`
	Tag                string  `json:"tag"`
	FirstCategory      string  `json:"first_category"`
	SecondCategory     string  `json:"second_category"`
	CoverURL           string  `json:"cover_url"`
	PlayURL            string  `json:"play_url"`
	Duration           int64   `json:"duration"`
	AuthorID           string  `json:"author_id"`
	Author             string  `json:"author,omitempty"`
	AuthorLevel        string  `json:"author_level,omitempty"`
	Brief              string  `json:"brief,omitempty"`
	Episode            int     `json:"episode,omitempty"`
	TotalEpisode       int     `json:"total_episode,omitempty"`
	PreviewURLHttp     string  `json:"preview_url_http,omitempty"`
	PreviewURLHttps    string  `json:"preview_url_https,omitempty"`
	PreviewVedioHeight int     `json:"preview_video_height,omitempty"`
	PreviewVedioWidth  int     `json:"preview_video_width,omitempty"`
	VideoHeight        int     `json:"video_height,omitempty"`
	VideoWidth         int     `json:"video_width,omitempty"`
	VideoSize          int     `json:"video_size,omitempty"`
	VideoInfoExt       string  `json:"video_info_ext,omitempty"`
	VideoScore         float64 `json:"video_score,omitempty"`
	CommentCnt         int     `json:"comment_cnt,omitempty"`
	LikeCnt            int     `json:"like_cnt,omitempty"`
	CollectCnt         int     `json:"collect_cnt,omitempty"`
	ChildrenMode       string  `json:"children_mode,omitempty"`
	LastModifyTime     int64   `json:"last_modify_time,omitempty"`
	CreateTime         string  `json:"create_time,omitempty"`
	Cost               string  `json:"cost,omitempty"`
	CostInfo           string  `json:"cost_info,omitempty"`
	Extend             string  `json:"extend,omitempty"`
	Hot                int     `json:"hot,omitempty"`
}

func (p *DuertvPush) FromBangumi(e *bangumi.Episode, b *bangumi.Content, partner string) bool {
	p.ID = epkey(e.ID)
	p.Pid = sskey(b.Season.ID)
	p.Partner = partner
	// 操作类型 1-上架 2-下架 3-删除
	p.ResourceStatus = 1
	switch b.OpType {
	case _pgcOffline, _pgcOffline2:
		// 上下线标识，1：上线；-1下线
		p.ResourceStatus = -1
	}
	p.SerialName = b.Name
	if b.Name == "" {
		for _, st := range b.SeasonSeries {
			if st.Title != "" {
				p.SerialName = st.Title
				break
			}
		}
	}
	p.Name = e.IndexTitle
	// 如果当前index_title是1、2、3、3.5的，手动拼接成第xx集
	if _, err := strconv.ParseFloat(e.IndexTitle, 64); err == nil {
		p.Name = fmt.Sprintf("第%s集", e.IndexTitle)
	}
	if p.SerialName == "" {
		return false
	}
	p.Name = p.SerialName + "：" + p.Name
	// 逗号分割改成 / 分割
	p.AliasName = strings.Replace(b.Alias, ",", "/", -1)
	p.Type = b.ContentTypeString()
	var t []string
	for _, tag := range b.Tag {
		t = append(t, tag.Name)
	}
	p.Tag = strings.Join(t, "/")
	if e.Duration > 0 {
		p.Duration = e.Duration
	}
	p.Season = b.Season.Index
	if b.Season.TotalCount > 0 {
		p.TotalEpisodes = b.Season.TotalCount
	}
	p.Episode = e.Index
	if b.PubRealTime > 0 {
		pudata, _ := strconv.Atoi(xtime.Time(b.PubRealTime).Time().Format("2006"))
		p.ReleaseDate = pudata
	}
	if b.Mtime > 0 {
		pudata, _ := strconv.Atoi(xtime.Time(b.Mtime).Time().Format("20060102"))
		p.UpdateTime = pudata
	}
	p.Cost = e.EpisodeCost()
	p.Introduction = b.Intro
	p.PosterURL = b.CoverImage
	//p.EpURL = e.Cover //这里先保留不漏出，否则每张EP封面都漏出了
	p.ThumbURL = b.CoverImage
	p.Token = model.FillURI(b.Season.ID, e.ID, model.ParamHandler(model.SearchPrune(b.Season.ID, e.ID, model.GotoPGC), model.EntranceCommonSearch, b.Name))
	// extend拓展字段，里面放入json数据
	extend := map[string]interface{}{
		"uri": p.Token,
	}
	bExtend, _ := json.Marshal(extend)
	p.Extend = string(bExtend)
	p.DataType = "pgc"
	// 逗号分割改成 / 分割
	p.Region = strings.Replace(b.Country, ",", "/", -1)
	return true
}

func (p *DuertvPush) FromBangumiSeason(b *bangumi.Content, partner string) bool {
	const (
		_sec = 60
	)
	p.ID = sskey(b.Season.ID)
	p.Pid = sskey(b.Season.ID)
	p.Partner = partner
	// 操作类型 1-上架 2-下架 3-删除
	p.ResourceStatus = 1
	switch b.OpType {
	case _pgcOffline, _pgcOffline2:
		// 上下线标识，1：上线；-1下线
		p.ResourceStatus = -1
	}
	p.SerialName = b.Name
	if b.Name == "" {
		for _, st := range b.SeasonSeries {
			if st.Title != "" {
				p.SerialName = st.Title
				break
			}
		}
	}
	if p.SerialName == "" {
		return false
	}
	p.Name = p.SerialName
	// 逗号分割改成 / 分割
	p.AliasName = strings.Replace(b.Alias, ",", "/", -1)
	p.Type = b.ContentTypeString()
	var t []string
	for _, tag := range b.Tag {
		t = append(t, tag.Name)
	}
	p.Tag = strings.Join(t, "/")
	if b.Duration > 0 {
		// 专辑的播放时长 分钟数，需要转换成秒
		p.Duration = b.Duration * _sec
	}
	p.Season = b.Season.Index
	if b.Season.TotalCount > 0 {
		p.TotalEpisodes = b.Season.TotalCount
	}
	if b.PubRealTime > 0 {
		pudata, _ := strconv.Atoi(xtime.Time(b.PubRealTime).Time().Format("2006"))
		p.ReleaseDate = pudata
	}
	if b.Mtime > 0 {
		pudata, _ := strconv.Atoi(xtime.Time(b.Mtime).Time().Format("20060102"))
		p.UpdateTime = pudata
	}
	p.Cost = b.Season.SeasonCost()
	p.Introduction = b.Intro
	p.PosterURL = b.CoverImage
	epURL := "" //season的横图字段
	if len(b.Episodes) > 0 {
		epURL = b.Episodes[0].Cover
	}
	p.ThumbURL = b.CoverImage
	p.Token = model.FillURI(b.Season.ID, 0, model.ParamHandler(model.SearchPrune(b.Season.ID, 0, model.GotoPGC), model.EntranceCommonSearch, b.Name))
	// extend拓展字段，里面放入json数据
	extend := map[string]interface{}{
		"uri":    p.Token,
		"view":   b.PlayCount,
		"ep_url": epURL,
	}
	bExtend, _ := json.Marshal(extend)
	p.Extend = string(bExtend)
	p.DataType = "pgc"
	// 逗号分割改成 / 分割
	p.Region = strings.Replace(b.Country, ",", "/", -1)
	p.Hot = mathLog10(int(b.PlayCount))
	return true
}

func (p *DuertvPush) FromBangumiEPOffshelve(e *bangumi.OffshelveEpInfo, b *bangumi.Offshelve, partner string) bool {
	epid := fromEpID(e.PlayURL)
	if epid == "" {
		return false
	}
	p.ID = epid
	p.Pid = sskey(b.SeasonID)
	p.Partner = partner
	// 操作类型 1-上架 -1 -下架
	p.ResourceStatus = -1
	p.SerialName = b.Name
	p.Name = b.Name
	p.Type = b.OffshelveTypeString()
	p.Episode = e.Index
	pudata, _ := strconv.Atoi(time.Now().Format("2006"))
	p.ReleaseDate = pudata
	update, _ := strconv.Atoi(time.Now().Format("20060102"))
	p.UpdateTime = update
	p.DataType = "pgc"
	return true
}

func (p *DuertvPush) FromBangumiSeasonOffshelve(b *bangumi.Offshelve, partner string) bool {
	p.ID = sskey(b.SeasonID)
	p.Pid = sskey(b.SeasonID)
	p.Partner = partner
	// 操作类型 1-上架 -1 -下架
	p.ResourceStatus = -1
	p.SerialName = b.Name
	p.Name = b.Name
	p.Type = b.OffshelveTypeString()
	p.Episode = 1
	pudata, _ := strconv.Atoi(time.Now().Format("2006"))
	p.ReleaseDate = pudata
	update, _ := strconv.Atoi(time.Now().Format("20060102"))
	p.UpdateTime = update
	p.DataType = "pgc"
	return true
}

func fromEpID(url string) string {
	r := regexp.MustCompile(_epRex)
	fIndex := r.FindStringIndex(url)
	if len(fIndex) == 0 {
		return ""
	}
	return url[fIndex[0]:fIndex[1]]
}

func (p *DuertvPush) FromBangumiMessage(b *bangumi.DatabusEntity, partner string) bool {
	if b.EntityChange == nil || b.EntityChange.PayLoad == nil {
		return false
	}
	p.ID = fmt.Sprintf("ep%s", b.EntityChange.EntityID)
	p.Pid = fmt.Sprintf("ss%s", b.EntityID)
	p.Partner = partner
	// 操作类型 1-上架 -1 -下架
	p.ResourceStatus = -1
	p.Type = b.EntityChange.PayLoad.DataBusTypeString()
	p.Episode = 1
	pudata, _ := strconv.Atoi(time.Now().Format("2006"))
	p.ReleaseDate = pudata
	update, _ := strconv.Atoi(time.Now().Format("20060102"))
	p.UpdateTime = update
	p.DataType = "pgc"
	return true
}

func (p *DuertvPushUGC) FromUGC(a *api.Arc, partner string, chans []*chanGRPC.Channel, reg map[int32]*region.Region) bool {
	// 过滤
	if filterUGC(a) {
		return false
	}
	// 付费视频
	p.Cost = "免费"
	if a.AttrVal(api.AttrBitIsPUGVPay) == api.AttrYes {
		p.Cost = "付费"
	}
	p.ID = ckey(a.FirstCid)
	p.Pid = avkey(a.Aid)
	p.Partner = partner
	// 上下线标识，1：上线；-1下线
	p.ResourceStatus = 1
	// state >= 0 表示稿件可以被用户正常访问，其他情况都无法打开稿件页
	if a.State < 0 {
		p.ResourceStatus = -1
	}
	p.Name = a.Title
	// tag标签处理
	var tags []string
	for _, cl := range chans {
		tags = append(tags, cl.Name)
	}
	p.Tag = a.TypeName
	if len(tags) > 0 {
		p.Tag = strings.Join(tags, "/")
	}
	// 一级分区处理
	p.FirstCategory = a.TypeName
	if regInfo, ok := reg[a.TypeID]; ok {
		p.FirstCategory = regInfo.Name
	}
	p.SecondCategory = a.TypeName
	p.Duration = 1
	if a.Duration > 0 {
		p.Duration = a.Duration
	}
	p.Episode = 1
	p.TotalEpisode = 1
	if a.Videos > 0 {
		p.TotalEpisode = int(a.Videos)
	}
	p.Episode = 1
	p.LastModifyTime = a.PubDate.Time().Unix()
	p.CreateTime = strconv.FormatInt(a.PubDate.Time().Unix(), 10)
	p.Brief = a.Desc
	p.CoverURL = a.Pic
	p.PlayURL = model.FillURI(a.Aid, a.FirstCid, model.ParamHandler(model.SearchPrune(a.Aid, a.FirstCid, model.GotoAv), model.EntranceCommonSearch, a.Title))
	// extend拓展字段，里面放入json数据
	extend := map[string]interface{}{
		"uri": p.PlayURL,
		"uid": a.Author.Mid,
	}
	bExtend, _ := json.Marshal(extend)
	p.Extend = string(bExtend)
	return true
}

func (p *DuertvPushUGC) FromUGCCollection(a *api.Arc, partner string, chans []*chanGRPC.Channel, reg map[int32]*region.Region) bool {
	// 过滤
	if filterUGC(a) {
		return false
	}
	// 付费视频
	p.Cost = "免费"
	if a.AttrVal(api.AttrBitIsPUGVPay) == api.AttrYes {
		p.Cost = "付费"
	}
	p.ID = avkey(a.Aid)
	p.Pid = avkey(a.Aid)
	p.Partner = partner
	// 上下线标识，1：上线；-1下线
	p.ResourceStatus = 1
	// state >= 0 表示稿件可以被用户正常访问，其他情况都无法打开稿件页
	if a.State < 0 {
		p.ResourceStatus = -1
	}
	p.Name = a.Title
	// tag标签处理
	var tags []string
	for _, cl := range chans {
		tags = append(tags, cl.Name)
	}
	p.Tag = a.TypeName
	if len(tags) > 0 {
		p.Tag = strings.Join(tags, "/")
	}
	// 一级分区处理
	p.FirstCategory = a.TypeName
	if regInfo, ok := reg[a.TypeID]; ok {
		p.FirstCategory = regInfo.Name
	}
	p.SecondCategory = a.TypeName
	p.Duration = 1
	if a.Duration > 0 {
		p.Duration = a.Duration
	}
	p.Episode = 1
	p.TotalEpisode = 1
	if a.Videos > 0 {
		p.TotalEpisode = int(a.Videos)
	}
	p.LastModifyTime = a.PubDate.Time().Unix()
	p.CreateTime = strconv.FormatInt(a.PubDate.Time().Unix(), 10)
	p.Brief = a.Desc
	p.CoverURL = a.Pic
	p.PlayURL = model.FillURI(a.Aid, 0, model.ParamHandler(model.SearchPrune(a.Aid, a.FirstCid, model.GotoAv), model.EntranceCommonSearch, a.Title))
	// extend拓展字段，里面放入json数据
	extend := map[string]interface{}{
		"uri":   p.PlayURL,
		"view":  a.Stat.View,
		"reply": a.Stat.Reply,
		"fav":   a.Stat.Fav,
		"uid":   a.Author.Mid,
	}
	bExtend, _ := json.Marshal(extend)
	p.Extend = string(bExtend)
	p.Hot = mathLog10(int(a.Stat.View))
	return true
}

func mathLog10(number int) int {
	if number == 0 {
		return 0
	}
	return int(math.Log10(float64(number)) * 10)
}

func filterUGC(a *api.Arc) bool {
	// 过滤原因：互动视频
	return a.AttrVal(api.AttrBitSteinsGate) == api.AttrYes
}
