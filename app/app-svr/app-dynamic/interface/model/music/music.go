package music

// 音频卡资源
type MusicResItem struct {
	ID          int64  `json:"id"`          // 音频id
	UpId        int64  `json:"upId"`        // up主id
	Title       string `json:"title"`       // 动态标题
	Upper       string `json:"upper"`       // up主名称
	Cover       string `json:"cover"`       // 封面地址
	Author      string `json:"author"`      // 作者信息
	CTime       int64  `json:"ctime"`       // 过审时间，动态生成时间
	ReplyCnt    int32  `json:"replyCnt"`    // 评论数
	PlayCnt     int32  `json:"playCnt"`     // 播放数
	Intro       string `json:"intro"`       // 简介
	Schema      string `json:"schema"`      // 路由schema
	TypeInfo    string `json:"typeInfo"`    // 类型信息
	UpperAvatar string `json:"upperAvatar"` // up主头像
}
