package model

const (
	TGLAccessTokenHeader    = "Authorization"
	TGLAccessTokenBody      = "Bearer %s"
	TGLAccessTokenGrantType = "client_credentials"
)

// ContributeVideoReq /contribute/video
type ContributeVideoReq struct {
	Action    int64  `form:"action" json:"action" validate:"required"`  // 投稿活动，由TGL分配
	GameID    int64  `form:"game_id" json:"game_id"`                    // 游戏ID，具体值从游戏列表接口获取，如：王者：362
	Title     string `form:"title" json:"title" validate:"required"`    // 标题，必须
	Summary   string `form:"summary" json:"summary"`                    // 摘要，可选
	VideoURL  string `form:"video_url" json:"video_url"`                // 视频源文件地址，和第三方平台视频ID二选一
	Cover     string `form:"cover" json:"cover" validate:"required"`    // 封面图，多个地址用英文逗号隔开，必须
	Author    string `form:"author" json:"author"`                      // 作者昵称，可选
	Duration  int64  `form:"duration" json:"duration" validate:"min=1"` // 视频时长，必须
	OuterVID  string `form:"outer_vid" json:"outer_vid"`                // 第三方平台用户
	OuterUser string `form:"outer_user" json:"outer_user"`              // 第三方平台用户
	ExtTags   string `form:"ext_tags" json:"ext_tags"`                  // 稿件tags
}

// ContributeVideoReply /contribute/video
type ContributeVideoReply struct {
	Status  int    `json:"status_code"`
	Message string `json:"message"`
}

// OauthAccessTokenReq /oauth/access_token
type OauthAccessTokenReq struct {
	GrantType    string `form:"grant_type" json:"grant_type"`
	ClientID     string `form:"client_id" json:"client_id"`
	ClientSecret string `form:"client_secret" json:"client_secret"`
}

// OauthAccessTokenReply /oauth/access_token
type OauthAccessTokenReply struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Status      int    `json:"status_code"`
	Message     string `json:"message"`
}
