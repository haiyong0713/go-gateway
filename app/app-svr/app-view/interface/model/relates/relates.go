package relates

import (
	"encoding/json"

	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	viewModule "go-gateway/app/app-svr/app-view/interface/model/view"

	pageApi "git.bilibili.co/bapis/bapis-go/bilibili/pagination"
)

type RelatesFeedGRPCRequest struct {
	Aid         int64               `json:"aid"`          //稿件id
	Mid         int64               `json:"mid"`          //用户id
	Build       int64               `json:"build"`        //客户端版本号
	Buvid       string              `json:"buvid"`        //设备号id
	TrackId     string              `json:"track_id"`     //请求view接口时的trackid（标记用户的上一次请求）
	Plat        int8                `json:"plat"`         //平台
	MobileApp   string              `json:"mobile_app"`   //客户端类型
	Network     string              `json:"network"`      //网络
	Device      string              `json:"device"`       //设备
	DisableRcmd int                 `json:"disable_rcmd"` //关闭个性化推荐，1关闭
	PageIndex   int64               `json:"page_index"`   //相关推荐请求页数(新版本使用next参数分页)
	SessionId   string              `json:"session_id"`   //唯一标识一个播放详情页
	FromSpmid   string              `json:"from_spmid"`   //上级页面
	Spmid       string              `json:"spmid"`        //当前页
	From        string              `json:"from"`         //来源from
	Ip          string              `json:"ip"`           //用户ip
	Slocale     string              `json:"slocale"`
	Clocale     string              `json:"clocale"`
	Pagination  *pageApi.Pagination `json:"pagination"` //向下分页参数
	RefreshNum  int32               `json:"refresh_num"`
}

type AIRecommendResponse struct {
	Relates     []*viewModule.Relate     `json:"relates"`      //相关推荐结果
	RelateInfoc *viewModule.RelatesInfoc `json:"relate_infoc"` //相关推荐上报信息
	ReturnCode  string                   `json:"return_code"`  //AI返回code
	UserFeature string
	PvFeature   json.RawMessage
	PlayParam   int    `json:"play_param"` // 1=play automatically the relates, 0=not
	Next        string `json:"next"`       //分页参数
}

func FromRelates(in []*viewModule.Relate) (out []*viewApi.Relate) {
	for _, v := range in {
		if v == nil {
			continue
		}
		out = append(out, &viewApi.Relate{
			Aid:               v.Aid,
			Pic:               v.Pic,
			Title:             v.Title,
			Author:            v.Author,
			Stat:              &v.Stat,
			Duration:          v.Duration,
			Goto:              v.Goto,
			Param:             v.Param,
			Uri:               v.URI,
			JumpUrl:           v.JumpURL,
			Rating:            v.Rating,
			Reserve:           v.Reserve,
			From:              v.From,
			Desc:              v.Desc,
			RcmdReason:        v.RcmdReason,
			Badge:             v.Badge,
			Cid:               v.Cid,
			SeasonType:        v.SeasonType,
			RatingCount:       v.RatingCount,
			TagName:           v.TagName,
			PackInfo:          v.PackInfo,
			Notice:            v.Notice,
			Button:            v.Button,
			Trackid:           v.TrackID,
			NewCard:           int32(v.NewCard),
			RcmdReasonStyle:   v.ReasonStyle,
			CoverGif:          v.CoverGif,
			Cm:                v.CM,
			ReserveStatus:     v.ReserveStatus,
			RcmdReasonExtra:   v.RcmdReasonExtra,
			RecThreePoint:     v.RecThreePoint,
			UniqueId:          v.UniqueId,
			MaterialId:        v.MaterialId,
			FromSourceType:    v.FromSourceType,
			FromSourceId:      v.FromSourceId,
			BadgeStyle:        v.BadgeStyle,
			PowerIconStyle:    v.PowerIconStyle,
			ReserveStatusText: v.ReserveStatusText,
			Cover:             v.Cover,
		})
	}
	return
}
