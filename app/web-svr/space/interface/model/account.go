package model

import (
	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	livemdl "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	pangugsgrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	seriesgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/series"
	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
)

// NavNum nav num struct.
type NavNum struct {
	Video     int64 `json:"video"`
	Bangumi   int32 `json:"bangumi"`
	Cinema    int32 `json:"cinema"`
	Channel   *Num  `json:"channel"`
	Favourite *Num  `json:"favourite"`
	Tag       int   `json:"tag"`
	Article   int32 `json:"article"`
	Playlist  int32 `json:"playlist"`
	Album     int64 `json:"album"`
	Audio     int   `json:"audio"`
	Pugv      int64 `json:"pugv"`
	SeasonNum int64 `json:"season_num"`
}

// Num num struct.
type Num struct {
	Master int `json:"master"`
	Guest  int `json:"guest"`
}

// UpStat up stat struct.
type UpStat struct {
	Archive struct {
		View int64 `json:"view"`
	} `json:"archive"`
	Article struct {
		View int64 `json:"view"`
	} `json:"article"`
	Likes int64 `json:"likes"`
}

type School struct {
	Name string `json:"name"`
}

type Profession struct {
	Name       string `json:"name"`
	Department string `json:"department"`
	Title      string `json:"title"`
	IsShow     int32  `json:"is_show"`
}

// AccInfo account info.
type AccInfo struct {
	Mid            int64                     `json:"mid"`
	Name           string                    `json:"name"`
	Sex            string                    `json:"sex"`
	Face           string                    `json:"face"`
	FaceNft        int32                     `json:"face_nft"`
	FaceNftType    pangugsgrpc.NFTRegionType `json:"face_nft_type"`
	Sign           string                    `json:"sign"`
	Rank           int32                     `json:"rank"`
	Level          int32                     `json:"level"`
	JoinTime       int32                     `json:"jointime"`
	Moral          int32                     `json:"moral"`
	Silence        int32                     `json:"silence"`
	Coins          float64                   `json:"coins"`
	FansBadge      bool                      `json:"fans_badge"`
	FansMedal      *FansMedal                `json:"fans_medal"`
	Official       accmdl.OfficialInfo       `json:"official"`
	Vip            accmdl.VipInfo            `json:"vip"`
	Pendant        accmdl.PendantInfo        `json:"pendant"`
	Nameplate      accmdl.NameplateInfo      `json:"nameplate"`
	UserHonourInfo *accmdl.UserHonourInfo    `json:"user_honour_info"`
	IsFollowed     bool                      `json:"is_followed"`
	TopPhoto       string                    `json:"top_photo"`
	Theme          interface{}               `json:"theme"`
	SysNotice      interface{}               `json:"sys_notice"`
	LiveRoom       *Live                     `json:"live_room"`
	Birthday       string                    `json:"birthday"`
	School         *School                   `json:"school"`
	Profession     *Profession               `json:"profession"`
	Tags           []string                  `json:"tags"`
	Series         *Series                   `json:"series"`
	IsSeniorMember int32                     `json:"is_senior_member"`
	LiveMCNInfo    *MCNInfo                  `json:"mcn_info"`
	GaiaResType    GaiaResponseType          `json:"gaia_res_type"`
	GaiaData       *gaiamdl.RuleCheckReply   `json:"gaia_data"`
	IsRisk         bool                      `json:"is_risk"`
	ElecInfo       *ElecPlusInfo             `json:"elec"` // 充电信息
}

type ElecPlusInfo struct {
	ShowInfo *ShowInfo `json:"show_info"` // 开关
}

type ShowInfo struct {
	Show    bool   `json:"show"`
	State   int8   `json:"state"`
	Title   string `json:"title"`
	Icon    string `json:"icon"`
	JumpURL string `json:"jump_url"`
}

type MCNInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Series struct {
	UserUpgradeStatus seriesgrpc.UserUpgradeStatus `json:"user_upgrade_status"`
	ShowUpgradeWindow bool                         `json:"show_upgrade_window"`
}

type FansMedal struct {
	Show  bool               `json:"show"`
	Wear  bool               `json:"wear"`
	Medal *livemdl.MedalInfo `json:"medal"`
}

// AccBlock acc block
type AccBlock struct {
	Status     int `json:"status"`
	IsDue      int `json:"is_due"`
	IsAnswered int `json:"is_answered"`
}

// TopPhoto top photo struct.
type TopPhoto struct {
	SImg         string `json:"s_img"`
	LImg         string `json:"l_img"`
	AndroidImg   string `json:"android_img"`
	IphoneImg    string `json:"iphone_img"`
	IpadImg      string `json:"ipad_img"`
	ThumbnailImg string `json:"thumbnail_img"`
	Sid          int64  `json:"sid"`
}

// Relation .
type Relation struct {
	Relation   interface{} `json:"relation"`
	BeRelation interface{} `json:"be_relation"`
}

type VisitAct struct {
	LoginMid int64  `json:"login_mid"`
	Mid      int64  `json:"mid"`
	Referer  string `json:"referer"`
	Buvid    string `json:"buvid"`
	Path     string `json:"path"`
	Ts       int64  `json:"ts"`
}

func filterHonorTag(in []*accmdl.HonourTag) []*accmdl.HonourTag {
	out := make([]*accmdl.HonourTag, 0, len(in))
	for _, t := range in {
		scene := sets.NewString(t.Scene...)
		if scene.Has("space") {
			out = append(out, t)
		}
	}
	return out
}

// FromCard from account card.
func (ai *AccInfo) FromCard(c *accmdl.ProfileStatReply) {
	ai.Mid = c.Profile.Mid
	ai.Name = c.Profile.Name
	ai.Rank = c.Profile.Rank
	ai.Face = c.Profile.Face
	ai.FaceNft = c.Profile.FaceNftNew
	ai.Sex = c.Profile.Sex
	ai.Silence = c.Profile.Silence
	ai.Sign = c.Profile.Sign
	ai.Level = c.Profile.Level
	ai.Official = c.Profile.Official
	ai.Vip = c.Profile.Vip
	ai.Pendant = c.Profile.Pendant
	ai.Nameplate = c.Profile.Nameplate
	ai.Coins = c.Coins
	ai.UserHonourInfo = &accmdl.UserHonourInfo{
		Tags: filterHonorTag(c.GetUserHonourInfo().GetTags()),
	}
	ai.Birthday = c.Profile.Birthday.Time().Format("01-02")
	ai.Profession = &Profession{
		Name:       c.GetProfile().Profession.GetName(),
		Department: c.GetProfile().Profession.GetDepartment(),
		Title:      c.GetProfile().Profession.GetTitle(),
		IsShow:     c.GetProfile().Profession.GetIsShow(),
	}
	ai.School = &School{
		Name: c.GetSchool().GetName(),
	}
	ai.IsSeniorMember = c.Profile.IsSeniorMember
}

var (
	// DefaultProfileStat .
	DefaultProfileStat = &accmdl.ProfileStatReply{
		Profile:   DefaultProfile,
		LevelInfo: accmdl.LevelInfo{},
	}
	// DefaultProfile .
	DefaultProfile = &accmdl.Profile{
		Name: "bilibili",
		Sex:  "保密",
		Face: "https://static.hdslb.com/images/member/noface.gif",
		Sign: "哔哩哔哩 (゜-゜)つロ 干杯~-bilibili",
		Rank: 5000,
	}
)

// ProfileStat profile with stat.
type ProfileStat struct {
	*accmdl.Profile
	LevelExp  accmdl.LevelInfo `json:"level_exp"`
	Coins     float64          `json:"coins"`
	Following int64            `json:"following"`
	Follower  int64            `json:"follower"`
}
