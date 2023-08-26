package card

import (
	"encoding/json"
	"go-common/library/time"
)

const (
	ContentReTypeUrl       = 0
	ContentReTypeGameCardS = 1
	ContentReTypeAv        = 2
	ContentReTypePgcEp     = 3
	ContentReTypeLive      = 4
)

const (
	ContentCTypeAv      = 0
	ContentCTypeLive    = 1
	ContentCTypeArticle = 2
)

type ContentCard struct {
	Id      int64       `json:"id"`
	Title   string      `json:"title"`
	Cover   *ContCover  `json:"cover"`
	Jump    *ContJump   `json:"jump"`
	Button  *ContButton `json:"button"`
	Content []*ContItem `json:"content"`
	Ctime   time.Time   `json:"ctime"`
	Mtime   time.Time   `json:"mtime"`
	CUname  string      `json:"c_uname"`
	MUname  string      `json:"m_uname"`
	Deleted int32       `json:"deleted"`
}

type ContCover struct {
	MCover string `json:"m_cover" form:"m_cover"`
}

type ContJump struct {
	ReType  int32  `json:"re_type" form:"re_type"`
	ReValue string `json:"re_value" form:"re_value"`
}

type ContExtra struct {
	Content []*ContItem `json:"content" form:"content"`
}

type ContButton struct {
	ReType  int32  `json:"re_type" form:"re_type"`
	ReValue string `json:"re_value" form:"re_value"`
}

type ContItem struct {
	Title   string `json:"ctitle" validate:"required"`
	ReType  int32  `json:"ctype" validate:"required"`
	ReValue string `json:"cvalue" validate:"required"`
}

type AddContentCardReq struct {
	Uid      int64       `json:"uid" form:"uid"`
	Username string      `json:"username" form:"username"`
	Title    string      `json:"title" form:"title" validate:"required"`
	Cover    *ContCover  `json:"cover" form:"cover"`
	Jump     *ContJump   `json:"jump" form:"jump"`
	Button   *ContButton `json:"button" form:"button"`
	Content  []*ContItem `json:"content" form:"content" validate:"required"`
}

type AddContentCardResp struct {
	CardId int64 `json:"card_id"`
}

type UpdateContentCardReq struct {
	Uid      int64       `json:"uid" form:"uid"`
	Username string      `json:"username" form:"username"`
	CardId   int64       `json:"card_id" form:"card_id" validate:"required"`
	Title    string      `json:"title" form:"title" validate:"required"`
	Cover    *ContCover  `json:"cover" form:"cover"`
	Jump     *ContJump   `json:"jump" form:"jump"`
	Button   *ContButton `json:"button" form:"button"`
	Content  []*ContItem `json:"content" form:"content" validate:"required"`
}

type DeleteContentCardReq struct {
	Uid      int64  `json:"uid" form:"uid"`
	Username string `json:"username" form:"username"`
	CardId   int64  `json:"card_id" form:"card_id" validate:"required"`
}

type QueryContentCardReq struct {
	Uid      int64  `json:"uid" form:"uid"`
	Username string `json:"username" form:"username"`
	CardId   int64  `json:"card_id" form:"card_id" validate:"required"`
}

type QueryContentCardResp struct {
	CardId  int64       `json:"card_id"`
	Title   string      `json:"title"`
	Cover   *ContCover  `json:"cover"`
	Jump    *ContJump   `json:"jump"`
	Button  *ContButton `json:"button"`
	Content []*ContItem `json:"content"`
	Ctime   time.Time   `json:"ctime"`
	Mtime   time.Time   `json:"mtime"`
	CUname  string      `json:"c_uname"`
	MUname  string      `json:"m_uname"`
}

type ListContentCardReq struct {
	Uid      int64  `json:"uid" form:"uid"`
	Username string `json:"username" form:"username"`
	CardId   int64  `json:"card_id" form:"card_id"`
	Keyword  string `json:"keyword" form:"keyword"`
	Pn       int    `json:"pn" form:"pn" default:"1"`
	Ps       int    `json:"ps" form:"ps" default:"20"`
}

type ListContentCardResp struct {
	Page *Page           `json:"page"`
	List []*ContListItem `json:"list"`
}

type ContListItem struct {
	CardId  int64       `json:"card_id"`
	Title   string      `json:"title"`
	Cover   *ContCover  `json:"cover"`
	Jump    *ContJump   `json:"jump"`
	Button  *ContButton `json:"button"`
	Content []*ContItem `json:"content"`
	Ctime   time.Time   `json:"ctime"`
	Mtime   time.Time   `json:"mtime"`
	CUname  string      `json:"c_uname"`
	MUname  string      `json:"m_uname"`
}

func ConvertContentCard(title string, cover *ContCover, jump *ContJump, btn *ContButton, content []*ContItem) (card *ResourceCard, err error) {
	var (
		coverBytes, jumpBytes, buttonBytes, extraBytes []byte
	)

	card = &ResourceCard{
		CardType: CardTypeContent,
		Title:    title,
	}

	if coverBytes, err = json.Marshal(cover); err != nil {
		return
	}
	card.Cover = string(coverBytes)

	if jumpBytes, err = json.Marshal(jump); err != nil {
		return
	}
	card.JumpInfo = string(jumpBytes)

	contExtra := &ContExtra{Content: content}
	if extraBytes, err = json.Marshal(contExtra); err != nil {
		return
	}
	card.ExtraInfo = string(extraBytes)

	if len(btn.ReValue) > 0 {
		if buttonBytes, err = json.Marshal(btn); err != nil {
			return
		}
		card.Button = string(buttonBytes)
	}

	return
}

func ParseContentCard(card *ResourceCard) (ret *ContentCard, err error) {
	var (
		extra = &ContExtra{}
	)

	ret = &ContentCard{
		Id:      card.Id,
		Title:   card.Title,
		Ctime:   card.Ctime,
		Mtime:   card.Mtime,
		CUname:  card.CUname,
		MUname:  card.MUname,
		Deleted: card.Deleted,
	}

	if len(card.Button) > 0 {
		if err = json.Unmarshal([]byte(card.Button), &ret.Button); err != nil {
			return
		}
	}
	if len(card.Cover) > 0 {
		if err = json.Unmarshal([]byte(card.Cover), &ret.Cover); err != nil {
			return
		}
	}
	if len(card.JumpInfo) > 0 {
		if err = json.Unmarshal([]byte(card.JumpInfo), &ret.Jump); err != nil {
			return
		}
	}
	if len(card.ExtraInfo) > 0 {
		if err = json.Unmarshal([]byte(card.ExtraInfo), extra); err != nil {
			return
		}
	}

	ret.Content = extra.Content
	return
}
