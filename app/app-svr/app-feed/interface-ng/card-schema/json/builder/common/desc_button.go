package jsoncommon

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

func ConstructDescButtonFromTag(in *taggrpc.Tag) *jsoncard.Button {
	out := &jsoncard.Button{
		Type:    appcardmodel.ButtonGrey,
		Text:    in.Name,
		URI:     appcardmodel.FillURI(appcardmodel.GotoTag, 0, 0, strconv.FormatInt(in.Id, 10), nil),
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
	}
	return out
}

func ConstructDescButtonFromAuthor(in arcgrpc.Author) *jsoncard.Button {
	out := &jsoncard.Button{
		Type:    appcardmodel.ButtonGrey,
		Text:    in.Name,
		URI:     appcardmodel.FillURI(appcardmodel.GotoMid, 0, 0, strconv.FormatInt(in.Mid, 10), nil),
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
	}
	return out
}
