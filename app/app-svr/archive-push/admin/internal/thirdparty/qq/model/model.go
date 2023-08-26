package model

const (
	ContentTypeKey  = "content-type"
	ContentTypeJSON = "application/json"
)

type Config struct {
	CMC *CMCConfig
	TGL *TGLConfig
}

type CMCConfig struct {
	Host   string
	Secret string
	IBIZ   string
	Source string
	SExt9  string
	Oauth2 *OAuth2Config
}

type TGLConfig struct {
	Host   string
	Action int64
	GameID int64
	Oauth2 *OAuth2Config
}

type OAuth2Config struct {
	ClientID string
	Secret   string
}
