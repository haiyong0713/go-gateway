package article

import "time"

const (
	KeyRetryArtBinlog = "article_binlog_retry"
	PubFromReply      = 1
)

type PubFailMsg struct {
	NoteId  int64
	Mid     int64
	Reason  string
	ErrCode int
}

func (v *PubFailMsg) ToCache() *ArtDtlCache {
	return &ArtDtlCache{
		Cvid:       -1,
		NoteId:     v.NoteId,
		Mid:        v.Mid,
		PubStatus:  PubStatusBreak,
		PubReason:  v.Reason,
		PubVersion: time.Now().Unix(),
	}
}
