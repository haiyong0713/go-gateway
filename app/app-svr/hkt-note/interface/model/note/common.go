package note

type Device struct {
	AccessKey string `form:"access_key"`
	Platform  string `form:"platform"`
}

type Links struct {
	CheeseQALink string `json:"cheese_qa_link"`
}

func ToInfocPlat(device Device) int64 {
	switch device.Platform {
	case "web":
		return _platWeb
	case "android":
		return _platAndroid
	case "ios":
		return _platIOS
	}
	return 0
}
