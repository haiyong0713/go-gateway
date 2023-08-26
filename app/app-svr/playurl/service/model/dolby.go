package model

type DolbyInfo struct {
	Buvid     string
	Mid       int64
	Ctime     string
	MobiApp   string
	Platform  string
	Build     int32
	Aid       int64
	Cid       int64
	Type      string
	Scene     string
	DolbyType int64
}

type DolbyConf struct {
	Dolby         int64
	IsVip         bool
	TeenagersMode int64
	LessonsMode   int64
	MobiApp       string
	Device        string
}

func (dc *DolbyConf) SupportDolby() bool {
	if !dc.IsVip || dc.Dolby == 0 || dc.TeenagersMode == 1 || dc.LessonsMode == 1 ||
		dc.Device != "phone" || dc.MobiApp != "iphone" {
		return false
	}
	return true
}
