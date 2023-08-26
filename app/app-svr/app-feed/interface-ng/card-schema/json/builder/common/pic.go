package jsoncommon

import (
	"fmt"
	"strconv"

	"go-gateway/app/app-svr/app-card/interface/model"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
)

type PicCommon struct{}

func (PicCommon) ConstructPictureURI(dynamicID int64) string {
	return appcardmodel.FillURI(appcardmodel.GotoPicture, 0, 0, strconv.FormatInt(dynamicID, 10), nil)
}

func (PicCommon) ConstructArgsFromPicture(picture *bplus.Picture) jsoncard.Args {
	out := jsoncard.Args{
		UpID:   picture.Mid,
		UpName: picture.NickName,
	}
	if len(picture.TopicInfos) > 0 {
		out.Tid = picture.TopicInfos[0].TopicID
		out.Tname = picture.TopicInfos[0].TopicName
	}
	for _, ti := range picture.TopicInfos {
		if ti.IsActivity == 1 {
			out.ReportExtraInfo = &jsoncard.Report{
				DynamicActivity: ti.TopicName,
			}
			break
		}
	}
	return out
}

func (PicCommon) ConstructDescButtonFromPicture(picture *bplus.Picture) *jsoncard.Button {
	if len(picture.Topics) <= 0 {
		return nil
	}
	out := &jsoncard.Button{
		Type:    appcardmodel.ButtonGrey,
		Text:    picture.Topics[0],
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
	}
	if picture.IsNewChannel {
		for _, ti := range picture.TopicInfos {
			if ti.TopicName == picture.Topics[0] {
				out.URI = ti.TopicLink
			}
		}
	}
	if out.URI == "" {
		out.URI = appcardmodel.FillURI(appcardmodel.GotoPictureTag, 0, 0, picture.Topics[0], nil)
	}
	return out
}

func (PicCommon) ConstructThreePointFromPicture(picture *bplus.Picture) *jsoncard.ThreePoint {
	return &jsoncard.ThreePoint{
		DislikeReasons: constructThreePointDislikeReason(picture),
	}
}

func (PicCommon) ConstructThreePointV2FromPicture(picture *bplus.Picture) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	dislikeReasons := constructThreePointDislikeReason(picture)
	out = append(out, &jsoncard.ThreePointV2{
		Title:    "不感兴趣",
		Subtitle: "(选择后将减少相似内容推荐)",
		Reasons:  dislikeReasons,
		Type:     model.ThreePointDislike,
	})
	return out
}

func constructThreePointDislikeReason(picture *bplus.Picture) []*jsoncard.DislikeReason {
	dislikeReasons := []*jsoncard.DislikeReason{}
	if picture.NickName != "" {
		dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
			ID:    _upper,
			Name:  fmt.Sprintf("UP主:%s", picture.NickName),
			Toast: _dislikeToast,
		})
	}
	if len(picture.TopicInfos) > 0 && picture.TopicInfos[0].TopicName != "" {
		dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{
			ID:    _channel,
			Name:  fmt.Sprintf("话题:%s", picture.TopicInfos[0].TopicName),
			Toast: _dislikeToast,
		})
	}
	dislikeReasons = append(dislikeReasons, &jsoncard.DislikeReason{ID: _noSeason, Name: "不感兴趣", Toast: _dislikeToast})
	return dislikeReasons
}
