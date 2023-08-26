package story

import (
	"fmt"

	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
)

type EntranceLoader interface {
	Icon() string
	JumpURI(item *ai.SubItems) string
	Type() int64
	Title(item *ai.SubItems) string
	Display() bool
}

const (
	_topic       = 1 // 话题
	_aspiration  = 2 // 志愿填报
	_tools       = 3 // 工具
	_inspiration = 4 // 灵感
	_music       = 5 // 音乐
	_search      = 6 // 搜索

	_cooperate   = 9   // 合拍
	_sticker     = 64  // 拍摄特效
	_rhythm1     = 14  // 模板1
	_rhythm2     = 44  // 模板2
	_rhythm3     = 46  // 模板3
	_picToVideo1 = 72  // 图片转视频
	_picToVideo2 = -72 // 图片转视频
	_bgm         = 127 // bgm
)

//go:generate easyjson -all creative_entrance.go

//easyjson:json
type EntranceExtra struct {
	MaterialType   int64  `json:"material_type,omitempty"`
	MaterialId     int64  `json:"material_id,omitempty"`
	InspirationId  int64  `json:"inspiration_id,omitempty"`
	IsJumpMaterial int8   `json:"is_jump_material,omitempty"`
	Tags           string `json:"tags,omitempty"`
	MaterialName   string `json:"material_name,omitempty"`
	MaterialSource string `json:"material_source,omitempty"` // 本次不需要
	Keyword        string `json:"keyword,omitempty"`
}

type empty struct {
}

func (e *empty) Icon() string {
	return ""
}
func (e *empty) JumpURI(_ *ai.SubItems) string {
	return ""
}
func (e *empty) Type() int64 {
	return 0
}
func (e *empty) Title(_ *ai.SubItems) string {
	return ""
}
func (e *empty) Display() bool {
	return false
}

type topic struct{}

func (t *topic) Icon() string {
	return "https://i0.hdslb.com/bfs/activity-plat/static/20220909/0977767b2e79d8ad0a36a731068a83d7/SrAArWWQyQ.png"
}
func (t *topic) JumpURI(item *ai.SubItems) string {
	return fmt.Sprintf("https://m.bilibili.com/topic-detail?topic_id=%d&page_from=story&from_aid=%d",
		item.TopicID, item.ID)
}
func (t *topic) Type() int64 {
	return _topic
}
func (t *topic) Title(item *ai.SubItems) string {
	return item.TopicTitle
}
func (t *topic) Display() bool {
	return true
}

type aspiration struct{}

func (a *aspiration) Icon() string {
	return "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/tYJst4HpgR.png"
}
func (a *aspiration) JumpURI(_ *ai.SubItems) string {
	return "bilibili://search?from=app_story_diversion&keyword=志愿填报&direct_return=true"
}
func (a *aspiration) Type() int64 {
	return _aspiration
}
func (a *aspiration) Title(item *ai.SubItems) string {
	return item.TopicTitle
}
func (a *aspiration) Display() bool {
	return true
}

type tools struct {
	extra EntranceExtra
}

func (t *tools) Icon() string {
	switch t.extra.MaterialType {
	case _cooperate:
		return "https://i0.hdslb.com/bfs/archive/3d0237b74de576b606e5f639b3f8349063e6fe24.png"
	case _sticker:
		return "https://i0.hdslb.com/bfs/archive/92f3d3c685b5ca48089d7c7d396a0a9bce329058.png"
	case _rhythm1, _rhythm2, _rhythm3:
		return "https://i0.hdslb.com/bfs/archive/1f9cf20aa94b03c545a60e469610a74a70fff474.png"
	case _picToVideo1, _picToVideo2:
		return "https://i0.hdslb.com/bfs/archive/e83966dba6e8ce448e79b4c37f8a418c3ec0d622.png"
	default:
		return ""
	}
}
func (t *tools) JumpURI(item *ai.SubItems) string {
	switch t.extra.MaterialType {
	case _cooperate:
		return fmt.Sprintf("bilibili://uper/center_plus?tab_index=1&relation_from=story&topic_id=%d&post_config={\"first_entrance\":\"竖屏信息流\"}&cooperate_id=%d", item.TopicID, t.extra.MaterialId)
	case _sticker:
		return fmt.Sprintf("bilibili://uper/center_plus?tab_index=1&relation_from=story&topic_id=%d&post_config={\"first_entrance\":\"竖屏信息流\"}&sticker_id_v2=%d", item.TopicID, t.extra.MaterialId)
	case _rhythm1, _rhythm2, _rhythm3:
		if toolNeedUserCenter(item.MobiApp(), item.Build()) {
			return fmt.Sprintf("bilibili://uper/user_center/add_archive/?from=2&topic_id=%d&rhythm_id_v2=%d&is_detail=1&relation_from=story&post_config={\"first_entrance\":\"竖屏信息流\"}", item.TopicID, t.extra.MaterialId)
		}
		return fmt.Sprintf("bilibili://uper/center_plus?tab_index=3&relation_from=story&topic_id=%d&post_config={\"first_entrance\":\"竖屏信息流\"}&rhythm_id_v2=%d", item.TopicID, t.extra.MaterialId)
	case _picToVideo1, _picToVideo2:
		return fmt.Sprintf("bilibili://uper/center_plus?tab_index=2&relation_from=story&topic_id=%d&post_config={\"first_entrance\":\"竖屏信息流\"}", item.TopicID)
	default:
		return ""
	}
}
func (t *tools) Type() int64 {
	return _tools
}
func (t *tools) Title(item *ai.SubItems) string {
	return item.TopicTitle
}
func (t *tools) Display() bool {
	return true
}

type inspiration struct {
	extra EntranceExtra
}

func (i *inspiration) Icon() string {
	return "https://i0.hdslb.com/bfs/archive/07a0e10dc6d8b883b7c88a0bdc2c18013e975c92.png"
}
func (i *inspiration) JumpURI(item *ai.SubItems) string {
	if i.extra.IsJumpMaterial == 0 || backup46ToInspiration(item.MobiApp(), item.Build(), i.extra.MaterialType) ||
		backup44ToInspiration(item.MobiApp(), item.Build(), i.extra.MaterialType) {
		return fmt.Sprintf("https://member.bilibili.com/york/up-inspiration/detail?id=%d&navhide=1&from=8", i.extra.InspirationId)
	}
	switch i.extra.MaterialType {
	case _cooperate:
		return fmt.Sprintf("bilibili://uper/center_plus?is_inspiration=1&tags=%s&relation_from=creative-inspiration-story_%d&post_config={\"first_entrance\":\"创作灵感story投稿点击\"}&topic_id=%d&tab_index=1&cooperate_id=%d", i.extra.Tags, i.extra.InspirationId, item.TopicID, i.extra.MaterialId)
	case _sticker:
		return fmt.Sprintf("bilibili://uper/center_plus?is_inspiration=1&tags=%s&relation_from=creative-inspiration-story_%d&post_config={\"first_entrance\":\"创作灵感story投稿点击\"}&topic_id=%d&tab_index=1&sticker_id_v2=%d", i.extra.Tags, i.extra.InspirationId, item.TopicID, i.extra.MaterialId)
	case _bgm:
		return fmt.Sprintf("bilibili://uper/center_plus?is_inspiration=1&tags=%s&relation_from=creative-inspiration-story_%d&post_config={\"first_entrance\":\"创作灵感story投稿点击\"}&topic_id=%d&tab_index=1&bgm_id=%d&bgm_name=%s", i.extra.Tags, i.extra.InspirationId, item.TopicID, i.extra.MaterialId, i.extra.MaterialName)
	case _rhythm1, _rhythm2, _rhythm3:
		if toolNeedUserCenter(item.MobiApp(), item.Build()) {
			return fmt.Sprintf("bilibili://uper/user_center/add_archive/?is_inspiration=1&tab_index=3&tags=%s&from=2&topic_id=%d&rhythm_id_v2=%d&is_detail=1&relation_from=creative-inspiration-story_%d&post_config={\"first_entrance\":\"创作灵感story投稿点击\"}", i.extra.Tags, item.TopicID, i.extra.MaterialId, i.extra.InspirationId)
		}
		return fmt.Sprintf("bilibili://uper/center_plus?is_inspiration=1&tags=%s&tab_index=3&relation_from=creative-inspiration-story_%d&topic_id=%d&post_config={\"first_entrance\":\"创作灵感story投稿点击\"}&rhythm_id_v2=%d", i.extra.Tags, i.extra.InspirationId, item.TopicID, i.extra.MaterialId)
	default:
		return ""
	}
}
func (i *inspiration) Type() int64 {
	return _inspiration
}
func (i *inspiration) Title(item *ai.SubItems) string {
	return item.TopicTitle
}
func (i *inspiration) Display() bool {
	return true
}

type music struct {
	extra EntranceExtra
}

func (m *music) Icon() string {
	return "https://i0.hdslb.com/bfs/archive/c53690059458acc2e5caf7e85ad4aa97aec72c36.png"
}
func (m *music) JumpURI(_ *ai.SubItems) string {
	return fmt.Sprintf("https://member.bilibili.com/york/up-inspiration/detail?id=%d&navhide=1&from=10", m.extra.InspirationId)
}
func (m *music) Type() int64 {
	return _music
}
func (m *music) Title(item *ai.SubItems) string {
	return item.TopicTitle
}
func (m *music) Display() bool {
	return true
}

type search struct {
	extra EntranceExtra
}

func (s *search) Icon() string {
	return "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/tYJst4HpgR.png"
}
func (s *search) JumpURI(item *ai.SubItems) string {
	return fmt.Sprintf("bilibili://search?from=app_story_diversion&direct_return=true&keyword=%s&from_avid=%d&from_trackid=%s",
		s.extra.Keyword, item.ID, item.TrackID)
}
func (s *search) Type() int64 {
	return _search
}
func (s *search) Title(item *ai.SubItems) string {
	return item.TopicTitle
}
func (s *search) Display() bool {
	return true
}

func toolNeedUserCenter(mobiApp string, build int) bool {
	return (mobiApp == "android" && build >= 6880000) || (mobiApp == "iphone" && build >= 68800000)
}

func backup44ToInspiration(mobiApp string, build int, materialType int64) bool {
	return materialType == _rhythm2 &&
		((mobiApp == "android" && build < 6800000) || (mobiApp == "iphone" && build < 68300000))
}

func backup46ToInspiration(mobiApp string, build int, materialType int64) bool {
	return materialType == _rhythm3 &&
		((mobiApp == "android" && build < 6820000) || (mobiApp == "iphone" && build < 68300000))
}
