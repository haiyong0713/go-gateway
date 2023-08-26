package view

import (
	"fmt"
	"strconv"

	"github.com/thoas/go-funk"

	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	musicmdl "go-gateway/app/app-svr/app-view/interface/model/music"

	channelApi "git.bilibili.co/bapis/bapis-go/community/interface/channel"

	notes "go-gateway/app/app-svr/hkt-note/service/api"
)

var (
	//cell type
	CellRange           = []string{CellS11Type, CellUgcTabType, CellBiJianType, CellToolType, CellInspirationType, CellMusicType, CellNoteType, CellOgvTabType}
	CellMusicType       = "bgm"
	CellNoteType        = "notes"
	CellTopicType       = "topic"
	CellS11Type         = "s11"
	CellUgcTabType      = "ugc_tab"
	CellOgvTabType      = "ogv_tab"
	CellBiJianType      = "bijian"
	CellToolType        = "tool"
	CellInspirationType = "Inspiration"
	CellNotes           = "UP主笔记" //笔记标签
	PlayMusicIcon       = "https://i0.hdslb.com/bfs/activity-plat/static/20220406/c4a9becc6e5e0e555db370b87b6e4bac/cGKI45H89U.svga"
	PlayMusicStaticIcon = "https://i0.hdslb.com/bfs/activity-plat/static/20220401/c4a9becc6e5e0e555db370b87b6e4bac/wd0MrcvzzQ.png"
	CellTextColor       = "#61666D" //标签-文本颜色
	CellTextColorNight  = "#A2A7AE" //标签-文本颜色夜间
	CellColor           = "#F6F7F8" //cell背景
	CellColorNight      = "#1E2022" //cell  背景夜间
	IconS11             = "https://i0.hdslb.com/bfs/app/6659a545e1cb0895d01fd74dd6b507540c4ff7b9.png"
	IconS11Night        = "https://i0.hdslb.com/bfs/app/b1ddacfcd456744d9cf8d1b243733b69e42f9a3f.png"
	IconTopic           = "https://i0.hdslb.com/bfs/app/95d83802f5339262d1b9b8037f9d7e5918f275dd.png"
	IconTopicNight      = "https://i0.hdslb.com/bfs/app/7f0525aff3deb1109be324da3dd901e7f37d91f0.png"
	IconNote            = "https://i0.hdslb.com/bfs/app/ce9a3bb548498e256f2306bc0903aad468309b39.png"
	IconNoteNight       = "https://i0.hdslb.com/bfs/app/2b4cede154ea2a99409083eeaf697a7f84a3aacf.png"
	IconMusic           = "https://i0.hdslb.com/bfs/app/0e138478bef077b3d6d4613882266c830291245a.png"
	IconMusicNight      = "https://i0.hdslb.com/bfs/app/e219f12990332ddf5d69817233cd510687f625e5.png"
	EndIconNight        = "https://i0.hdslb.com/bfs/app/ef84b1679018e26f1a0bedef0ebf54e55807d8f6.png"
	EndIcon             = "https://i0.hdslb.com/bfs/app/7c5a689d5063208fa32293c58f4a3d2bfb94cfb5.png"
	IconOgv             = "https://i0.hdslb.com/bfs/activity-plat/static/20220516/c7eae9464cdbb8d9eba2eeb835c8f0a4/CabLQvgPkt.png"
	IconOgvNight        = "https://i0.hdslb.com/bfs/activity-plat/static/20220516/c7eae9464cdbb8d9eba2eeb835c8f0a4/WeAkPi2oAV.png"
)

func SpecialCellPriorityNewVersion(ogvChan *channelApi.Channel, music *musicmdl.MusicInfo, note *notes.ArcTagReply, spmid string, topicCell *viewApi.SpecialCell, cellArr map[string]*viewApi.SpecialCell) []*viewApi.SpecialCell {
	if cellArr == nil {
		cellArr = make(map[string]*viewApi.SpecialCell)
	}
	if ogvChan != nil {
		cellArr[CellOgvTabType] = &viewApi.SpecialCell{
			Icon:             IconOgv,
			IconNight:        IconOgvNight,
			Text:             ogvChan.Name,
			TextColor:        CellTextColor,
			TextColorNight:   CellTextColorNight,
			CellType:         CellOgvTabType,
			EndIcon:          EndIcon,
			EndIconNight:     EndIconNight,
			CellBgcolor:      CellColor,
			CellBgcolorNight: CellColorNight,
			JumpType:         "new_page",
			JumpUrl:          fmt.Sprintf("bilibili://feed/channel?biz_id=%d&biz_type=0&source=%s", ogvChan.ID, spmid),
			Param:            fmt.Sprintf("%d", ogvChan.ID),
		}
	}
	if music != nil {
		cellArr[CellMusicType] = &viewApi.SpecialCell{
			Icon:             IconMusic,
			IconNight:        IconMusicNight,
			Text:             music.MusicTitle,
			TextColor:        CellTextColor,
			TextColorNight:   CellTextColorNight,
			CellType:         CellMusicType,
			EndIcon:          EndIcon,
			EndIconNight:     EndIconNight,
			CellBgcolor:      CellColor,
			CellBgcolorNight: CellColorNight,
			JumpUrl:          music.JumpUrl,
			Param:            music.MusicId,
		}
	}
	if note != nil && note.JumpLink != "" && note.TagShowText != "" {
		cellArr[CellNoteType] = &viewApi.SpecialCell{
			Icon:             IconNote,
			IconNight:        IconNoteNight,
			Text:             CellNotes,
			TextColor:        CellTextColor,
			TextColorNight:   CellTextColorNight,
			CellType:         CellNoteType,
			EndIcon:          EndIcon,
			EndIconNight:     EndIconNight,
			CellBgcolor:      CellColor,
			CellBgcolorNight: CellColorNight,
			JumpUrl:          note.JumpLink + "&detail=ugc_useful_area",
			Param:            strconv.FormatInt(note.NoteId, 10),
			NotesCount:       note.NotesCount,
		}
	}
	if v, ok := cellArr[CellInspirationType]; ok && topicCell != nil {
		if funk.Contains(v.Text, topicCell.Text) { //如果灵感中的名字包含话题名，说明重名，则要话题不展示
			topicCell = nil
		}
	}
	//排序
	var rly []*viewApi.SpecialCell
	for _, v := range CellRange {
		if _, ok := cellArr[v]; ok {
			rly = append(rly, cellArr[v])
		}
		if len(rly) == int(3) { //最多拿3个
			break
		}
	}
	//脏逻辑:保证第三个是话题 || 最后一个是话题
	if len(rly) < 3 && topicCell != nil {
		topicCell.Icon = IconTopic
		topicCell.IconNight = IconTopicNight
		rly = append(rly, topicCell)
	}
	if len(rly) == 3 && rly[2].CellType != CellTopicType && topicCell != nil {
		topicCell.Icon = IconTopic
		topicCell.IconNight = IconTopicNight
		rly[2] = topicCell
	}
	return rly
}
