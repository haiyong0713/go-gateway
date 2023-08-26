package common

const (
	// 本地上传的图片
	ImageUploadSourceLocal = "local"
	// 稿件截屏的图片
	ImageUploadSourceVideo         = "video"
	ImageShowInReplyMaxLimitDefault = 3
)

type ContentBody struct {
	Insert     interface{} `json:"insert"` // 只有string类型的正文
	Attributes interface{} `json:"attributes,omitempty"`
}



type ContentInsert struct {
	ImageUpload  *ImageUploadInsert `json:"imageUpload,omitempty"`
}

type ImageUploadInsert struct {
	Url string `json:"url,omitempty"`
	Source string `json:"source,omitempty"`
}