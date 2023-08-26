package custom

import "time"

// MenuTabExt .
type Config struct {
	ID               int64     `json:"id"`
	Tp               int32     `json:"tp"`
	Oid              int64     `json:"oid"`
	Content          string    `json:"content"`
	Url              string    `json:"url"`
	HighlightContent string    `json:"highlight_content"`
	Image            string    `json:"image"`
	ImageBig         string    `json:"image_big"`
	State            int32     `json:"state"`
	OriginType       int32     `json:"origin_type"`
	AuditCode        int32     `json:"audit_code"`
	STime            time.Time `json:"stime"`
	ETime            time.Time `json:"etime"`
	CTime            time.Time `json:"ctime"`
	MTime            time.Time `json:"mtime"`
}
