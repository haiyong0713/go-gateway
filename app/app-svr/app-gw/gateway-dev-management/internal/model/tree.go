package model

import (
	"encoding/json"
)

// Token token
type Token struct {
	Token    string `json:"token"`
	UserName string `json:"user_name"`
	Secret   string `json:"secret"`
	Expired  int64  `json:"expired"`
}

// TokenResult token result
type TokenResult struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
	Status  int             `json:"status"`
}

// Resp tree resp
type Resp struct {
	Data []*Node `json:"data"`
}

// Node node
type Node struct {
	TreeID      int    `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        int    `json:"type"`
	Role        int    `json:"role"`
	DiscoveryID string `json:"discovery_id"`
}

// Tree tree model
type Tree struct {
	Project string  `json:"project"`
	Subs    []*Tree `json:"subs"`
}

type TreeReply struct {
	Options []*Options `json:"options"`
}

type Options struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
