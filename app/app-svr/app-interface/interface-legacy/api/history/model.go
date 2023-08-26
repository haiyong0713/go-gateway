package history

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-gateway/pkg/idsafe/bvid"

	cardm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/history"
)

func FromCardUGC(l *history.ListRes) (card *CursorItem_CardUgc) {
	var (
		page     int32
		subtitle string
	)
	if l.Videos > 1 { //多p视频
		page = l.History.Page
		subtitle = l.History.Part
	}
	bvID, _ := bvid.AvToBv(l.History.Oid)
	var shareSubtitle string
	//nolint:gomnd
	if l.View > 10000 {
		tmpView := strconv.FormatFloat(float64(l.View)/10000, 'f', 1, 64)
		shareSubtitle = "已观看" + strings.TrimSuffix(tmpView, ".0") + "万次"
	}
	return &CursorItem_CardUgc{
		CardUgc: &CardUGC{
			Cover:         l.Cover,
			Progress:      l.Progress,
			Duration:      l.Duration,
			Name:          l.Name,
			Mid:           l.Mid,
			Cid:           l.History.Cid,
			Page:          page,
			Subtitle:      subtitle,
			Bvid:          bvID,
			Videos:        l.Videos,
			ShortLink:     fmt.Sprintf("https://b23.tv/%s", bvID),
			ShareSubtitle: shareSubtitle,
			View:          l.View,
			State:         l.State,
			Badge:         l.Badge,
		},
	}
}

func FromCardOGV(l *history.ListRes) (card *CursorItem_CardOgv) {
	return &CursorItem_CardOgv{
		CardOgv: &CardOGV{
			Cover:    l.Cover,
			Progress: l.Progress,
			Duration: l.Duration,
			Subtitle: l.ShowTitle,
			Badge:    l.Badge,
			State:    l.State,
		},
	}
}

func FromCardArticle(l *history.ListRes) (card *CursorItem_CardArticle) {
	return &CursorItem_CardArticle{
		CardArticle: &CardArticle{
			Covers:           l.Covers,
			Name:             l.Name,
			Mid:              l.Mid,
			DisplayAttention: l.DisAtten == 1,
			Badge:            l.Badge,
			Relation:         buildRelation(l.Relation),
		},
	}
}

func FromCardLive(l *history.ListRes) (card *CursorItem_CardLive) {
	return &CursorItem_CardLive{
		CardLive: &CardLive{
			Cover:            l.Cover,
			Name:             l.Name,
			Mid:              l.Mid,
			Tag:              l.TagName,
			Status:           int32(l.LiveStatus),
			DisplayAttention: l.DisAtten == 1,
			Relation:         buildRelation(l.Relation),
		},
	}
}

func FromCardCheese(l *history.ListRes) (card *CursorItem_CardCheese) {
	return &CursorItem_CardCheese{
		CardCheese: &CardCheese{
			Cover:    l.Cover,
			Progress: l.Progress,
			Duration: l.Duration,
			Subtitle: l.ShowTitle,
			State:    l.State,
		},
	}
}

func buildRelation(rel *cardm.Relation) *Relation {
	if rel == nil {
		return nil
	}
	return &Relation{
		Status:     int32(rel.Status),
		IsFollow:   int32(rel.IsFollow),
		IsFollowed: int32(rel.IsFollowed),
	}
}

func FromTitle(title, kw string) string {
	if kw == "" {
		return title
	}
	reg, err := regexp.Compile(fmt.Sprintf("(?i)%s", regexp.QuoteMeta(kw)))
	if err != nil {
		log.Error("FromTitle regexp.Compile err(%+v)", err)
		return title
	}
	return reg.ReplaceAllStringFunc(title, func(s string) string {
		return fmt.Sprintf("<em class=\"keyword\">%s</em>", s)
	})
}
