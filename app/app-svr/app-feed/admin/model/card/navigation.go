package card

import (
	"encoding/json"

	"go-common/library/time"

	"go-gateway/app/app-svr/app-feed/ecode"
)

const (
	SECOND_LEVEL_LIMIT = 10
	THIRD_LEVEL_LIMIT  = 30
)

type NavigationCard struct {
	Id         int64            `json:"id"`
	Title      string           `json:"title"`
	Desc       string           `json:"desc"`
	Cover      *NavCover        `json:"cover"`
	Corner     *NavCorner       `json:"corner"`
	Button     *NavButton       `json:"button"`
	Navigation *Navigation      `json:"navigation"`
	Count      *NavigationCount `json:"count"`
	Ctime      time.Time        `json:"ctime"`
	Mtime      time.Time        `json:"mtime"`
	CUname     string           `json:"c_uname"`
	MUname     string           `json:"m_uname"`
}

type AddNavigationCardReq struct {
	Uid        int64       `json:"uid" form:"uid"`
	Username   string      `json:"username" form:"username"`
	Title      string      `json:"title" form:"title" validate:"required"`
	Desc       string      `json:"desc" form:"desc"`
	Cover      *NavCover   `json:"cover" form:"cover"`
	Corner     *NavCorner  `json:"corner" form:"corner"`
	Button     *NavButton  `json:"button" form:"button"`
	Navigation *Navigation `json:"navigation" form:"navigation" validate:"required"`
}

type AddNavigationCardResp struct {
	CardId int64 `json:"card_id"`
}

type UpdateNavigationCardReq struct {
	Uid        int64       `json:"uid" form:"uid"`
	Username   string      `json:"username" form:"username"`
	CardId     int64       `json:"card_id" form:"card_id" validate:"required"`
	Title      string      `json:"title" form:"title" validate:"required"`
	Desc       string      `json:"desc" form:"desc"`
	Cover      *NavCover   `json:"cover" form:"cover"`
	Corner     *NavCorner  `json:"corner" form:"corner"`
	Button     *NavButton  `json:"button" form:"button"`
	Navigation *Navigation `json:"navigation" form:"navigation" validate:"required"`
}

type DeleteNavigationCardReq struct {
	Uid      int64  `json:"uid" form:"uid"`
	Username string `json:"username" form:"username"`
	CardId   int64  `json:"card_id" form:"card_id" validate:"required"`
}

type QueryNavigationCardReq struct {
	Uid      int64  `json:"uid" form:"uid"`
	Username string `json:"username" form:"username"`
	CardId   int64  `json:"card_id" form:"card_id" validate:"required"`
}

type QueryNavigationCardResp struct {
	CardId     int64       `json:"card_id"`
	Title      string      `json:"title"`
	Desc       string      `json:"desc"`
	Cover      *NavCover   `json:"cover"`
	Corner     *NavCorner  `json:"corner"`
	Button     *NavButton  `json:"button"`
	Navigation *Navigation `json:"navigation"`
	Ctime      time.Time   `json:"ctime"`
	Mtime      time.Time   `json:"mtime"`
	CUname     string      `json:"c_uname"`
	MUname     string      `json:"m_uname"`
}

type ListNavigationCardReq struct {
	Uid      int64  `json:"uid" form:"uid"`
	Username string `json:"username" form:"username"`
	CardId   int64  `json:"card_id" form:"card_id"`
	Keyword  string `json:"keyword" form:"keyword"`
	Pn       int    `json:"pn" form:"pn" default:"1"`
	Ps       int    `json:"ps" form:"ps" default:"20"`
}

type ListNavigationCardResp struct {
	Page *Page                 `json:"page"`
	List []*NavigationListItem `json:"list"`
}

type NavCover struct {
	Type     int32  `json:"cover_type"`
	SunPic   string `json:"sun_pic"`
	NightPic string `json:"night_pic"`
	Width    int32  `json:"width"`
	Height   int32  `json:"height"`
}

type NavCorner struct {
	Type     int32  `json:"type"`
	Text     string `json:"text"`
	SunPic   string `json:"sun_pic"`
	NightPic string `json:"night_pic"`
	Width    int32  `json:"width"`
	Height   int32  `json:"height"`
}

type NavButton struct {
	Type    int32  `json:"type"`
	Text    string `json:"text"`
	ReType  int32  `json:"re_type"`
	ReValue string `json:"re_value"`
}

type Navigation struct {
	ModuleCount int32            `json:"module_count"`
	Children    []*Navigation2nd `json:"children"`
}

type Navigation2nd struct {
	Title     string           `json:"title"`
	Deletable int32            `json:"deletable"`
	Button    *NavButton       `json:"button"`
	Children  []*Navigation3rd `json:"children"`
}

type Navigation3rd struct {
	Title     string `json:"title"`
	ReType    int32  `json:"re_type"`
	ReValue   string `json:"re_value"`
	Deletable int32  `json:"deletable"`
}

type NavigationCount struct {
	SecondLevel int `json:"second_level"`
	ThirdLevel  int `json:"third_level"`
}

type NavigationExtraInfo struct {
	Navigation      *Navigation      `json:"nav"`
	NavigationCount *NavigationCount `json:"nav_count"`
}

type NavigationListItem struct {
	CardId          int64            `json:"card_id"`
	Title           string           `json:"title"`
	Desc            string           `json:"desc"`
	Cover           *NavCover        `json:"cover"`
	Ctime           time.Time        `json:"ctime"`
	Mtime           time.Time        `json:"mtime"`
	CUname          string           `json:"c_uname"`
	MUname          string           `json:"m_uname"`
	NavigationCount *NavigationCount `json:"navigation_count"`
}

type Page struct {
	Pn    int `json:"num"`
	Ps    int `json:"size"`
	Total int `json:"total"`
}

func ConvertNavigationCard(title, desc string, cover *NavCover, corner *NavCorner, button *NavButton, nav *Navigation) (card *ResourceCard, err error) {
	var (
		coverBytes, cornerBytes, btnBytes, extraBytes []byte
		extra                                         = &NavigationExtraInfo{NavigationCount: &NavigationCount{}}
	)

	card = &ResourceCard{
		Title:    title,
		Desc:     desc,
		CardType: CardTypeNavigation,
	}

	if nav.Children == nil {
		return nil, ecode.Navigation2ndEmpty
	}
	for _, child := range nav.Children {
		if child.Children == nil {
			return nil, ecode.Navigation3rdEmpty
		}
		extra.NavigationCount.ThirdLevel += len(child.Children)
	}
	if extra.NavigationCount.SecondLevel > SECOND_LEVEL_LIMIT {
		return nil, ecode.Navigation2ndExceeds
	}
	if extra.NavigationCount.ThirdLevel > THIRD_LEVEL_LIMIT {
		return nil, ecode.Navigation3rdExceeds
	}
	extra.NavigationCount.SecondLevel = len(nav.Children)
	extra.Navigation = nav

	if extraBytes, err = json.Marshal(extra); err != nil {
		return
	}
	if coverBytes, err = json.Marshal(cover); err != nil {
		return
	}
	if cornerBytes, err = json.Marshal(corner); err != nil {
		return
	}
	if btnBytes, err = json.Marshal(button); err != nil {
		return
	}
	card.ExtraInfo = string(extraBytes)
	card.Cover = string(coverBytes)
	card.Corner = string(cornerBytes)
	card.Button = string(btnBytes)

	return
}

func ParseNavigationCard(card *ResourceCard) (ret *NavigationCard, err error) {
	var (
		extra = &NavigationExtraInfo{}
	)

	ret = &NavigationCard{
		Id:     card.Id,
		Title:  card.Title,
		Desc:   card.Desc,
		Ctime:  card.Ctime,
		Mtime:  card.Mtime,
		CUname: card.CUname,
		MUname: card.MUname,
	}

	if len(card.Cover) > 0 {
		if err = json.Unmarshal([]byte(card.Cover), &ret.Cover); err != nil {
			return
		}
	}
	if len(card.Corner) > 0 {
		if err = json.Unmarshal([]byte(card.Corner), &ret.Corner); err != nil {
			return
		}
	}
	if len(card.Button) > 0 {
		if err = json.Unmarshal([]byte(card.Button), &ret.Button); err != nil {
			return
		}
	}
	if len(card.ExtraInfo) > 0 {
		if err = json.Unmarshal([]byte(card.ExtraInfo), extra); err != nil {
			return
		}
		ret.Navigation = extra.Navigation
		ret.Count = extra.NavigationCount
	}

	return
}
