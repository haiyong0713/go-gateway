package model

const (
	GotoAv        = "av"
	GotoBv        = "bv"
	GotoSearch    = "search"
	GotoArticle   = "article"
	GotoURL       = "url"
	GotoBangumi   = "bangumi"
	GotoPGCSeason = "pgc_season"
	GotoPGCEP     = "pgc_ep"
	GotoGame      = "game"
	GotoSpecial   = "special"

	PlatH5  = 15
	PlatXcx = 16

	LangHans = "hans"
	LangHant = "hant"
)

// FillURI deal app schema.
func FillURI(gt, param string, f func(uri string) string) (uri string) {
	switch gt {
	case GotoSearch:
		uri = "https://search.bilibili.com/all?keyword=" + param
	case GotoAv:
		uri = "https://www.bilibili.com/video/av" + param
	case GotoBv:
		uri = "https://www.bilibili.com/video/" + param
	case GotoArticle:
		uri = "https://www.bilibili.com/read/cv" + param
	case GotoPGCSeason:
		uri = "https://www.bilibili.com/bangumi/play/ss" + param
	case GotoPGCEP:
		uri = "https://www.bilibili.com/bangumi/play/ep" + param
	case GotoURL:
		uri = param
	}
	if f != nil {
		uri = f(uri)
	}
	return
}
