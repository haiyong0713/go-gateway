package web

// ArcMsg archive .
type ArcMsg struct {
	Action string      `json:"action"`
	Table  string      `json:"table"`
	New    *ArchiveSub `json:"new"`
	Old    *ArchiveSub `json:"old"`
}

// ArchiveSub archive .
type ArchiveSub struct {
	Aid     int64  `json:"aid"`
	Mid     int64  `json:"mid"`
	PubTime string `json:"pubtime"`
	CTime   string `json:"ctime"`
	MTime   string `json:"mtime"`
	State   int    `json:"state"`
}

type OutArcMsg struct {
	Action string     `json:"action"`
	Table  string     `json:"table"`
	New    *OutArcSub `json:"new"`
	Old    *OutArcSub `json:"old"`
}

type OutArcSub struct {
	Available int64  `json:"available"`
	Avid      int64  `json:"avid"`
	Click     int64  `json:"click"`
	IsGranted int64  `json:"is_granted"`
	Mid       int64  `json:"mid"`
	Pubtime   string `json:"pubtime"`
	Sdate     string `json:"sdate"`
}
