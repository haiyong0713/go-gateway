package session

import (
	"context"
	"encoding/json"
)

// IndexSession is
type IndexSession struct {
	ID      string `json:"id"`
	Time    int64  `json:"time"`
	Mid     int64  `json:"mid"`
	Request struct {
		Header map[string][]string `json:"header"`
		Query  map[string][]string `json:"query"`
	} `json:"request"`
	Response            string                     `json:"response"`
	AIRecommendResponse string                     `json:"ai_recommend_response"`
	Annotation          map[string]json.RawMessage `json:"annotation"`
}

type indexSessionKey struct{}

// NewContext returns a new Context that carries value.
func NewContext(ctx context.Context, s *IndexSession) context.Context {
	return context.WithValue(ctx, indexSessionKey{}, s)
}

// FromContext returns the IndexSession pointer stored in ctx.
func FromContext(ctx context.Context) (*IndexSession, bool) {
	s, ok := ctx.Value(indexSessionKey{}).(*IndexSession)
	return s, ok
}
