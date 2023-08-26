package web

import (
	"fmt"
	"unicode"

	"go-gateway/app/app-svr/archive/service/api"

	"go-gateway/pkg/idsafe/bvid"
)

type OutArc struct {
	ID        int64 `json:"id"`
	Aid       int64 `json:"aid"`
	SnapView  int64 `json:"snap_view"`
	IsDeleted int   `json:"is_deleted"`
}

type BaiduSitemap struct {
	Sitemapindex []*BaiduSitemapItem `json:"sitemapindex"`
}

type BaiduSitemapItem struct {
	Sitemap *BaiduSiteMapDetail `json:"sitemap"`
}

type BaiduSiteMapDetail struct {
	Loc     string `json:"loc"`
	Lastmod string `json:"lastmod"`
}

type PushArc struct {
	// 标题
	Title string `json:"title"`
	// 简介
	Body string `json:"body"`
	// 小程序内页链接
	Path string `json:"path"`
	// 封面
	Images []string `json:"images"`
	// 资源类型 1000
	MappType string `json:"mapp_type"`
	// 组员子类型 1002
	MappSubType string `json:"mapp_sub_type"`
	// 内容一级分类
	FeedType string `json:"feed_type"`
	// 内容二级分类
	FeedSubType string     `json:"feed_sub_type"`
	Ext         *PubArcExt `json:"ext"`
}

type PubArcExt struct {
	// 发布时间 2018 年 9 月 8 日
	PublishTime string `json:"publish_time"`
	// 视频时长 小时需要换算成分钟 61:20
	VideoDuration string `json:"video_duration"`
	PcURL         string `json:"pc_url"`
	H5URL         string `json:"h5_url"`
	RawType       string `json:"raw_type"`
	RawSubType    string `json:"raw_sub_type"`
}

// nolint:gomnd
func (out *PushArc) CopyFromArc(in *api.Arc, arcType map[int32]*api.Tp) {
	out.Title = in.Title
	out.Body = in.Desc
	if out.Body == "" {
		out.Body = out.Title
	}
	bvidStr, _ := bvid.AvToBv(in.Aid)
	out.Path = pushPath(bvidStr)
	out.Images = []string{in.Pic}
	out.MappType = "1000"
	out.MappSubType = "1002"
	out.FeedType = ""
	out.FeedSubType = ""
	var pTypeName string
	if typ, ok := arcType[in.TypeID]; ok && typ != nil {
		if pType, ok := arcType[typ.Pid]; ok && pType != nil {
			pTypeName = pType.Name
		}
	}
	out.Ext = &PubArcExt{
		PublishTime:   in.PubDate.Time().Format("2006年1月2日"),
		VideoDuration: fmt.Sprintf("%02d:%02d", in.Duration/60, in.Duration-(in.Duration/60*60)),
		PcURL:         fmt.Sprintf("https://www.bilibili.com/video/%s", bvidStr),
		H5URL:         fmt.Sprintf("https://m.bilibili.com/video/%s", bvidStr),
		RawType:       pTypeName,
		RawSubType:    in.TypeName,
	}
}

func (out *PushArc) ForbidArc() bool {
	// 长度6到40
	if titleLen := len([]rune(out.Title)); titleLen < 6 || titleLen > 40 {
		return true
	}
	// 必须有汉字
	for _, v := range out.Title {
		if unicode.Is(unicode.Han, v) {
			return false
		}
	}
	return true
}

type PushDelArc struct {
	Path string `json:"path"`
}

func (out *PushDelArc) FmtFromAid(aid int64) {
	bvidStr, _ := bvid.AvToBv(aid)
	out.Path = pushPath(bvidStr)
}

func pushPath(bvid string) string {
	return fmt.Sprintf("/pages/video/video?bvid=%s", bvid)
}

type XiaoMiArg struct {
	AppID     int64  `json:"appId"`
	SecretKey string `json:"secretKey"`
	Timestamp int64  `json:"timestamp"`
	MsgID     string `json:"msgId"`
	Msg       string `json:"msg"`
}

type XiaoMiMsg struct {
	MsgType int            `json:"msgType"`
	MsgData *XiaomiMsgData `json:"msgData"`
}

type XiaomiMsgData struct {
	Total    int64            `json:"total"`
	Articles []*XiaomiArticle `json:"articles"`
}

type XiaomiArticle struct {
	ID string `json:"id"`
}
