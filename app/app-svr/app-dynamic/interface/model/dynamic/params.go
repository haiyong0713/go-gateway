package dynamic

import "go-gateway/app/app-svr/app-dynamic/interface/api"

type Header struct {
	MobiApp  string `json:"mobi_app"`
	Device   string `json:"device"`
	Buvid    string `json:"buvid"`
	Platform string `json:"platform"`
	Build    int    `json:"build"`
	IP       string `json:"ip"`
}

type VideoMate struct {
	Qn        int `json:"qn"`
	Fnver     int `json:"fnver"`
	Fnval     int `json:"fnval"`
	ForceHost int `json:"force_host"`
	Fourk     int `json:"fourk"`
}

type DynVideoReq struct {
	Teenager       int        `json:"teenager"`
	UpdateBaseLine string     `json:"updateBaseLine"`
	Offset         string     `json:"offset"`
	Page           int        `json:"page"`
	Refresh        int        `json:"refresh"`
	VideoMate      *VideoMate `json:"video_mate"`
	Mid            int64      `json:"mid"`
}

type MaterialParams struct {
	Header    *Header
	VideoMate *VideoMate
	Teenager  int
	Mid       int64
	Dynamics  []*Dynamics
	UpList    *VdUpListRsp
}

type DynDetailsReq struct {
	Teenager  int        `json:"teenager"`
	DynIDs    []int64    `json:"DynIds"`
	Mid       int64      `json:"mid"`
	VideoMate *VideoMate `json:"video_mate"`
}

type EpMapV2Params struct {
	EpIDs    []int64
	MobiApp  string
	Platform string
	Device   string
	Build    int
	IP       string
	Fnver    int
	Fnval    int
}

type DynTopicsParams struct {
	DynIDs []int64 `json:"dynamic_ids"`
	Flexes string  `json:"flexes_json"`
}
type TopicParasParams struct {
	Paras []Paras `json:"paras"`
}
type TopicParasList struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Paras struct {
	List []TopicParasList `json:"list"`
	Type int              `json:"type"`
}

type LikeIconReq struct {
	Uid      int64      `json:"uid"`
	Platform int        `json:"platform"`
	Items    []*QryIcon `json:"items"`
}
type QryIcon struct {
	DynID      int64    `json:"dynamic_id"`
	TopicNames []string `json:"topic_names"`
}

type LikeBusiReq struct {
	Mids       []int64                    `json:"mids"`
	Businesses map[string][]*LikeBusiItem `json:"businesses"`
}
type LikeBusiItem struct {
	OrigID int64 `json:"origin_id"`
	MsgID  int64 `json:"message_id"`
}

type UserLikeItem struct {
	Mid  int64 `json:"mid"`
	Time int   `json:"time"`
}

type DynVideoPersonalReq struct {
	Teenager  int        `json:"teenager"`
	Offset    string     `json:"offset"`
	Page      int        `json:"page"`
	IsPreload int        `json:"is_preload"`
	VideoMate *VideoMate `json:"video_mate"`
	Mid       int64      `json:"mid"`
	HostUID   int64      `json:"host_uid"`
}

type DynUpdOffsetReq struct {
	HostUID    int64  `json:"host_uid"`
	ReadOffset string `json:"read_offset"`
	Mid        int64  `json:"mid"`
}

type DynBottomReq struct {
	Dynamics []*BottomReqItem `json:"dynamics"`
}

type BottomReqItem struct {
	DynId  int64   `json:"dynamic_id"`
	Bottom *Bottom `json:"bottom"`
}

type DynBriefsReq struct {
	Teenager int     `json:"teenagers_mode"`
	DynIDs   []int64 `json:"dynamic_ids"`
	Uid      int64   `json:"uid"`
}

func (v *VideoMate) FromSVideo(val *api.SVideoReq) {
	if val.PlayerPreload != nil {
		v.Qn = int(val.PlayerPreload.Qn)
		v.Fnver = int(val.PlayerPreload.Fnver)
		v.Fnval = int(val.PlayerPreload.Fnval)
		v.ForceHost = int(val.PlayerPreload.ForceHost)
		v.Fourk = int(val.PlayerPreload.Fourk)
		return
	}
	v.Qn = int(val.Qn)
	v.Fnver = int(val.Fnver)
	v.Fnval = int(val.Fnval)
	v.ForceHost = int(val.ForceHost)
	v.Fourk = int(val.Fourk)
}

func IsPopularSv(req *api.SVideoReq) bool {
	return req.Type == api.SVideoType_TypePopularIndex || req.Type == api.SVideoType_TypePopularHotword
}

func (h *Header) IsPad() bool {
	return h.MobiApp == "iphone" && h.Device == "pad"
}

func (h *Header) IsPadHD() bool {
	return h.MobiApp == "ipad"
}

func (h *Header) IsAndroidHD() bool {
	return h.MobiApp == "android_hd"
}

func (h *Header) IsAndroid() bool {
	return h.MobiApp == "android"
}

func (h *Header) IsPhone() bool {
	return h.MobiApp == "iphone" && h.Device == "phone"
}
