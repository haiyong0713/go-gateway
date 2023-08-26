package bili_link

type BiliLinkReport struct {
	ActType   int64  `form:"act_type"`
	Id        int64  `form:"id"`
	AccessKey string `form:"access_key"`
	Mid       int64  `form:"-"`
}
