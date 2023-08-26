package medialist

// MediaListInfo mediaList info
// http://bapi.bilibili.co/project/4627/interface/api/323797
type MediaListInfo struct {
	MediaId        int64  `json:"media_id"`
	Mid            int64  `json:"mid"`
	UserName       string `json:"user_name"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Cover          string `json:"cover"`
	OriginCover    string `json:"origin_cover"`
	Attr           int32  `json:"attr"`
	FansCount      int32  `json:"fans_count"`
	Count          int32  `json:"count"`
	ChangedColumns int32  `json:"changed_columns"`
	Authentication int32  `json:"authentication"`
}
