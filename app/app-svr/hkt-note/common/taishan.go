package common

import "go-common/library/net/rpc/warden"

type TaishanTableConfig struct {
	Name string `json:"name"`
	Auth struct {
		Token string `json:"token"`
	} `json:"auth"`
}

type TaishanCfg struct {
	TaishanRpc *warden.ClientConfig
	NoteReply  *TaishanTableConfig
}

var TaishanConfig = &TaishanCfg{
	TaishanRpc: &warden.ClientConfig{},
	NoteReply:  &TaishanTableConfig{},
}

// taishan key
const (
	//note_reply_format_{noteid} 笔记在评论区的新旧样式,value是区分新旧样式type
	Note_Reply_Format_Taishan_Key = "note_reply_format_%d"
	// cvid_mapping_rpid_{cvid} 公开笔记（专栏id）和 根评论的映射关系 key是cvid，value是rpid
	Cvid_Mapping_Rpid_Taishan_Key = "cvid_mapping_rpid_%d"
	// rpid_mapping_cvid_{rpid} 根评论到公开笔记（专栏id）的映射关系，key是rpid，value是cvid
	Rpid_Mapping_Cvid_Taishan_Key = "rpid_mapping_cvid_%d"
	// cvid_mapping_opid_     公开笔记 和 运营位的映射关系  key是cvid,value是opid
	Cvid_Mapping_Opid_Taishan_Key = "cvid_mapping_opid_%d"

	Cvid_Rpid_Attached     = 1
	Cvid_Rpid_Non_Attached = 2

	Cvid_Opid_Attached     = 1
	Cvid_Opid_Non_Attached = 2
)

type TaishanNoteReplyFormatInfo struct {
	FormatType int32 `json:"format-type"` //为1表示旧样式，为2表示新样式
}
type TaishanCvidMappingRpidInfo struct {
	Rpid   int64 `json:"rpid"`   //根评论id
	Status int32 `json:"status"` //关联状态
}

type TaishanRpidMappingCvidInfo struct {
	Cvid   int64 `json:"cvid"`   //专栏id
	Status int32 `json:"status"` //关联状态
}

type TaishanCvidMappingOpidInfo struct {
	Opid   int64 `json:"opid"`   //运营位id
	Status int32 `json:"status"` //关联状态
}
