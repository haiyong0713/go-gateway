package dolby

type ConfigParam struct {
	Brand string `form:"brand"`
	Model string `form:"model"`
}

type ConfigReply struct {
	File string `json:"file"`
	Hash string `json:"hash"`
}
