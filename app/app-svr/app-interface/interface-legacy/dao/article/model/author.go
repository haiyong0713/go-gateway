package model

// RecommendAuthorResp .
type RecommendAuthorResp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Trackid string `json:"trackid"`
	Data    []struct {
		UpID      int64  `json:"up_id"`
		RecReason string `json:"rec_reason"`
		RecType   int    `json:"rec_type"`
		Tid       int    `json:"tid"`
		Stid      int    `json:"second_tid"`
	} `json:"data"`
}

// RecommendAuthors .
type RecommendAuthors struct {
	Count   int          `json:"count"`
	Trackid string       `json:"trackid"`
	Authors []*RecAuthor `json:"authors"`
}

// RecAuthor .
type RecAuthor struct {
	*AccountCard
	RecReason string `json:"rec_reason"`
}

// WenhaoResp .
type WenhaoResp struct {
	Articles int `json:"articles"`
	Likes    int `json:"likes"`
}

// Empty empty struct.
type Empty struct {
}

// AllowResp .
type AllowResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Mid                 int  `json:"mid"`
		AllAllowed          bool `json:"all_allowed"`
		ControlActionStatus struct {
			SendDm struct {
				Allowed             bool   `json:"allowed"`
				DeniedByControlRole string `json:"denied_by_control_role"`
				IsExpirable         bool   `json:"is_expirable"`
				ExpireAt            int    `json:"expire_at"`
			} `json:"send-dm"`
			SendReply struct {
				Allowed             bool   `json:"allowed"`
				DeniedByControlRole string `json:"denied_by_control_role"`
				IsExpirable         bool   `json:"is_expirable"`
				ExpireAt            int    `json:"expire_at"`
			} `json:"send-reply"`
		} `json:"control_action_status"`
	} `json:"data"`
}
