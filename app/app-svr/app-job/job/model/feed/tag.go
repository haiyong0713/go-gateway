package feed

import (
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

type Tag struct {
	ID   int64  `json:"tag_id,omitempty"`
	Name string `json:"tag_name,omitempty"`
}

// AITag def.
func (v *Tag) AITag() *taggrpc.Tag {
	return &taggrpc.Tag{
		Id:   v.ID,
		Name: v.Name,
	}
}
