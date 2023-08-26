package share

// ShareInfoParam struct
type InfoParam struct {
	Bvid string `form:"bvid"`
	Mid  int64  `form:"mid"` // 分享人mid
}

type InfoReply struct {
	Arc       *ArcReply       `json:"arc,omitempty"`
	Author    *AuthorReply    `json:"author,omitempty"`
	Requester *RequesterReply `json:"requester,omitempty"`
}

type ArcReply struct {
	Aid            int64  `json:"aid,omitempty"`
	Title          string `json:"title,omitempty"`
	Pic            string `json:"pic,omitempty"`
	Duration       int64  `json:"duration,omitempty"`
	PremiereStatus int    `json:"premiere_status,omitempty"`
}

type AuthorReply struct {
	// Up主名称
	Name string `json:"name,omitempty"`
	// Up主头像地址
	Face string `json:"face,omitempty"`
	// 粉丝数
	Fans int64 `json:"fans,omitempty"`
}

type RequesterReply struct {
	// Up主名称
	Name string `json:"name,omitempty"`
	// Up主头像地址
	Face string `json:"face,omitempty"`
}
